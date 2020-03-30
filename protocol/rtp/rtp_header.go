package rtp

import (
	"errors"
	"log"
)

const HeaderSize = 12

type Header struct {
	Version        int
	Padding        int
	Extension      int
	CsrcCount      int
	Marker         int
	PayloadType    int
	SequenceNumber int
	Timestamp      int
	Ssrc           int
}

func NewHeader(payloadType int, sequenceNumber int, timestamp int) *Header {
	return &Header{
		// default values for current implementation
		Version:   2,
		Padding:   0,
		Extension: 0,
		CsrcCount: 0,
		Marker:    0,
		Ssrc:      9999,

		// parameters based on received values
		PayloadType:    payloadType,
		SequenceNumber: sequenceNumber,
		Timestamp:      timestamp,
	}
}

func (header Header) TransformToBytes() [HeaderSize]byte {
	return [HeaderSize]byte{
		byte(header.Version<<6 | header.Padding<<5 | header.Extension<<4 | header.CsrcCount),
		byte(header.Marker<<7 | header.PayloadType),
		byte(header.SequenceNumber >> 8),
		byte(header.SequenceNumber & 0xFF),
		byte(header.Timestamp >> 24),
		byte(header.Timestamp >> 16),
		byte(header.Timestamp >> 8),
		byte(header.Timestamp & 0xFF),
		byte(header.Ssrc >> 24),
		byte(header.Ssrc >> 16),
		byte(header.Ssrc >> 8),
		byte(header.Ssrc & 0xFF),
	}
}

func NewHeaderFromBytes(payload []byte) (*Header, error) {

	if len(payload) < HeaderSize {
		err := errors.New("header is too small, probably broken packet")
		return nil, err
	}

	resultRtpHeader := &Header{}
	headerAsBytes := payload[:HeaderSize]

	resultRtpHeader.Version = int((headerAsBytes[0]) >> 6)
	resultRtpHeader.Padding = int(headerAsBytes[0] & 32)
	resultRtpHeader.Extension = int(headerAsBytes[0] & 16)
	resultRtpHeader.CsrcCount = int(headerAsBytes[0] & 8)
	resultRtpHeader.Marker = int((headerAsBytes[1]) >> 7)
	resultRtpHeader.PayloadType = int(headerAsBytes[1] & 127)
	resultRtpHeader.SequenceNumber =
		(int(headerAsBytes[2]) << 8) + int(headerAsBytes[3])
	resultRtpHeader.Timestamp =
		(int(headerAsBytes[4]) << 24) + (int(headerAsBytes[5]) << 16) +
			(int(headerAsBytes[6]) << 8) + int(headerAsBytes[7])
	resultRtpHeader.Ssrc =
		(int(headerAsBytes[8]) << 24) + (int(headerAsBytes[9]) << 16) +
			(int(headerAsBytes[10]) << 8) + int(headerAsBytes[11])

	return resultRtpHeader, nil
}

func (header Header) Log() {
	log.Printf("RTP Header:\n"+
		"Version: %v, Padding: %v, Extension: %v, CsrcCount: %v, Marker: %v, "+
		"PayloadType: %v, SequenceNumber: %v, TimeStamp: %v, Ssrc: %v",
		header.Version, header.Padding, header.Extension, header.CsrcCount, header.Marker,
		header.PayloadType, header.SequenceNumber, header.Timestamp, header.Ssrc)
}
