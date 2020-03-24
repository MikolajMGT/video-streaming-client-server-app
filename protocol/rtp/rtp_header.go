package rtp

import (
	"errors"
	"log"
)

const HeaderSize = 12

type Header struct {
	Version        byte
	Padding        byte
	Extension      byte
	CsrcCount      byte
	Marker         byte
	PayloadType    byte
	SequenceNumber int16
	Timestamp      int32
	Ssrc           int32
}

func NewHeader(payloadType byte, sequenceNumber int16, timestamp int32) *Header {
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
		header.Version<<6 | header.Padding<<5 | header.Extension<<4 | header.CsrcCount,
		header.Marker<<7 | header.PayloadType,
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

	resultRtpHeader.Version = (headerAsBytes[0]) >> 6
	resultRtpHeader.Padding = headerAsBytes[0] & 32
	resultRtpHeader.Extension = headerAsBytes[0] & 16
	resultRtpHeader.CsrcCount = headerAsBytes[0] & 8
	resultRtpHeader.Marker = (headerAsBytes[1]) >> 7
	resultRtpHeader.PayloadType = headerAsBytes[1] & 127
	resultRtpHeader.SequenceNumber =
		(int16(headerAsBytes[2]) << 8) + int16(headerAsBytes[3])
	resultRtpHeader.Timestamp =
		(int32(headerAsBytes[4]) << 24) + (int32(headerAsBytes[5]) << 16) +
			(int32(headerAsBytes[6]) << 8) + int32(headerAsBytes[7])
	resultRtpHeader.Ssrc =
		(int32(headerAsBytes[8]) << 24) + (int32(headerAsBytes[9]) << 16) +
			(int32(headerAsBytes[10]) << 8) + int32(headerAsBytes[11])

	return resultRtpHeader, nil
}

func (header Header) Log() {
	log.Printf("RTP Header:\n"+
		"Version: %v, Padding: %v, Extension: %v, CsrcCount: %v, Marker: %v, "+
		"PayloadType: %v, SequenceNumber: %v, TimeStamp: %v, Ssrc: %v",
		header.Version, header.Padding, header.Extension, header.CsrcCount, header.Marker,
		header.PayloadType, header.SequenceNumber, header.Timestamp, header.Ssrc)
}
