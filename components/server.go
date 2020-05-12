package components

import (
	"bufio"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net"
	"os"
	"strconv"
	"streming_server/protocol/rtp"
	"streming_server/protocol/rtsp/message"
	"streming_server/protocol/rtsp/state"
	"streming_server/util"
	"streming_server/video"
	"strings"
)

type RtspServer struct {
	recvClient           *RtspClient
	rtpSender            *RtpSender
	congestionController *CongestionController
	frameLoader          *FrameLoader
	frameSync            *video.FrameSync
	clientConnection     net.Conn
	State                state.State
	mainChannel          chan *rtp.Packet
	privateChannel       chan *rtp.Packet
	clientsideServerPort string
	videoFileName        string
	sessionId            string
	sequentialNumber     int
	isClientSide         bool
}

func NewServer(clientConnection net.Conn, mainChannel chan *rtp.Packet, privateChannel chan *rtp.Packet) *RtspServer {
	log.Println("[RTSP] server started")
	return &RtspServer{
		clientConnection: clientConnection,
		sessionId:        uuid.New().String(),
		State:            state.Init,
		mainChannel:      mainChannel,
		privateChannel:   privateChannel,
		isClientSide:     false,
	}
}

// used when client is currently streaming video
func NewClientsideServer(port int) *RtspServer {
	log.Println("[RTSP] clientside server started")
	listener, err := net.Listen("tcp", fmt.Sprint(":", port))
	if err != nil {
		log.Fatalln("[RTSP] cannot open connection:", err)
	}
	clientConnection, err := listener.Accept()
	if err != nil {
		log.Fatalln("[RTSP] error while connecting with client:", err)
	}
	return &RtspServer{
		clientConnection: clientConnection,
		sessionId:        uuid.New().String(),
		State:            state.Init,
		isClientSide:     false,
	}
}

func (srv *RtspServer) SendResponse() {
	response := util.FormatHeader(srv.sequentialNumber, srv.sessionId)
	_, err := srv.clientConnection.Write([]byte(response))
	if err != nil {
		log.Fatalln("[RTSP] cannot send response:", err)
	}

}

func (srv *RtspServer) Start() {
	// waiting for initial SETUP request
	for {
		requestType := srv.ParseRequest()
		if requestType == message.Setup || srv.State == state.Detached {
			break
		}
	}

	// handling further requests
	for {
		srv.ParseRequest()
		if srv.State == state.Detached {
			break
		}
	}
}

func (srv *RtspServer) ParseRequest() message.Message {
	bufferedReader := bufio.NewReader(srv.clientConnection)
	requestElements := util.ReadRequestElements(bufferedReader)
	if len(requestElements) == 0 {
		// client disconnected
		srv.State = state.Detached
		return ""
	}

	requestType := requestElements[0]
	seqNumber, err := strconv.Atoi(requestElements[4])
	if err != nil {
		log.Fatalln("[RTSP] error while parsing request:", err)
	}
	srv.sequentialNumber = seqNumber
	if requestType == message.Setup {
		fileName := requestElements[1]
		ports, err := util.ParseParameter(requestElements[6], "client_port")
		if err == nil {
			portAsInt, _ := strconv.Atoi(ports[0])
			srv.OnSetup(portAsInt)
			srv.clientsideServerPort = ports[1]

		}
		srv.videoFileName = fileName
	} else if requestType == message.Record && srv.State == state.Ready {
		srv.onRecord()
	} else if requestType == message.Play && srv.State == state.Ready {
		srv.onPlay()
	} else if requestType == message.Pause && (srv.State == state.Playing || srv.State == state.Recording) {
		srv.OnPause()
	} else if requestType == message.Teardown {
		srv.OnTeardown()
	} else if requestType == message.Describe {
		srv.OnDescribe()
	}

	return message.Message(requestType)
}

func (srv *RtspServer) OnSetup(rtpDestinationPort int) {
	srv.frameSync = video.NewFrameSync()
	srv.frameLoader = NewFrameLoader(srv.frameSync, srv.privateChannel)
	rtcpReceiver := NewRtcpReceiver()
	srv.congestionController = NewCongestionController(rtcpReceiver, srv.frameSync)
	srv.rtpSender = NewRtpSender(srv.clientConnection.RemoteAddr(), rtpDestinationPort,
		srv.congestionController, rtcpReceiver, srv.frameSync)

	srv.congestionController.SetRtpSender(srv.rtpSender)
	srv.congestionController.Start()

	srv.State = state.Ready

	_, err := srv.clientConnection.Write([]byte(util.PrepareSetupResponse(
		srv.sequentialNumber, rtcpReceiver.ServerPort, srv.sessionId),
	))
	if err != nil {
		log.Fatalln("[RTSP] error while sending message:", err)
	}

	log.Println("[RTSP] State changed: READY")
}

func (srv *RtspServer) onRecord() {
	if srv.State == state.Ready {
		if srv.recvClient == nil {
			address := strings.Split(srv.clientConnection.RemoteAddr().String(), ":")[0]
			srv.recvClient = NewServersideClient(srv, address, srv.clientsideServerPort, "livestream")
			srv.recvClient.onSetup()
			srv.isClientSide = true
		}
		srv.recvClient.onPlay()
		srv.SendResponse()
		srv.State = state.Recording
		log.Println("[RTSP] State changed: RECORDING")
	}
}

func (srv *RtspServer) onPlay() {
	srv.SendResponse()
	if !srv.isClientSide {
		srv.frameLoader.Start()
	}
	srv.rtpSender.Start()
	srv.State = state.Playing
	log.Println("[RTSP] State changed: PLAYING")
}

func (srv *RtspServer) OnPause() {
	if srv.State == state.Recording {
		srv.recvClient.onPause()
	} else {
		srv.rtpSender.Stop()
	}
	srv.SendResponse()
	srv.State = state.Ready
	log.Println("[RTSP] State changed: READY")
}

func (srv *RtspServer) OnTeardown() {
	srv.congestionController.Stop()
	srv.rtpSender.Close()
	if srv.recvClient != nil {
		srv.recvClient.onTeardown()
	}
	srv.SendResponse()
	srv.State = state.Init
	log.Println("[RTSP] State changed: INIT")
}

func (srv *RtspServer) OnDescribe() {
	_, err := srv.clientConnection.Write([]byte(util.PrepareDescribeResponse(
		srv.sequentialNumber, os.Args[1], MjpegType, srv.sessionId, srv.videoFileName),
	))
	if err != nil {
		log.Fatalln("[RTSP] error while sending message:", err)
	}
}

func (srv *RtspServer) CloseConnection() {
	err := srv.clientConnection.Close()
	if err != nil {
		log.Println("[RTSP] error while closing connection:", err)
	}
}
