package components

import (
	"fmt"
	"net"
	"streming_server/client/ui"
	"streming_server/client/video"
	"streming_server/protocol/rtp"
	"time"
)

const DefaultRtpInterval = 20
const DefaultRtpPort = 25000

type RtpReceiver struct {
	FrameSync         *video.FrameSync
	View              *ui.View
	Ticker            *time.Ticker
	Interval          time.Duration
	UdpCon            *net.PacketConn
	HighestRecvSeqNum int
	CumulativeLost    int
	ExpectedSeqNum    int
	TotalBytes        int
	buffer            []byte
	doneCheck         chan bool
	StartTime         int64
	started           bool
}

func NewRtpReceiver(frameSync *video.FrameSync, view *ui.View, serverAddress string,
) *RtpReceiver {

	address := fmt.Sprintf("%v:%v", serverAddress, DefaultRtpPort)
	udpConn, _ := net.ListenPacket("udp", address)

	return &RtpReceiver{
		FrameSync:      frameSync,
		View:           view,
		Interval:       DefaultRtpInterval * time.Millisecond,
		UdpCon:         &udpConn,
		buffer:         make([]byte, 300_000),
		doneCheck:      make(chan bool),
		ExpectedSeqNum: 0,
		TotalBytes:     0,
		started:        false,
		// TODO HighestRecvSeqNum, CumulativeLost initial?
	}
}

func (receiver *RtpReceiver) receive() {
	packetLength, _, _ := (*receiver.UdpCon).ReadFrom(receiver.buffer)

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

	receiver.FrameSync.AddFrame(rtpPacket.Payload, rtpPacket.Header.SequenceNumber)
	receiver.View.UpdateImage(receiver.FrameSync.NextFrame())
}

func (receiver *RtpReceiver) Start() {
	receiver.started = true
	receiver.Ticker = time.NewTicker(receiver.Interval)

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
		receiver.doneCheck <- true
		receiver.Ticker.Stop()
		receiver.started = false
	}
}
