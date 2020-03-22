package rtp

type Packet struct {
	RtpHeader   Header
	PayloadSize uint32
	Payload     []byte
}

func (packet Packet) NewRtpPacket(rtpHeader Header, payloadSize uint32, payload []byte) *Packet {
	return &Packet{
		RtpHeader:   rtpHeader,
		PayloadSize: payloadSize,
		Payload:     payload,
	}
}

func (packet Packet) NewRtpPacketFromBytes(packetAsBytes []byte, packetSize uint32) *Packet {
	headerBytes := packetAsBytes[:12]
	header, _ := NewRtpHeaderFromBytes(headerBytes)

	return &Packet{
		RtpHeader:   *header,
		PayloadSize: packetSize - HeaderSize,
		Payload:     packetAsBytes[12:],
	}
}
