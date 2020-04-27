package components

import (
	"fmt"
	"log"
	"net"
	"streming_server/protocol/rtcp"
	"streming_server/util"
	"strings"
	"time"
)

type RtcpReceiver struct {
	ticker          *time.Ticker
	interval        time.Duration
	udpCon          *net.PacketConn
	congestionLevel int
	buffer          []byte
	doneCheck       chan bool
	ServerAddress   string
}

func NewRtcpReceiver(clientAddress net.Addr) *RtcpReceiver {
	addressAndPort := strings.Split(clientAddress.String(), ":")
	address := fmt.Sprintf("%v:%v", addressAndPort[0], 0)
	udpConn, err := net.ListenPacket("udp", address)
	if err != nil {
		log.Fatalln("[RTCP] error while opening connection:", err)
	}
	serverAddress := udpConn.LocalAddr().String()

	return &RtcpReceiver{
		interval:        DefaultRtcpInterval * time.Millisecond,
		udpCon:          &udpConn,
		buffer:          make([]byte, 24),
		doneCheck:       make(chan bool),
		congestionLevel: util.NoCongestion,
		ServerAddress:   serverAddress,
	}
}

func (r *RtcpReceiver) receive() {
	_, _, err := (*r.udpCon).ReadFrom(r.buffer)
	if err != nil {
		log.Println("[RTCP] error while reading packet:", err)
		return
	}
	rtcpPacket := rtcp.NewPacketFromBytes(r.buffer)
	rtcpPacket.Log()

	r.congestionLevel = util.ResolveCongestionLevel(rtcpPacket.FractionLost)
}

func (r *RtcpReceiver) Start() {
	r.ticker = time.NewTicker(r.interval)

	go func() {
		for {
			select {
			case <-r.doneCheck:
				return
			case <-r.ticker.C:
				r.receive()
			}
		}
	}()
}

func (r *RtcpReceiver) Stop() {
	r.doneCheck <- true
	r.ticker.Stop()
}
