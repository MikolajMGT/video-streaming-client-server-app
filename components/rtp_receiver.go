package components

import (
	"fmt"
	"log"
	"net"
	"streming_server/protocol/large_udp"
	"streming_server/protocol/rtp"
	"streming_server/ui"
	"streming_server/video"
	"strings"
	"time"
)

const DefaultRtpInterval = 1

type RtpReceiver struct {
	Server            *RtspServer
	FrameSync         *video.FrameSync
	View              *ui.View
	Ticker            *time.Ticker
	Interval          time.Duration
	UdpCon            *large_udp.LargeUdpPack
	HighestRecvSeqNum int
	CumulativeLost    int
	ExpectedSeqNum    int
	TotalBytes        int
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
	largeUdpCon := large_udp.NewLargeUdpPackWithSize(udpConn, 64000)

	return &RtpReceiver{
		FrameSync:      frameSync,
		View:           view,
		Interval:       DefaultRtpInterval * time.Microsecond,
		UdpCon:         largeUdpCon,
		doneCheck:      make(chan bool),
		ExpectedSeqNum: 0,
		TotalBytes:     0,
		started:        false,
		ListeningPort:  listeningPort,
		// TODO Review statistics calculation
	}
}

func NewRtpReceiverWithServer(server *RtspServer, frameSync *video.FrameSync, view *ui.View) *RtpReceiver {

	udpConn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		fmt.Println(err)
	}
	listeningPort := strings.Split(udpConn.LocalAddr().String(), ":")[1]
	largeUdpCon := large_udp.NewLargeUdpPackWithSize(udpConn, 64000)

	return &RtpReceiver{
		Server:         server,
		FrameSync:      frameSync,
		View:           view,
		Interval:       DefaultRtpInterval * time.Microsecond,
		UdpCon:         largeUdpCon,
		doneCheck:      make(chan bool),
		ExpectedSeqNum: 0,
		TotalBytes:     0,
		started:        false,
		ListeningPort:  listeningPort,
		// TODO HighestRecvSeqNum, CumulativeLost initial?
	}
}

func (receiver *RtpReceiver) receive() {
	log.Println("RECEIVE")
	buf := make([]byte, 300000)
	packetLength, full, _ := receiver.UdpCon.ReadFrom(buf)
	buf = buf[:packetLength]

	if !full {
		return
	}

	//current unix time in milliseconds
	currentTime := time.Now().UnixNano() / int64(time.Millisecond)
	totalPlayTime := currentTime - receiver.StartTime
	receiver.StartTime = currentTime

	rtpPacket := rtp.NewPacketFromBytes(buf, packetLength)
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
}

func (receiver *RtpReceiver) receiveAndForward() {
	log.Println("RECEIVE AND FORWARD")

	buf := make([]byte, 300000)
	packetLength, full, _ := receiver.UdpCon.ReadFrom(buf)
	buf = buf[:packetLength]

	if !full {
		return
	}

	//current unix time in milliseconds
	currentTime := time.Now().UnixNano() / int64(time.Millisecond)
	receiver.StartTime = currentTime

	rtpPacket := rtp.NewPacketFromBytes(buf, packetLength)
	rtpPacket.Header.Log()

	receiver.ExpectedSeqNum++
	if rtpPacket.Header.SequenceNumber > receiver.HighestRecvSeqNum {
		receiver.HighestRecvSeqNum = rtpPacket.Header.SequenceNumber
	}
	if receiver.ExpectedSeqNum != rtpPacket.Header.SequenceNumber {
		receiver.CumulativeLost++
	}

	receiver.TotalBytes += len(rtpPacket.Payload)

	receiver.Server.MainChannel <- rtpPacket
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
				if receiver.Server != nil {
					receiver.receiveAndForward()
				} else {
					receiver.receive()
				}
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
