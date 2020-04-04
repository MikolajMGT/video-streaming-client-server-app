package large_udp

import (
	"encoding/binary"
)

type Packet struct {
	Header     *Header
	HeaderSize uint16
	Payload    []byte
}

func NewPacket(header *Header, payload []byte) *Packet {
	return &Packet{
		Header:     header,
		HeaderSize: header.Size(),
		Payload:    payload,
	}
}

func NewPacketFromBytes(data []byte) *Packet {
	headerSize := binary.LittleEndian.Uint16(data[0:2])
	return &Packet{
		HeaderSize: headerSize,
		Header:     NewHeaderFromBytes(data[2:headerSize]),
		Payload:    data[2+headerSize:],
	}
}

func (p *Packet) ToBytes() []byte {
	result := make([]byte, 2+int(p.HeaderSize)+len(p.Payload))
	binary.LittleEndian.PutUint16(result[0:2], p.HeaderSize)
	copy(result[2:2+int(p.HeaderSize)], p.Header.ToBytes())
	copy(result[2+int(p.HeaderSize):], p.Payload)
	return result
}
