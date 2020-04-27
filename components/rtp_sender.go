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
	rtcpReceiver         *RtcpReceiver
	congestionController *CongestionController
	frameSync            *video.FrameSync
	ticker               *time.Ticker
	clientConnection     *large_udp.LargeUdpConn
	interval             time.Duration
	frameBuffer          []byte
	doneCheck            chan bool
	started              bool
}

func NewRtpSender(
	clientAddress net.Addr, destinationPort int, congestionController *CongestionController,
	rtcpReceiver *RtcpReceiver, frameSync *video.FrameSync,
) *RtpSender {

	addressAndPort := strings.Split(clientAddress.String(), ":")
	address, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%v:%v", addressAndPort[0],
		destinationPort))
	if err != nil {
		log.Fatalln("[RTP] cannot resolve address:", err)
	}
	clientConnection, err := net.DialUDP("udp", nil, address)
	if err != nil {
		log.Fatalln("[RTP] cannot resolve rtp connection:", err)
	}
	largeUdpConn := large_udp.NewLargeUdpConnWithSize(clientConnection, 64000)

	result := RtpSender{
		rtcpReceiver:         rtcpReceiver,
		congestionController: congestionController,
		frameSync:            frameSync,
		interval:             time.Duration(DefaultInterval) * time.Millisecond,
		frameBuffer:          make([]byte, 300000),
		doneCheck:            make(chan bool),
		clientConnection:     largeUdpConn,
		started:              false,
	}

	return &result
}

func (s *RtpSender) SendFrame() {

	if s.frameSync.Empty() {
		return
	}
	data := s.frameSync.NextFrame()

	if len(data) == 0 {
		s.Stop()
		return
	}

	//s.congestionController.AdjustCompressionQuality(s.frameBuffer, imageLength)
	rtpPacket := rtp.NewPacket(
		rtp.NewHeader(
			MjpegType, s.frameSync.CurrentSeqNum, s.frameSync.CurrentSeqNum*s.frameSync.FramePeriod,
		),
		len(data), data,
	)
	_, err := s.clientConnection.Write(rtpPacket.TransformToBytes())
	if err != nil {
		log.Println("[RTP] error while sending packet:", err)
		return
	}
	log.Printf("Sent frame no. %v with size %v", s.frameSync.CurrentSeqNum, len(data))
	rtpPacket.Header.Log()
}

func (s *RtpSender) Start() {
	s.rtcpReceiver.Start()
	s.started = true
	s.ticker = time.NewTicker(s.interval)

	go func() {
		for {
			select {
			case <-s.doneCheck:
				return
			case <-s.ticker.C:
				s.SendFrame()
			}
		}
	}()
}

func (s *RtpSender) Stop() {
	if s.started {
		s.rtcpReceiver.Stop()
		s.doneCheck <- true
		s.ticker.Stop()
	}
}

func (s *RtpSender) UpdateInterval(newInterval time.Duration) {
	s.ticker.Stop()
	s.ticker = time.NewTicker(newInterval)
}
