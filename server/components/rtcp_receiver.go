package components

import (
	"fmt"
	"net"
	"streming_server/protocol/rtcp"
	"streming_server/server/util"
	"strings"
	"time"
)

const DefaultRtcpInterval = 400
const Port = "19001"

type RtcpReceiver struct {
	Ticker          *time.Ticker
	Interval        time.Duration
	UdpCon          *net.PacketConn
	CongestionLevel int
	buffer          []byte
	doneCheck       chan bool
}

func NewRtcpReceiver(clientAddress net.Addr) *RtcpReceiver {
	addressAndPort := strings.Split(clientAddress.String(), ":")
	address := fmt.Sprintf("%v:%v", addressAndPort[0], Port)
	udpConn, _ := net.ListenPacket("udp", address)

	return &RtcpReceiver{
		Interval:        DefaultRtcpInterval * time.Millisecond,
		UdpCon:          &udpConn,
		buffer:          make([]byte, 24),
		doneCheck:       make(chan bool),
		CongestionLevel: util.NoCongestion,
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
