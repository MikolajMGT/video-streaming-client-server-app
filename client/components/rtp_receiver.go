package components

import (
	"fmt"
	"net"
	"streming_server/client/ui"
	"streming_server/client/video"
	"streming_server/protocol/large_udp"
	"streming_server/protocol/rtp"
	"strings"
	"time"
)

const DefaultRtpInterval = 1
const DefaultRtpPort = 25002

type RtpReceiver struct {
	FrameSync         *video.FrameSync
	View              *ui.View
	Ticker            *time.Ticker
	Interval          time.Duration
	UdpCon            *large_udp.LargeUdpPack
	HighestRecvSeqNum int
	CumulativeLost    int
	ExpectedSeqNum    int
	TotalBytes        int
	buffer            []byte
	doneCheck         chan bool
	StartTime         int64
	started           bool
	ListeningPort     string
}

func NewRtpReceiver(frameSync *video.FrameSync, view *ui.View) *RtpReceiver {

	udpConn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		fmt.Println(err)
	}
	listeningPort := strings.Split(udpConn.LocalAddr().String(), ":")[1]
	largeUdpCon := large_udp.NewLargeUdpPackWithSize(udpConn, 64_000)

	return &RtpReceiver{
		FrameSync:      frameSync,
		View:           view,
		Interval:       DefaultRtpInterval * time.Microsecond,
		UdpCon:         largeUdpCon,
		buffer:         make([]byte, 300_000),
		doneCheck:      make(chan bool),
		ExpectedSeqNum: 0,
		TotalBytes:     0,
		started:        false,
		ListeningPort:  listeningPort,
		// TODO HighestRecvSeqNum, CumulativeLost initial?
	}
}

func (receiver *RtpReceiver) receive() {
	packetLength, full, _ := receiver.UdpCon.ReadFrom(receiver.buffer)

	if !full {
		return
	}

	//current unix time in milliseconds
	currentTime := time.Now().UnixNano() / int64(time.Millisecond)
	totalPlayTime := currentTime - receiver.StartTime
	receiver.StartTime = currentTime

	rtpPacket := rtp.NewPacketFromBytes(receiver.buffer, packetLength)
	rtpPacket.Header.Log()

	receiver.ExpectedSeqNum++
	if rtpPacket.Header.SequenceNumber > receiver.HighestRecvSeqNum {
		receiver.HighestRecvSeqNum = rtpPacket.Header.SequenceNumber
	}
	if receiver.ExpectedSeqNum != rtpPacket.Header.SequenceNumber {
		receiver.CumulativeLost++
	}

	dataRate := 0.0
	if totalPlayTime != 0 {
		dataRate = float64(receiver.TotalBytes) / (float64(totalPlayTime) / 1000.0)
	}
	fractionLost := receiver.CumulativeLost / receiver.HighestRecvSeqNum
	receiver.TotalBytes += len(rtpPacket.Payload)

	receiver.View.UpdateStatistics(receiver.TotalBytes, fractionLost, dataRate)

	data := make([]byte, len(rtpPacket.Payload))
	copy(data, rtpPacket.Payload)
	receiver.FrameSync.AddFrame(data, rtpPacket.Header.SequenceNumber)
	//receiver.View.UpdateImage(receiver.FrameSync.NextFrame())
}

func (receiver *RtpReceiver) Start() {
	receiver.started = true
	receiver.Ticker = time.NewTicker(receiver.Interval)
	receiver.doneCheck = make(chan bool)

	go func() {
		for {
			select {
			case <-receiver.doneCheck:
				return
			case <-receiver.Ticker.C:
				receiver.receive()
			}
		}
	}()
}

func (receiver *RtpReceiver) Stop() {
	if receiver.started {
		close(receiver.doneCheck)
		receiver.Ticker.Stop()
		receiver.started = false
	}
}
