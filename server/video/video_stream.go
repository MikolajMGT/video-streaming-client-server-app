package video

import (
	"bytes"
	"gocv.io/x/gocv"
	"image/jpeg"
	"streming_server/protocol/rtp"
)

const UdpMaxPayloadSize = 65535 - rtp.HeaderSize

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
		//gocv.Resize(*stream.VideoMat, stream.VideoMat, image.Point{}, 0.5, 0.5, gocv.InterpolationLinear)

		frame, _ := stream.VideoMat.ToImage()
		buf := new(bytes.Buffer)
		_ = jpeg.Encode(buf, frame, nil)
		//length := len(buf.Bytes())
		//for length > UdpMaxPayloadSize {
		//	fmt.Println("old:", len(buf.Bytes()))
		//	gocv.Resize(*stream.VideoMat, stream.VideoMat, image.Point{}, 0.5, 0.5, gocv.InterpolationLinear)
		//	frame, _ = stream.VideoMat.ToImage()
		//	buf = new(bytes.Buffer)
		//	_ = jpeg.Encode(buf, frame, nil)
		//	length = len(buf.Bytes())
		//	fmt.Println("new:", length)
		//}

		copy(frameBuffer, buf.Bytes())
		return len(buf.Bytes())
	} else {
		return 0
	}
}
