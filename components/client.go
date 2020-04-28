package components

import (
	"fmt"
	"log"
	"net"
	"streming_server/protocol/rtsp/message"
	"streming_server/protocol/rtsp/state"
	"streming_server/ui"
	"streming_server/util"
	"streming_server/video"
	"strings"
	"time"
)

type RtspClient struct {
	server           *RtspServer
	rtcpSender       *RtcpSender
	rtpReceiver      *RtpReceiver
	imageRefresh     *ImageRefresh
	frameSync        *video.FrameSync
	broadcast        *Broadcast
	view             *ui.View
	serverConnection net.Conn
	state            state.State
	videoFileName    string
	sessionId        string
	sequentialNumber int
	isServerside     bool
}

func NewClient(serverAddress string, serverPort string, videoFileName string) *RtspClient {
	log.Println("[RTSP] client started")

	rtspClient := &RtspClient{
		videoFileName:    videoFileName,
		state:            state.Init,
		sequentialNumber: 1,
	}

	frameSync := video.NewFrameSync()
	view := ui.NewView(frameSync,
		rtspClient.onSetup, rtspClient.onRecord, rtspClient.onPlay,
		rtspClient.onPause, rtspClient.onDescribe, rtspClient.onTeardown,
	)

	rtpReceiver := NewRtpReceiver(frameSync, view)
	serverConnection, err := net.Dial("tcp", fmt.Sprintf("%v:%v", serverAddress, serverPort))
	if err != nil {
		log.Fatalln("[RTSP] cannot connect to the server:", err)
	}

	rtspClient.rtcpSender = NewRtcpSender(rtpReceiver)
	rtspClient.imageRefresh = NewImageRefresh(view, frameSync)
	rtspClient.frameSync = frameSync
	rtspClient.rtpReceiver = rtpReceiver
	rtspClient.view = view
	rtspClient.serverConnection = serverConnection
	rtspClient.isServerside = false
	rtspClient.view.StartGUI()

	log.Println("[RTSP] connected with server.")
	return rtspClient
}

// used to receive video from streaming client
func NewServersideClient(server *RtspServer, serverAddress string, serverPort string, videoFileName string) *RtspClient {
	log.Println("[RTSP] serverside client started")

	rtspClient := &RtspClient{
		videoFileName:    videoFileName,
		state:            state.Init,
		sequentialNumber: 1,
	}

	frameSync := video.NewFrameSync()
	rtpReceiver := NewRtpReceiverWithServer(server, frameSync)
	serverConnection, err := net.Dial("tcp", fmt.Sprintf("%v:%v", serverAddress, serverPort))
	if err != nil {
		log.Fatalln("[RTSP] cannot connect to the server:", err)
	}

	rtspClient.rtcpSender = NewRtcpSender(rtpReceiver)
	rtspClient.frameSync = frameSync
	rtspClient.rtpReceiver = rtpReceiver
	rtspClient.serverConnection = serverConnection
	rtspClient.isServerside = true

	log.Println("[RTSP] connected with server.")
	return rtspClient
}

func (rc *RtspClient) onSetup() {
	log.Println("[GUI] setup button has been pressed.")
	if rc.state == state.Init {
		rc.sequentialNumber = 1

		rc.sendRequest(message.Setup)
		replyCode := rc.parseResponse()

		if replyCode == "200" {
			rc.state = state.Ready
			log.Println("[RTSP] State change to READY")
		}
	}
}

func (rc *RtspClient) onRecord() {
	log.Println("[GUI] record button has been pressed.")
	if rc.state == state.Ready {
		rc.sendRequest(message.Record)
		if rc.server == nil {
			rc.server = NewClientsideServer()
			log.Printf("[RTSP] received new connection from %v",
				rc.server.clientConnection.RemoteAddr().String())
			// setup
			rc.server.ParseRequest()
			rc.server.SendResponse()

			rc.broadcast = NewBroadcast(rc.server, rc.frameSync, rc.view)
		}
		replyCode := rc.parseResponse()
		if replyCode == "200" {
			// play
			rc.server.ParseRequest()
			rc.server.SendResponse()
			rc.broadcast.Start()
			rc.state = state.Recording
			log.Println("[RTSP] State change to RECORDING")
		}
	}
}

