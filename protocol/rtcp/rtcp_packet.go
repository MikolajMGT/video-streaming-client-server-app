package rtcp

import (
	"encoding/binary"
	"log"
	"math"
)

const BodySize = 16

type Packet struct {
	Header         Header
	FractionLost   float64
	CumulativeLost uint32
	HighestSeqNum  uint32
}

func NewPacket(fractionLost float64, cumulativeLost int, highestSeqNum int) *Packet {
	return &Packet{
		Header:         *NewHeader(),
		FractionLost:   fractionLost,
		CumulativeLost: uint32(cumulativeLost),
		HighestSeqNum:  uint32(highestSeqNum),
	}
}

func NewPacketFromBytes(packetAsBytes []byte) *Packet {
	headerBytes := packetAsBytes[:HeaderSize]
	header := NewHeaderFromBytes(headerBytes)
	bits := binary.LittleEndian.Uint64(packetAsBytes[HeaderSize:16])

	return &Packet{
		Header:         *header,
		FractionLost:   math.Float64frombits(bits),
		CumulativeLost: binary.LittleEndian.Uint32(packetAsBytes[16:20]),
		HighestSeqNum:  binary.LittleEndian.Uint32(packetAsBytes[20:24]),
	}
}

func (packet *Packet) TransformToBytes() []byte {
	result := make([]byte, HeaderSize+BodySize)
	headerAsBytes := packet.Header.TransformToBytes()

	for index, item := range headerAsBytes {
		result[index] = item
	}

	bits := math.Float64bits(packet.FractionLost)
	binary.LittleEndian.PutUint64(result[8:16], bits)
	binary.LittleEndian.PutUint32(result[16:20], packet.CumulativeLost)
	binary.LittleEndian.PutUint32(result[20:24], packet.HighestSeqNum)
	return result
}

func (packet *Packet) Log() {
	log.Printf("RTCP:\n"+
		"Fraction Lost: %v, Cumulative Lost: %v, Highest Seq Num: %v",
		packet.FractionLost, packet.CumulativeLost, packet.HighestSeqNum)
}
