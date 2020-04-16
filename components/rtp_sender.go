package components

import (
	"fmt"
	"log"
	"net"
	"streming_server/protocol/large_udp"
	"streming_server/protocol/rtp"
	"streming_server/video"
	"strings"
	"time"
)

const MjpegType = 26
const DefaultInterval = 10

type RtpSender struct {
	RtcpReceiver         *RtcpReceiver
	CongestionController *CongestionController
	FrameSync            *video.FrameSync
	Ticker               *time.Ticker
	ClientConnection     *large_udp.LargeUdpConn
	Interval             time.Duration
	FrameBuffer          []byte
	doneCheck            chan bool
	started              bool
}

func NewRtpSender(
	clientAddress net.Addr, destinationPort int, congestionController *CongestionController,
	rtcpReceiver *RtcpReceiver, frameSync *video.FrameSync,
) *RtpSender {

	addressAndPort := strings.Split(clientAddress.String(), ":")
	address, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%v", addressAndPort[0],
		destinationPort))
	clientConnection, _ := net.DialUDP("udp", nil, address)
	largeUdpConn := large_udp.NewLargeUdpConnWithSize(clientConnection, 64000)

	//interval := time.Millisecond * time.Duration(videoStream.FramePeriod)

	result := RtpSender{
		RtcpReceiver:         rtcpReceiver,
		CongestionController: congestionController,
		FrameSync:            frameSync,
		Interval:             time.Duration(DefaultInterval) * time.Millisecond,
		FrameBuffer:          make([]byte, 300000),
		doneCheck:            make(chan bool),
		started:              false,
		ClientConnection:     largeUdpConn,
	}

	return &result
}

func (sender *RtpSender) SendFrame() {

	if sender.FrameSync.Empty() {
		return
	}
	data := sender.FrameSync.NextFrame()

	if len(data) == 0 {
		sender.Stop()
		return
	}

	//sender.CongestionController.AdjustCompressionQuality(sender.FrameBuffer, imageLength)
	rtpPacket := rtp.NewPacket(
		rtp.NewHeader(
			MjpegType, sender.FrameSync.CurrentSeqNum, sender.FrameSync.CurrentSeqNum*sender.FrameSync.FramePeriod,
		),
		len(data), data,
	)
	_, _ = sender.ClientConnection.Write(rtpPacket.TransformToBytes())
	log.Printf("Sent frame no. %v with size %v", sender.FrameSync.CurrentSeqNum, len(data))
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
				sender.SendFrame()
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
