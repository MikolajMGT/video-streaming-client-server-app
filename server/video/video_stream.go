package video

import (
	"bytes"
	"gocv.io/x/gocv"
	"image/jpeg"
)

type Stream struct {
	VideoCapture *gocv.VideoCapture
	VideoMat     *gocv.Mat
	FramesNumber int
	FrameCounter int
	FramePeriod  int
}

func NewStream(fileName string) *Stream {
	vc, _ := gocv.VideoCaptureFile(fileName)
	m := gocv.NewMat()

	framesNumber := int(vc.Get(gocv.VideoCaptureFrameCount))
	framePeriod := int(1000 / vc.Get(gocv.VideoCaptureFPS))

	return &Stream{
		VideoCapture: vc,
		VideoMat:     &m,
		FramesNumber: framesNumber,
		FrameCounter: 0,
		FramePeriod:  framePeriod,
	}
}

func (stream *Stream) NextFrame(frameBuffer []byte) int {
	if stream.FrameCounter < stream.FramesNumber {
		stream.FrameCounter++

		_ = stream.VideoCapture.Read(stream.VideoMat)
		frame, _ := stream.VideoMat.ToImage()

		buf := new(bytes.Buffer)
		_ = jpeg.Encode(buf, frame, nil)
		copy(frameBuffer, buf.Bytes())
		return len(buf.Bytes())
	} else {
		return 0
	}
}
