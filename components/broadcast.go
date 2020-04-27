package components

import (
	"bytes"
	"gocv.io/x/gocv"
	"image/jpeg"
	"log"
	"streming_server/ui"
	"streming_server/video"
	"time"
)

type Broadcast struct {
	server       *RtspServer
	frameSync    *video.FrameSync
	view         *ui.View
	videoCapture *gocv.VideoCapture
	videoMat     *gocv.Mat
	ticker       *time.Ticker
	interval     time.Duration
	seqNum       int
	doneCheck    chan bool
	started      bool
}

func NewBroadcast(srv *RtspServer, sync *video.FrameSync, view *ui.View) *Broadcast {
	videoMat := gocv.NewMat()
	return &Broadcast{
		server:    srv,
		frameSync: sync,
		view:      view,
		videoMat:  &videoMat,
		interval:  33 * time.Millisecond,
		seqNum:    1,
		doneCheck: make(chan bool),
		started:   false,
	}
}

func (br *Broadcast) nextFrame() {
	br.videoCapture.Read(br.videoMat)
	img, err := br.videoMat.ToImage()
	if err != nil {
		log.Println("[ERROR] unable to intercept frame from camera:", err)
		return
	}

	buffer := new(bytes.Buffer)
	err = jpeg.Encode(buffer, img, nil)
	if err != nil {
		log.Println("[ERROR] unable to compress frame to jpeg:", err)
		return
	}

	br.frameSync.AddFrame(buffer.Bytes(), br.seqNum)
	br.server.frameSync.AddFrame(buffer.Bytes(), br.seqNum)

	br.view.UpdateImage()
	br.seqNum++
}

func (br *Broadcast) Start() {
	videoCapture, err := gocv.VideoCaptureDevice(0)
	if err != nil {
		log.Fatalln("[ERROR] unable to connect with webcam:", err)
	}

	br.videoCapture = videoCapture
	br.started = true
	br.ticker = time.NewTicker(br.interval)

	go func() {
		for {
			select {
			case <-br.doneCheck:
				return
			case <-br.ticker.C:
				br.nextFrame()
			}
		}
	}()
}

func (br *Broadcast) Stop() {
	if br.started {
		br.doneCheck <- true
		br.ticker.Stop()
		br.started = false

		err := br.videoCapture.Close()
		if err != nil {
			log.Fatalln("[ERROR] cannot disconnect with webcam properly:", err)
		}
	}
}
