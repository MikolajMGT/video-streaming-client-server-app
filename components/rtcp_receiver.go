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
	started         bool
	ServerPort      string
}

func NewRtcpReceiver(clientAddress net.Addr) *RtcpReceiver {
	addressAndPort := strings.Split(clientAddress.String(), ":")
	address := fmt.Sprintf("%v:%v", addressAndPort[0], 0)
	udpConn, err := net.ListenPacket("udp", address)
	if err != nil {
		log.Fatalln("[RTCP] error while opening connection:", err)
	}
	serverPort := strings.Split(udpConn.LocalAddr().String(), ":")[1]

	return &RtcpReceiver{
		interval:        DefaultRtcpInterval * time.Millisecond,
		udpCon:          &udpConn,
		buffer:          make([]byte, 24),
		doneCheck:       make(chan bool),
		congestionLevel: util.NoCongestion,
		ServerPort:      serverPort,
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
	r.started = true
	r.ticker = time.NewTicker(r.interval)
	r.doneCheck = make(chan bool)

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
	if r.started {
		close(r.doneCheck)
		r.ticker.Stop()
		r.started = false
	}
}
