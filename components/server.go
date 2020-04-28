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
	videoFileName        string
	sessionId            string
	sequentialNumber     int
	componentsStarted    bool
	IsStreaming          bool
}

func NewServer(clientConnection net.Conn, mainChannel chan *rtp.Packet, privateChannel chan *rtp.Packet) *RtspServer {
	log.Println("[RTSP] server started")
	return &RtspServer{
		clientConnection:  clientConnection,
		sessionId:         uuid.New().String(),
		State:             state.Init,
		mainChannel:       mainChannel,
		privateChannel:    privateChannel,
		componentsStarted: false,
		IsStreaming:       false,
	}
}

// used when client is currently streaming video
func NewClientsideServer() *RtspServer {
	log.Println("[RTSP] clientside server started")
	listener, err := net.Listen("tcp", fmt.Sprint(":", 26000))
	if err != nil {
		log.Fatalln("[RTSP] cannot open connection:", err)
	}
	clientConnection, err := listener.Accept()
	if err != nil {
		log.Fatalln("[RTSP] error while connecting with client:", err)
	}
	return &RtspServer{
		clientConnection:  clientConnection,
		sessionId:         uuid.New().String(),
		State:             state.Init,
		componentsStarted: false,
		IsStreaming:       false,
	}
}

func (srv *RtspServer) SendResponse() {
	response := fmt.Sprintf("%vSession: %v\r\n", util.FormatHeader(srv.sequentialNumber), srv.sessionId)
	_, err := srv.clientConnection.Write([]byte(response))
	if err != nil {
		log.Fatalln("[RTSP] cannot send response:", err)
	}

}

func (srv *RtspServer) Start() {
	// waiting for initial SETUP request
	for {
		requestType := srv.ParseRequest()
		if requestType == message.Setup || requestType == message.Exit {
			break
		}
	}

	// handling further requests
	for {
		requestType := srv.ParseRequest()
		if requestType == message.Exit {
			srv.ShutDown()
			break
		}
	}
}

func (srv *RtspServer) ShutDown() {
	if srv.componentsStarted {
		srv.congestionController.Stop()
		srv.rtpSender.Stop()
	}
}

func (srv *RtspServer) ParseRequest() message.Message {
	bufferedReader := bufio.NewReader(srv.clientConnection)
	requestElements := util.ReadRequestElements(bufferedReader)
	if len(requestElements) == 0 {
		return message.Exit
	}

	requestType := requestElements[0]
	seqNumber, err := strconv.Atoi(requestElements[4])
	if err != nil {
		log.Fatalln("[RTSP] error while parsing request:", err)
	}
	srv.sequentialNumber = seqNumber
	if requestType == message.Setup {
		fileName := requestElements[1]
		port, err := util.ParseParameter(requestElements[6], "client_port")
		if err == nil {
			portAsInt, _ := strconv.Atoi(port)
			srv.OnSetup(portAsInt)

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
	rtcpReceiver := NewRtcpReceiver(srv.clientConnection.RemoteAddr())
	srv.congestionController = NewCongestionController(rtcpReceiver, srv.frameSync)
	srv.rtpSender = NewRtpSender(srv.clientConnection.RemoteAddr(), rtpDestinationPort,
		srv.congestionController, rtcpReceiver, srv.frameSync)

	srv.congestionController.SetRtpSender(srv.rtpSender)
	srv.congestionController.Start()

	srv.componentsStarted = true
	srv.State = state.Ready

	_, err := srv.clientConnection.Write([]byte(util.PrepareSetupResponse(
		srv.sequentialNumber, rtcpReceiver.ServerPort),
	))
	if err != nil {
		log.Fatalln("[RTSP] error while sending message:", err)
	}

	log.Println("[RTSP] State changed: READY")
}

func (srv *RtspServer) onRecord() {
	if srv.State == state.Ready {
		if srv.recvClient == nil {
			srv.recvClient = NewServersideClient(srv, "127.0.0.1", "26000", "livestream")
			srv.recvClient.onSetup()
			srv.IsStreaming = true
		}
		srv.recvClient.onPlay()
		srv.SendResponse()
		srv.rtpSender.Start()
		srv.State = state.Recording
		log.Println("[RTSP] State changed: RECORDING")
	}
}

func (srv *RtspServer) onPlay() {
	srv.SendResponse()
	if !srv.IsStreaming {
		srv.frameLoader.Start()
	}
	srv.rtpSender.Start()
	srv.State = state.Playing
	log.Println("[RTSP] State changed: PLAYING")
}

func (srv *RtspServer) OnPause() {
	if srv.State == state.Recording {
		srv.recvClient.onPause()
	}
	srv.SendResponse()
	srv.rtpSender.Stop()
	srv.State = state.Ready
	log.Println("[RTSP] State changed: READY")
}

func (srv *RtspServer) OnTeardown() {
	srv.SendResponse()
	srv.rtpSender.Stop()
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
