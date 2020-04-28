package large_udp

import (
	"log"
	"math"
	"net"
)

const DefaultPacketSize = 512

type LargeUdpConn struct {
	UdpConn    *net.UDPConn
	PacketSize int
	CurrSeqNum uint16
}

func NewLargeUdpConn(udpConn *net.UDPConn) *LargeUdpConn {
	return &LargeUdpConn{
		UdpConn:    udpConn,
		PacketSize: DefaultPacketSize,
		CurrSeqNum: 1,
	}
}

func NewLargeUdpConnWithSize(udpConn *net.UDPConn, desiredPacketSize int) *LargeUdpConn {
	return &LargeUdpConn{
		UdpConn:    udpConn,
		PacketSize: desiredPacketSize,
		CurrSeqNum: 1,
	}
}

func (luc *LargeUdpConn) Write(data []byte) (int, error) {
	dataLength := len(data)
	packetsNumber := int(math.Ceil(float64(dataLength) / (float64(luc.PacketSize))))
	if packetsNumber <= 1 {
		coSeqNums := make([]uint16, 0)
		packet := NewPacket(NewHeader(luc.CurrSeqNum, coSeqNums),
			data,
		)
		n, err := luc.UdpConn.Write(packet.ToBytes())
		luc.CurrSeqNum++
		return n, err
	}

	log.Println("Performed split")

	coSeqNums := make([]uint16, packetsNumber)
	for i := 0; i < packetsNumber; i++ {
		coSeqNums[i] = luc.CurrSeqNum + uint16(i)
	}

	nTotal := 0
	for i := 0; i < packetsNumber; i++ {
		upperBound := (i + 1) * luc.PacketSize
		if upperBound > dataLength {
			upperBound = dataLength
		}

		packet := NewPacket(NewHeader(luc.CurrSeqNum, coSeqNums),
			data[i*luc.PacketSize:upperBound],
		)

		n, err := luc.UdpConn.Write(packet.ToBytes())
		nTotal += n - int(packet.HeaderSize) - 2
		luc.CurrSeqNum++

		if err != nil {
			return nTotal, err
		}
	}

	return nTotal, nil
}
