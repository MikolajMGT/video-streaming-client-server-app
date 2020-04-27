package components

import (
	"streming_server/ui"
	"streming_server/video"
	"time"
)

type ImageRefresh struct {
	view      *ui.View
	frameSync *video.FrameSync
	ticker    *time.Ticker
	interval  time.Duration
	doneCheck chan bool
	started   bool
}

func NewImageRefresh(view *ui.View, frameSync *video.FrameSync) *ImageRefresh {
	return &ImageRefresh{
		view:      view,
		frameSync: frameSync,
		doneCheck: make(chan bool),
		started:   false,
	}
}

func (ir *ImageRefresh) SetInterval(interval time.Duration) {
	ir.interval = interval
}

func (ir *ImageRefresh) updateImageInGui() {
	if !ir.frameSync.Empty() {
		ir.view.UpdateImage()
	}
}

func (ir *ImageRefresh) Start() {
	ir.started = true
	ir.ticker = time.NewTicker(33 * time.Millisecond)

	go func() {
		for {
			select {
			case <-ir.doneCheck:
				return
			case <-ir.ticker.C:
				ir.updateImageInGui()
			}
		}
	}()
}

func (ir *ImageRefresh) Stop() {
	if ir.started {
		ir.doneCheck <- true
		ir.ticker.Stop()
		ir.started = false
	}
}
