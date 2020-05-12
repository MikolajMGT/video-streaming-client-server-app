package components

import (
	"log"
	"net"
	"streming_server/protocol/rtcp"
	"time"
)

const DefaultRtcpInterval = 5

type RtcpSender struct {
	rtpReceiver        *RtpReceiver
	ticker             *time.Ticker
	serverConnection   *net.UDPConn
	interval           time.Duration
	doneCheck          chan bool
	lastHighSeqNum     int
	lastCumulativeLost int
	started            bool
}

func NewRtcpSender(rtpReceiver *RtpReceiver) *RtcpSender {
	interval := time.Second * time.Duration(DefaultRtcpInterval)

	result := RtcpSender{
		rtpReceiver: rtpReceiver,
		interval:    interval,
		doneCheck:   make(chan bool),
		started:     false,
	}

	return &result
}

func (s *RtcpSender) InitConnection(serverAddress string) {
	address, err := net.ResolveUDPAddr("udp", serverAddress)
	if err != nil {
		log.Fatalln("[RTCP] error while resolving rtcp address:", err)
	}
	senderConnection, err := net.DialUDP("udp", nil, address)
	if err != nil {
		log.Fatalln("[RTCP] error while connecting with rtcp receiver:", err)
	}
	s.serverConnection = senderConnection
}

func (s *RtcpSender) sendFeedback() {
	expectedPacketsNumber := s.rtpReceiver.highestRecvSeqNum - s.lastHighSeqNum
	lostPacketsNumber := s.rtpReceiver.cumulativeLost - s.lastCumulativeLost
	s.lastHighSeqNum = s.rtpReceiver.highestRecvSeqNum
	s.lastCumulativeLost = s.rtpReceiver.cumulativeLost

	lastFractionLost := 0.0
	if expectedPacketsNumber != 0 {
		lastFractionLost = float64(lostPacketsNumber) / float64(expectedPacketsNumber)
	}

	rtpPacket := rtcp.NewPacket(lastFractionLost, s.lastCumulativeLost, s.lastHighSeqNum)
	_, err := s.serverConnection.Write(rtpPacket.TransformToBytes())
	if err != nil {
		log.Println("[RTCP] error while sending packet:", err)
		return
	}
	log.Println("[RTCP] feedback packet has been sent to the server.")
}

func (s *RtcpSender) Start() {
	s.started = true
	s.ticker = time.NewTicker(s.interval)

	go func() {
		for {
			select {
			case <-s.doneCheck:
				return
			case <-s.ticker.C:
				s.sendFeedback()
			}
		}
	}()
}

func (s *RtcpSender) Stop() {
	if s.started {
		s.doneCheck <- true
		s.ticker.Stop()
		s.started = false
	}
}

func (s *RtcpSender) Close() {
	s.Stop()
	err := s.serverConnection.Close()
	if err != nil {
		log.Println("[RTCP] error while closing connection:", err)
	}
}
