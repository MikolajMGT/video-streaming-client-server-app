package large_udp

import (
	"encoding/binary"
)

type Header struct {
	SeqNum       uint16
	CoSeqNum     []uint16
	CoSeqNumSize uint16
}

func NewHeader(seqNum uint16, coSeqNum []uint16) *Header {
	return &Header{
		SeqNum:       seqNum,
		CoSeqNum:     coSeqNum,
		CoSeqNumSize: uint16(len(coSeqNum)),
	}
}

func NewHeaderFromBytes(data []byte) *Header {

	seqNum := binary.LittleEndian.Uint16(data[0:2])
	coSeqNumSize := binary.LittleEndian.Uint16(data[2:4])
	coSeqNum := make([]uint16, coSeqNumSize)

	lowerBound := 4
	for i := 0; i < int(coSeqNumSize); i++ {
		coSeqNum[i] = binary.LittleEndian.Uint16(data[lowerBound : lowerBound+2])
		lowerBound += 2
	}

	return &Header{
		SeqNum:       seqNum,
		CoSeqNumSize: coSeqNumSize,
		CoSeqNum:     coSeqNum,
	}
}

func (h *Header) ToBytes() []byte {
	result := make([]byte, h.Size())

	binary.LittleEndian.PutUint16(result[0:2], h.SeqNum)
	binary.LittleEndian.PutUint16(result[2:4], h.CoSeqNumSize)

	lowerBound := 4
	for _, val := range h.CoSeqNum {
		binary.LittleEndian.PutUint16(result[lowerBound:lowerBound+2], val)
		lowerBound += 2
	}

	return result
}

func (h *Header) Size() uint16 {
	return h.CoSeqNumSize*2 + 4
}
