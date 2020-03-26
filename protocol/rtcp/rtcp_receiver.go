package rtcp

import (
	"net"
	"streming_server/server/congestion"
	"time"
)

type Receiver struct {
	Ticker          time.Ticker
	Interval        time.Duration
	UdpCon          net.UDPConn
	CongestionLevel int
	buffer          []byte
	doneCheck       chan bool
}

func NewReceiver(interval time.Duration, udpConn net.UDPConn) *Receiver {
	return &Receiver{
		Interval:  interval,
		UdpCon:    udpConn,
		buffer:    []byte{},
		doneCheck: make(chan bool),
	}
}

func (receiver Receiver) receive() {
	// TODO err handling
	_, _, _ = receiver.UdpCon.ReadFromUDP(receiver.buffer)
	rtcpPacket := NewPacketFromBytes(receiver.buffer)
	rtcpPacket.Log()

	receiver.CongestionLevel = congestion.ResolveCongestionLevel(rtcpPacket.FractionLost)
}

func (receiver Receiver) StartReceiving() {
	receiver.Ticker = *time.NewTicker(receiver.Interval)

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

func (receiver Receiver) StopReceiving() {
	receiver.doneCheck <- true
	receiver.Ticker.Stop()
}
