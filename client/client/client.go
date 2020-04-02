package client

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"streming_server/client/components"
	"streming_server/client/ui"
	"streming_server/client/video"
	"streming_server/protocol/rtsp/message"
	"streming_server/protocol/rtsp/state"
	"strings"
	"time"
)

type RtspClient struct {
	RtcpSender       *components.RtcpSender
	RtpReceiver      *components.RtpReceiver
	ImageRefresh     *components.ImageRefresh
	FrameSync        *video.FrameSync
	View             *ui.View
	ServerConnection net.Conn
	State            state.State
	VideoFileName    string
	SessionId        string
	SequentialNumber int
}

func NewRtspClient(serverAddress string, serverPort string, videoFileName string) *RtspClient {
	log.Println("[RTSP] client started")

	rtpClient := &RtspClient{
		VideoFileName: videoFileName,
		State:         state.Init,
	}

	frameSync := video.NewFrameSync()
	view := ui.NewView(frameSync,
		rtpClient.onSetup, rtpClient.onPlay, rtpClient.onPause, rtpClient.onDescribe, rtpClient.onTeardown,
	)

	rtpReceiver := components.NewRtpReceiver(frameSync, view, serverAddress)
	rtcpSender := components.NewRtcpSender(rtpReceiver, serverAddress)
	imageRefresh := components.NewImageRefresh(view, frameSync)
	serverConnection, _ := net.Dial("tcp", fmt.Sprintf("%v:%v", serverAddress, serverPort))

	rtpClient.FrameSync = frameSync
	rtpClient.RtcpSender = rtcpSender
	rtpClient.RtpReceiver = rtpReceiver
	rtpClient.ImageRefresh = imageRefresh
	rtpClient.View = view
	rtpClient.ServerConnection = serverConnection
	log.Println("[RTSP] connected with server.")
	rtpClient.View.StartGUI()
	return rtpClient
}

func (rtspClient *RtspClient) onSetup() {
	log.Println("[GUI] Setup button has been pressed.")

	if rtspClient.State == state.Init {
		rtspClient.SequentialNumber = 1

		rtspClient.sendRequest(message.Setup)
		replyCode := rtspClient.parseResponse()

		if replyCode == "200" {
			rtspClient.State = state.Ready
			log.Println("[RTSP] state change to READY")
		}
	}
}

func (rtspClient *RtspClient) onPlay() {
	log.Println("[GUI] Pause button has been pressed.")

	if rtspClient.State == state.Ready {
		rtspClient.SequentialNumber++
		rtspClient.RtpReceiver.StartTime = time.Now().UnixNano() / int64(time.Millisecond)

		rtspClient.sendRequest(message.Play)
		replyCode := rtspClient.parseResponse()

		if replyCode == "200" {
			rtspClient.State = state.Playing
			rtspClient.RtpReceiver.Start()
			rtspClient.RtcpSender.Start()
			rtspClient.ImageRefresh.Start()

			log.Println("[RTSP] state change to Playing")
		}
	}
}

func (rtspClient *RtspClient) onPause() {
	log.Println("[GUI] Pause button has been pressed.")

	if rtspClient.State == state.Playing {
		rtspClient.SequentialNumber++

		rtspClient.sendRequest(message.Pause)
		replyCode := rtspClient.parseResponse()

		if replyCode == "200" {
			rtspClient.State = state.Ready
			rtspClient.RtpReceiver.Stop()
			rtspClient.RtcpSender.Stop()
			rtspClient.ImageRefresh.Stop()

			log.Println("[RTSP] state change to READY")
		}
	}
}

func (rtspClient *RtspClient) onDescribe() {
	log.Println("[GUI] Describe button has been pressed.")

	rtspClient.SequentialNumber++
	rtspClient.sendRequest(message.Describe)
	replyCode := rtspClient.parseResponse()

	if replyCode == "200" {
		log.Println("[RTSP] Received response for DESCRIBE")
	}
}

func (rtspClient *RtspClient) onTeardown() {
	log.Println("[GUI] Teardown button has been pressed.")

	rtspClient.SequentialNumber++
	rtspClient.sendRequest(message.Teardown)
	replyCode := rtspClient.parseResponse()

	if replyCode == "200" {
		rtspClient.State = state.Init

		rtspClient.RtpReceiver.Stop()
		rtspClient.RtcpSender.Stop()
		rtspClient.ImageRefresh.Stop()

		log.Println("[RTSP] new RTSP state INIT")
	}
}

func (rtspClient *RtspClient) sendRequest(requestType message.Message) {
	request := fmt.Sprintf("%v %v RTSP/1.0\r\nCSeq: %v\r\n",
		requestType, rtspClient.VideoFileName, rtspClient.SequentialNumber)

	if requestType == message.Setup {
		request += fmt.Sprintf("Transport: RTP/UDP; client_port= %v\r\n", components.DefaultRtpPort)
	} else if requestType == message.Describe {
		request += fmt.Sprintf("Accept: application/sdp\r\n")
	} else {
		request += fmt.Sprintf("Session: %v\r\n", rtspClient.SessionId)
	}

	_, _ = rtspClient.ServerConnection.Write([]byte(request))
}

func (rtspClient *RtspClient) parseResponse() string {
	log.Println("[RTSP] received response from server")

	responseBytes := make([]byte, 10_000)
	_, _ = rtspClient.ServerConnection.Read(responseBytes)
	responseLines := strings.Split(string(responseBytes), "\r\n")

	// cut whitespace at the end after split
	responseLines = responseLines[0 : len(responseLines)-1]

	requestElements := make([]string, 0)
	for _, line := range responseLines {
		log.Println("\t[RTSP message]", line)
		requestElements = append(requestElements, strings.Split(line, " ")...)
	}

	replyCode := requestElements[1]
	if replyCode == "200" {
		thirdLineParam := requestElements[5]
		if rtspClient.State == state.Init {
			if thirdLineParam == "Session:" {
				sessionId := requestElements[6]
				rtspClient.SessionId = sessionId
			} else if thirdLineParam == "Frame-Period:" {
				framePeriod, _ := strconv.Atoi(requestElements[6])
				rtspClient.ImageRefresh.Interval = time.Duration(framePeriod) * time.Millisecond
			}
		}
	} else {
		log.Printf("Server returned response with error code %v", replyCode)
	}

	return replyCode
}
