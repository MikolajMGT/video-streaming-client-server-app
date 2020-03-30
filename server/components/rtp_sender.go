package components

import (
	"fmt"
	"log"
	"net"
	"streming_server/protocol/rtp"
	"streming_server/server/video"
	"strings"
	"time"
)

const MjpegType = 26

type RtpSender struct {
	RtcpReceiver         *RtcpReceiver
	CongestionController *CongestionController
	VideoStream          *video.Stream
	Ticker               *time.Ticker
	ClientConnection     *net.UDPConn
	Interval             time.Duration
	FrameBuffer          []byte
	doneCheck            chan bool
	started              bool
}

func NewRtpSender(
	clientAddress net.Addr, destinationPort int, congestionController *CongestionController,
	rtcpReceiver *RtcpReceiver, videoStream *video.Stream,
) *RtpSender {

	addressAndPort := strings.Split(clientAddress.String(), ":")
	address, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%v", addressAndPort[0],
		destinationPort))
	clientConnection, _ := net.DialUDP("udp", nil, address)

	interval := time.Millisecond * time.Duration(videoStream.FramePeriod)

	result := RtpSender{
		RtcpReceiver:         rtcpReceiver,
		CongestionController: congestionController,
		VideoStream:          videoStream,
		Interval:             interval,
		FrameBuffer:          make([]byte, 300_000),
		doneCheck:            make(chan bool),
		started:              false,
		ClientConnection:     clientConnection,
	}

	return &result
}

func (sender *RtpSender) sendFrame() {
	imageLength := sender.VideoStream.NextFrame(sender.FrameBuffer)

	if imageLength == 0 {
		sender.Stop()
		return
	}

	sender.CongestionController.AdjustCompressionQuality(sender.FrameBuffer, imageLength)
	rtpPacket := rtp.NewPacket(
		rtp.NewHeader(
			MjpegType, sender.VideoStream.FrameCounter, sender.VideoStream.FrameCounter*sender.VideoStream.FramePeriod,
		),
		imageLength, sender.FrameBuffer[0:imageLength],
	)
	_, _ = sender.ClientConnection.Write(rtpPacket.TransformToBytes())

	log.Printf("Sent frame no. %v with size %v", sender.VideoStream.FrameCounter, imageLength)
	rtpPacket.Header.Log()
}

func (sender *RtpSender) Start() {
	sender.RtcpReceiver.Start()
	sender.started = true
	sender.Ticker = time.NewTicker(sender.Interval)

	go func() {
		for {
			select {
			case <-sender.doneCheck:
				return
			case <-sender.Ticker.C:
				sender.sendFrame()
			}
		}
	}()
}

func (sender *RtpSender) Stop() {
	if sender.started {
		sender.RtcpReceiver.Stop()
		sender.doneCheck <- true
		sender.Ticker.Stop()
	}
}

func (sender *RtpSender) UpdateInterval(newInterval time.Duration) {
	sender.Ticker.Stop()
	sender.Ticker = time.NewTicker(newInterval)
}
