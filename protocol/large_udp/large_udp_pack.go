package large_udp

import (
	"fmt"
	"net"
)

type LargeUdpPack struct {
	UdpPack    net.PacketConn
	PacketSize int
	CurrSeqNum uint16

	// joining mechanism
	packetsToJoin map[uint16][]byte
}

func NewLargeUdpPack(udpConn net.PacketConn) *LargeUdpPack {
	return &LargeUdpPack{
		UdpPack:       udpConn,
		PacketSize:    DefaultPacketSize,
		CurrSeqNum:    1,
		packetsToJoin: make(map[uint16][]byte),
	}
}

func NewLargeUdpPackWithSize(udpConn net.PacketConn, desiredPacketSize int) *LargeUdpPack {
	return &LargeUdpPack{
		UdpPack:       udpConn,
		PacketSize:    desiredPacketSize,
		CurrSeqNum:    1,
		packetsToJoin: make(map[uint16][]byte),
	}
}

func (luc *LargeUdpPack) ReadFrom(b []byte) (int, bool, error) {
	buffer := make([]byte, 65507)
	n, _, err := luc.UdpPack.ReadFrom(buffer)
	buffer = buffer[:n]

	packet := NewPacketFromBytes(buffer)

	nTotal := n - int(packet.HeaderSize) - 2
	if err != nil {
		return nTotal, false, err
	}

	if packet.Header.CoSeqNumSize == 0 {
		copy(b, packet.Payload)
		return nTotal, true, nil
	}

	luc.packetsToJoin[packet.Header.SeqNum] = packet.Payload

	full := true
	for _, seqNum := range packet.Header.CoSeqNum {
		if seqNum != packet.Header.SeqNum && luc.packetsToJoin[seqNum] == nil {
			full = false
		}
	}

	if full {
		total := 0
		for _, seqNum := range packet.Header.CoSeqNum {
			dataLen := len(luc.packetsToJoin[seqNum])
			copy(b[total:], luc.packetsToJoin[seqNum])
			delete(luc.packetsToJoin, seqNum)
			total += dataLen
		}
		return total, true, nil
	}

	if packet.Header.CoSeqNumSize > 1 {
		fmt.Println("Performed join")
	}
	fmt.Println("map", len(luc.packetsToJoin))
	return nTotal, false, nil
}

//func (luc *LargeUdpPack) ClearMemoryIfPossible(currentSeqNum int) {
//	for k, _ := range luc.packetsToJoin {
//		if k < uint16(currentSeqNum) {
//			delete(luc.packetsToJoin, k)
//		}
//	}
//}
