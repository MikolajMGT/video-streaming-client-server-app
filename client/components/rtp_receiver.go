package components

import (
	"fmt"
	"net"
	"streming_server/protocol/rtp"
	"time"
)

const DefaultRtpInterval = 20

type RtpReceiver struct {
	Ticker            *time.Ticker
	Interval          time.Duration
	UdpCon            *net.PacketConn
	HighestRecvSeqNum int
	CumulativeLost    int
	ExpectedSeqNum    int
	TotalBytes        int
	buffer            []byte
	doneCheck         chan bool
	startTime         int64
}

func NewRtcpReceiver(serverAddress string, serverPort int) *RtpReceiver {
	address := fmt.Sprintf("%v:%v", serverAddress, serverPort)
	udpConn, _ := net.ListenPacket("udp", address)

	return &RtpReceiver{
		Interval:       DefaultRtpInterval * time.Millisecond,
		UdpCon:         &udpConn,
		buffer:         make([]byte, 300_000),
		doneCheck:      make(chan bool),
		ExpectedSeqNum: 0,
		TotalBytes:     0,
		// TODO HighestRecvSeqNum, CumulativeLost initial?
	}
}

func (receiver *RtpReceiver) receive() {
	packetLength, _, _ := (*receiver.UdpCon).ReadFrom(receiver.buffer)

	// current unix time in milliseconds
	currentTime := time.Now().UnixNano() / int64(time.Millisecond)
	totalPlayTime := currentTime - receiver.startTime
	receiver.startTime = currentTime

	rtpPacket := rtp.NewPacketFromBytes(receiver.buffer, packetLength)
	rtpPacket.Header.Log()

	receiver.ExpectedSeqNum++
	if rtpPacket.Header.SequenceNumber > receiver.HighestRecvSeqNum {
		receiver.HighestRecvSeqNum = rtpPacket.Header.SequenceNumber
	}
	if receiver.ExpectedSeqNum != rtpPacket.Header.SequenceNumber {
		receiver.CumulativeLost++
	}

	dataRate := 0
	if totalPlayTime != 0 {
		dataRate = (receiver.TotalBytes) / int(totalPlayTime/1000.0)
	}
	fractionLost := receiver.CumulativeLost / receiver.HighestRecvSeqNum
	receiver.TotalBytes += len(rtpPacket.Payload)

	// TODO show it in GUI
	fmt.Println(dataRate, fractionLost, receiver.TotalBytes)
}

func (receiver *RtpReceiver) Start() {
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
	receiver.doneCheck <- true
	receiver.Ticker.Stop()
}
