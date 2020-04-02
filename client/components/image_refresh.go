package components

import (
	"streming_server/client/ui"
	"streming_server/client/video"
	"time"
)

type ImageRefresh struct {
	View      *ui.View
	FrameSync *video.FrameSync
	Ticker    *time.Ticker
	Interval  time.Duration
	doneCheck chan bool
	started   bool
}

func NewImageRefresh(view *ui.View, frameSync *video.FrameSync) *ImageRefresh {
	return &ImageRefresh{
		View:      view,
		FrameSync: frameSync,
		doneCheck: make(chan bool),
		started:   false,
	}
}

func (imgRef *ImageRefresh) updateImageInGui() {
	if !imgRef.FrameSync.Empty() {
		imgRef.View.UpdateImage()
	}
}

func (imgRef *ImageRefresh) Start() {
	imgRef.started = true
	imgRef.Ticker = time.NewTicker(imgRef.Interval)
	//imgRef.doneCheck = make(chan bool)

	go func() {
		for {
			select {
			case <-imgRef.doneCheck:
				return
			case <-imgRef.Ticker.C:
				imgRef.updateImageInGui()
			}
		}
	}()
}

func (imgRef *ImageRefresh) Stop() {
	if imgRef.started {
		imgRef.doneCheck <- true
		imgRef.Ticker.Stop()
		imgRef.started = false
	}
}
