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
	state                state.State
	videoFileName        string
	sessionId            string
	sequentialNumber     int
	componentsStarted    bool
	mainChannel          chan *rtp.Packet
	privateChannel       chan *rtp.Packet
	IsStreaming          bool
}

func NewServer(clientConnection net.Conn, mainChannel chan *rtp.Packet, privateChannel chan *rtp.Packet) *RtspServer {
	log.Println("[RTSP] server started")
	return &RtspServer{
		clientConnection:  clientConnection,
		sessionId:         uuid.New().String(),
		state:             state.Init,
		componentsStarted: false,
		mainChannel:       mainChannel,
		privateChannel:    privateChannel,
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
		state:             state.Init,
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
		//srv.congestionController.Stop()
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
		rtpDestinationPort, err := strconv.Atoi(requestElements[8])
		if err != nil {
			log.Fatalln("[RTSP] error while parsing request:", err)
		}
		srv.videoFileName = fileName
		srv.OnSetup(rtpDestinationPort)
	} else if requestType == message.Record && (srv.state == state.Ready || srv.state == state.Recording) {
		srv.onRecord()
	} else if requestType == message.Play && srv.state == state.Ready {
		srv.onPlay()
	} else if requestType == message.Pause && srv.state == state.Playing {
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
	//srv.congestionController.Start()

	srv.componentsStarted = true
	srv.state = state.Ready

	_, err := srv.clientConnection.Write([]byte(util.PrepareSetupResponse(
		srv.sequentialNumber, srv.rtpSender.frameSync.FramePeriod, rtcpReceiver.ServerAddress),
	))
	if err != nil {
		log.Fatalln("[RTSP] error while sending message:", err)
	}

	log.Println("[RTSP] state changed: READY")
}

func (srv *RtspServer) onRecord() {
	srv.recvClient = NewServersideClient(srv, "127.0.0.1", "26000", "livestream")
	srv.IsStreaming = true

	srv.recvClient.onSetup()
	srv.recvClient.onPlay()
	srv.SendResponse()
	srv.rtpSender.Start()
	if srv.state == state.Ready {
		srv.state = state.Recording
	} else {
		srv.state = state.Ready
	}
	log.Println("[RTSP] state changed: PLAYING")
}

func (srv *RtspServer) onPlay() {
	srv.SendResponse()
	if !srv.IsStreaming {
		srv.frameLoader.Start()
	}
	srv.rtpSender.Start()
	srv.state = state.Playing
	log.Println("[RTSP] state changed: PLAYING")
}

func (srv *RtspServer) OnPause() {
	srv.SendResponse()
	srv.rtpSender.Stop()
	srv.state = state.Ready
	log.Println("[RTSP] state changed: READY")
}

func (srv *RtspServer) OnTeardown() {
	srv.SendResponse()
	srv.rtpSender.Stop()
	log.Println("[RTSP] state changed: INIT")
}

func (srv *RtspServer) OnDescribe() {
	_, err := srv.clientConnection.Write([]byte(util.PrepareDescribeResponse(
		srv.sequentialNumber, os.Args[1], MjpegType, srv.sessionId, srv.videoFileName),
	))
	if err != nil {
		log.Fatalln("[RTSP] error while sending message:", err)
	}
}