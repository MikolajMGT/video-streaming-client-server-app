package components

import (
	"bytes"
	"gocv.io/x/gocv"
	"image/jpeg"
	"streming_server/protocol/rtp"
	"streming_server/ui"
	"streming_server/video"
	"time"
)

type Broadcast struct {
	Server       *RtspServer
	FrameSync    *video.FrameSync
	View         *ui.View
	VideoCapture *gocv.VideoCapture
	VideoMat     *gocv.Mat
	Ticker       *time.Ticker
	Interval     time.Duration
	SeqNum       int
	doneCheck    chan bool
	mainChannel  chan *rtp.Packet
	started      bool
}

func NewBroadcast(srv *RtspServer, sync *video.FrameSync, view *ui.View) *Broadcast {
	videoMat := gocv.NewMat()
	return &Broadcast{
		Server:    srv,
		FrameSync: sync,
		View:      view,
		VideoMat:  &videoMat,
		Interval:  30 * time.Millisecond,
		SeqNum:    1,
		doneCheck: make(chan bool),
		started:   false,
	}
}

func (br *Broadcast) nextFrame() {
	br.VideoCapture.Read(br.VideoMat)
	img, _ := br.VideoMat.ToImage()

	buffer := new(bytes.Buffer)
	_ = jpeg.Encode(buffer, img, nil)

	br.FrameSync.AddFrame(buffer.Bytes(), br.SeqNum)
	br.Server.FrameSync.AddFrame(buffer.Bytes(), br.SeqNum)

	br.View.UpdateImage()
	br.SeqNum++
}

func (br *Broadcast) Start() {
	br.VideoCapture, _ = gocv.VideoCaptureDevice(0)
	br.started = true
	br.Ticker = time.NewTicker(br.Interval)

	go func() {
		for {
			select {
			case <-br.doneCheck:
				return
			case <-br.Ticker.C:
				br.nextFrame()
			}
		}
	}()
}

func (br *Broadcast) Stop() {
	if br.started {
		br.doneCheck <- true
		br.Ticker.Stop()
		br.started = false
		_ = br.VideoCapture.Close()
	}
}