func (rc *RtspClient) onPlay() {
	log.Println("[GUI] play button has been pressed.")

	if rc.state == state.Ready {
		rc.sequentialNumber++
		rc.rtpReceiver.SetStartTime(time.Now().UnixNano() / int64(time.Millisecond))

		rc.sendRequest(message.Play)
		replyCode := rc.parseResponse()

		if replyCode == "200" {
			rc.state = state.Playing
			rc.rtpReceiver.Start()
			rc.rtcpSender.Start()
			if !rc.isServerside {
				rc.imageRefresh.Start()
			}
			log.Println("[RTSP] State change to Playing")
		}
	}
}

func (rc *RtspClient) onPause() {
	log.Println("[GUI] pause button has been pressed.")

	if rc.state == state.Playing {
		rc.sequentialNumber++

		rc.sendRequest(message.Pause)
		replyCode := rc.parseResponse()

		if replyCode == "200" {
			rc.state = state.Ready
			rc.rtpReceiver.Stop()
			rc.rtcpSender.Stop()
			if rc.imageRefresh != nil {
				rc.imageRefresh.Stop()
			}

			log.Println("[RTSP] State change to READY")
		}
	} else if rc.state == state.Recording {
		rc.sendRequest(message.Pause)
		// pause
		rc.server.ParseRequest()
		rc.server.SendResponse()

		replyCode := rc.parseResponse()
		if replyCode == "200" {
			rc.state = state.Ready
			rc.broadcast.Stop()
			log.Println("[RTSP] State change to READY")
		}
	}
}

func (rc *RtspClient) onDescribe() {
	log.Println("[GUI] describe button has been pressed.")

	rc.sequentialNumber++
	rc.sendRequest(message.Describe)
	replyCode := rc.parseResponse()

	if replyCode == "200" {
		log.Println("[RTSP] received response for DESCRIBE")
	}
}

func (rc *RtspClient) onTeardown() {
	log.Println("[GUI] teardown button has been pressed.")

	rc.sequentialNumber++
	rc.sendRequest(message.Teardown)
	replyCode := rc.parseResponse()

	if replyCode == "200" {
		rc.state = state.Init

		rc.rtpReceiver.Stop()
		rc.rtcpSender.Stop()
		rc.imageRefresh.Stop()

		log.Println("[RTSP] new client State: INIT")
	}
}

func (rc *RtspClient) sendRequest(requestType message.Message) {
	request := fmt.Sprintf("%v %v RTSP/1.0\r\nCSeq: %v\r\n",
		requestType, rc.videoFileName, rc.sequentialNumber)

	if requestType == message.Setup {
		request += fmt.Sprintf("Transport: RTP/UDP;client_port=%v\r\n", rc.rtpReceiver.listeningPort)
	} else if requestType == message.Describe {
		request += fmt.Sprintf("Accept: application/sdp\r\n")
	} else {
		request += fmt.Sprintf("Session: %v\r\n", rc.sessionId)
	}

	_, err := rc.serverConnection.Write([]byte(request))
	if err != nil {
		log.Fatalln("[RTSP] error while sending request to the server:", err)
	}
}

func (rc *RtspClient) parseResponse() string {
	log.Println("[RTSP] received response from server")

	responseBytes := make([]byte, 10000)
	_, err := rc.serverConnection.Read(responseBytes)
	if err != nil {
		log.Fatalln("[RTSP] error while reading response from server:", err)
	}
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
		if rc.state == state.Init {
			if thirdLineParam == "Session:" {
				sessionId := requestElements[6]
				rc.sessionId = sessionId
			} else if thirdLineParam == "Transport:" {
				if !rc.isServerside {
					rc.imageRefresh.SetInterval(time.Duration(33) * time.Millisecond)
				}
				port, err := util.ParseParameter(requestElements[6], "server_port")
				if err == nil {
					rc.rtcpSender.InitConnection(port)
				}
			}
		}
	} else {
		log.Printf("[RTSP] server returned response with error code %v", replyCode)
	}
	return replyCode
}
