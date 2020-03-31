package components

import (
	"fmt"
	"log"
	"net"
	"streming_server/protocol/rtcp"
	"time"
)

const DefaultRtcpInterval = 400
const DefaultRtcpPort = 19001

type RtcpSender struct {
	RtpReceiver        *RtpReceiver
	Ticker             *time.Ticker
	ServerConnection   *net.UDPConn
	Interval           time.Duration
	doneCheck          chan bool
	lastHighSeqNum     int
	lastCumulativeLost int
	started            bool
}

func NewRtcpSender(rtpReceiver *RtpReceiver, serverAddress string) *RtcpSender {
	address, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%v", serverAddress,
		DefaultRtcpPort))
	senderConnection, _ := net.DialUDP("udp", nil, address)
	interval := time.Millisecond * time.Duration(DefaultRtcpInterval)

	result := RtcpSender{
		RtpReceiver:      rtpReceiver,
		Interval:         interval,
		doneCheck:        make(chan bool),
		started:          false,
		ServerConnection: senderConnection,
	}

	return &result
}

func (sender *RtcpSender) sendFeedback() {
	expectedPacketsNumber := sender.RtpReceiver.HighestRecvSeqNum - sender.lastHighSeqNum
	lostPacketsNumber := sender.RtpReceiver.CumulativeLost - sender.lastCumulativeLost
	sender.lastHighSeqNum = sender.RtpReceiver.HighestRecvSeqNum
	sender.lastCumulativeLost = sender.RtpReceiver.CumulativeLost

	lastFractionLost := 0.0
	if expectedPacketsNumber != 0 {
		lastFractionLost = float64(lostPacketsNumber) / float64(expectedPacketsNumber)
	}

	rtpPacket := rtcp.NewPacket(lastFractionLost, sender.lastCumulativeLost, sender.lastHighSeqNum)
	_, _ = sender.ServerConnection.Write(rtpPacket.TransformToBytes())
	log.Println("[RTCP] Feedback packet has been sent to the server.")
}

func (sender *RtcpSender) Start() {
	sender.started = true
	sender.Ticker = time.NewTicker(sender.Interval)

	go func() {
		for {
			select {
			case <-sender.doneCheck:
				return
			case <-sender.Ticker.C:
				sender.sendFeedback()
			}
		}
	}()
}

func (sender *RtcpSender) Stop() {
	if sender.started {
		sender.doneCheck <- true
		sender.Ticker.Stop()
	}
}
