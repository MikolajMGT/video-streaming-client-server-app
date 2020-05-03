package rtcp

const HeaderSize = 8

type Header struct {
	Version              byte
	Padding              byte
	ReceptionReportCount byte
	PayloadType          byte
	Length               int16
	Ssrc                 int32
}

func NewHeader() *Header {
	return &Header{
		// default used values
		Version:              2,
		Padding:              0,
		ReceptionReportCount: 1,
		PayloadType:          201,
		Length:               32,
		Ssrc:                 9999,
	}
}

func NewHeaderFromBytes(payload []byte) *Header {
	resultHeader := &Header{}
	headerAsBytes := payload[:HeaderSize]

	resultHeader.Version = (headerAsBytes[0]) >> 6
	resultHeader.Padding = headerAsBytes[0] & 32
	resultHeader.ReceptionReportCount = headerAsBytes[0] & 16
	resultHeader.PayloadType = headerAsBytes[1]
	resultHeader.Length =
		(int16(headerAsBytes[2]) << 8) + int16(headerAsBytes[3])
	resultHeader.Ssrc =
		(int32(headerAsBytes[4]) << 24) + (int32(headerAsBytes[5]) << 16) +
			(int32(headerAsBytes[6]) << 8) + int32(headerAsBytes[7])

	return resultHeader
}

func (header *Header) TransformToBytes() [HeaderSize]byte {
	return [HeaderSize]byte{
		header.Version<<6 | header.Padding<<5 | header.ReceptionReportCount,
		header.PayloadType,
		byte(header.Length >> 8),
		byte(header.Length & 0xFF),
		byte(header.Ssrc >> 24),
		byte(header.Ssrc >> 16),
		byte(header.Ssrc >> 8),
		byte(header.Ssrc & 0xFF),
	}
}
