package server

import (
	"bufio"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net"
	"os"
	"strconv"
	"streming_server/protocol/rtsp/message"
	"streming_server/protocol/rtsp/state"
	"streming_server/server/components"
	"streming_server/server/util"
	"streming_server/server/video"
)

// TODO error handling

type RtspServer struct {
	RtpSender            *components.RtpSender
	CongestionController *components.CongestionController
	ClientConnection     *net.Conn
	State                state.State
	VideoFileName        string
	SessionId            string
	SequentialNumber     int
	componentsStarted    bool
}

func NewRtspServer(port string) *RtspServer {
	log.Println("[RTSP] Server started")
	listener, _ := net.Listen("tcp", fmt.Sprint(":", port))
	clientConnection, _ := listener.Accept()

	log.Printf("[RTSP] received new connection from %v", clientConnection.RemoteAddr().String())
	return &RtspServer{
		ClientConnection:  &clientConnection,
		SessionId:         uuid.New().String(),
		State:             state.Init,
		componentsStarted: false,
	}
}

func (srv *RtspServer) SendResponse() {
	response := fmt.Sprintf("%vSession: %v\r\n", util.FormatHeader(srv.SequentialNumber), srv.SessionId)
	_, _ = (*srv.ClientConnection).Write([]byte(response))

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
		srv.CongestionController.Stop()
		srv.RtpSender.Stop()
	}
}

func (srv *RtspServer) ParseRequest() message.Message {
	bufferedReader := bufio.NewReader(*srv.ClientConnection)
	requestElements := util.ReadRequestElements(bufferedReader)
	if len(requestElements) == 0 {
		return message.Exit
	}

	requestType := requestElements[0]
	seqNumber, _ := strconv.Atoi(requestElements[4])
	srv.SequentialNumber = seqNumber

	if requestType == message.Setup {

		fileName := requestElements[1]
		rtpDestinationPort, _ := strconv.Atoi(requestElements[8])
		srv.VideoFileName = fileName
		srv.OnSetup(fileName, rtpDestinationPort)

	} else if requestType == message.Play && srv.State == state.Ready {
		srv.onPlay()
	} else if requestType == message.Pause && srv.State == state.Playing {
		srv.OnPause()
	} else if requestType == message.Teardown {
		srv.OnTeardown()
	} else if requestType == message.Describe {
		srv.OnDescribe()
	}

	return message.Message(requestType)
}

func (srv *RtspServer) OnSetup(fileName string, rtpDestinationPort int) {
	videoStream := video.NewStream(fileName)
	rtcpReceiver := components.NewRtcpReceiver((*srv.ClientConnection).RemoteAddr())
	srv.CongestionController = components.NewCongestionController(rtcpReceiver, videoStream)
	srv.RtpSender = components.NewRtpSender((*srv.ClientConnection).RemoteAddr(), rtpDestinationPort,
		srv.CongestionController, rtcpReceiver, videoStream)

	srv.CongestionController.RtpSender = srv.RtpSender
	srv.CongestionController.Start()

	srv.componentsStarted = true
	srv.State = state.Ready

	_, _ = (*srv.ClientConnection).Write([]byte(util.PrepareSetupResponse(
		srv.SequentialNumber, srv.RtpSender.VideoStream.FramePeriod),
	))

	log.Println("[RTSP] state changed: READY")
}

func (srv *RtspServer) onPlay() {
	srv.SendResponse()
	srv.RtpSender.Start()
	srv.State = state.Playing
	log.Println("[RTSP] state changed: PLAYING")
}

func (srv *RtspServer) OnPause() {
	srv.SendResponse()
	srv.RtpSender.Stop()
	srv.State = state.Ready
	log.Println("[RTSP] state changed: READY")
}

func (srv *RtspServer) OnTeardown() {
	srv.SendResponse()
	srv.RtpSender.Stop()
	log.Println("[RTSP] state changed: INIT")
}

func (srv *RtspServer) OnDescribe() {
	_, _ = (*srv.ClientConnection).Write([]byte(util.PrepareDescribeResponse(
		srv.SequentialNumber, os.Args[1], components.MjpegType, srv.SessionId, srv.VideoFileName),
	))
}
