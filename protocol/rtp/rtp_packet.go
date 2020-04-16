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
	headerBytes := packetAsBytes[:12]
	header, _ := NewHeaderFromBytes(headerBytes)

	return &Packet{
		Header:      header,
		PayloadSize: packetSize - HeaderSize,
		Payload:     packetAsBytes[12:packetSize],
	}
}

func (packet Packet) TransformToBytes() []byte {
	result := make([]byte, HeaderSize+len(packet.Payload))
	headerAsBytes := packet.Header.TransformToBytes()

	var currentIndex int
	for index, item := range headerAsBytes {
		result[index] = item
		currentIndex = index + 1
	}

	for _, item := range packet.Payload {
		result[currentIndex] = item
		currentIndex++
	}

	return result
}

func (packet Packet) getLength() int {
	return HeaderSize + len(packet.Payload)
}
