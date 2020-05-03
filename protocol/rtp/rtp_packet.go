package rtp

type Packet struct {
	Header      *Header
	PayloadSize int
	Payload     []byte
}

func NewPacket(rtpHeader *Header, payloadSize int, payload []byte) *Packet {
	return &Packet{
		Header:      rtpHeader,
		PayloadSize: payloadSize,
		Payload:     payload,
	}
}

func NewPacketFromBytes(packetAsBytes []byte, packetSize int) *Packet {
	headerBytes := packetAsBytes[:HeaderSize]
	header := NewHeaderFromBytes(headerBytes)

	return &Packet{
		Header:      header,
		PayloadSize: packetSize - HeaderSize,
		Payload:     packetAsBytes[HeaderSize:packetSize],
	}
}

func (packet *Packet) TransformToBytes() []byte {
	result := make([]byte, 0, HeaderSize+len(packet.Payload))
	headerAsBytes := packet.Header.TransformToBytes()

	result = append(result[0:HeaderSize], headerAsBytes...)
	result = append(result[HeaderSize:], packet.Payload...)

	return result
}

func (packet *Packet) getLength() int {
	return HeaderSize + len(packet.Payload)
}
