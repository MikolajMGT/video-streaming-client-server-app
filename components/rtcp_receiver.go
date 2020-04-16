package components

import (
	"fmt"
	"net"
	"streming_server/protocol/rtcp"
	"streming_server/util"
	"strings"
	"time"
)

type RtcpReceiver struct {
	Ticker          *time.Ticker
	Interval        time.Duration
	UdpCon          *net.PacketConn
	CongestionLevel int
	buffer          []byte
	doneCheck       chan bool
	ServerAddress   string
}

func NewRtcpReceiver(clientAddress net.Addr) *RtcpReceiver {
	addressAndPort := strings.Split(clientAddress.String(), ":")
	address := fmt.Sprintf("%v:%v", addressAndPort[0], 0)
	udpConn, _ := net.ListenPacket("udp", address)
	serverAddress := udpConn.LocalAddr().String()

	return &RtcpReceiver{
		Interval:        DefaultRtcpInterval * time.Millisecond,
		UdpCon:          &udpConn,
		buffer:          make([]byte, 24),
		doneCheck:       make(chan bool),
		CongestionLevel: util.NoCongestion,
		ServerAddress:   serverAddress,
	}
}

func (receiver *RtcpReceiver) receive() {
	_, _, _ = (*receiver.UdpCon).ReadFrom(receiver.buffer)
	rtcpPacket := rtcp.NewPacketFromBytes(receiver.buffer)
	rtcpPacket.Log()

	receiver.CongestionLevel = util.ResolveCongestionLevel(rtcpPacket.FractionLost)
}

func (receiver *RtcpReceiver) Start() {
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

func (receiver *RtcpReceiver) Stop() {
	receiver.doneCheck <- true
	receiver.Ticker.Stop()
}
