package components

import (
	"streming_server/protocol/rtp"
	"streming_server/video"
	"time"
)

type FrameLoader struct {
	Ticker         *time.Ticker
	FrameSync      *video.FrameSync
	started        bool
	doneCheck      chan bool
	privateChannel chan *rtp.Packet
}

func NewFrameLoader(frameSync *video.FrameSync, privateChannel chan *rtp.Packet) *FrameLoader {
	return &FrameLoader{
		FrameSync:      frameSync,
		started:        false,
		doneCheck:      make(chan bool),
		privateChannel: privateChannel,
	}
}

func (fl *FrameLoader) Start() {
	fl.started = true

	go func() {
		for {
			select {
			case <-fl.doneCheck:
				return
			case packet := <-fl.privateChannel:
				fl.FrameSync.AddFrame(packet.Payload, packet.Header.SequenceNumber)
			}
		}
	}()
}

func (fl *FrameLoader) Stop() {
	if fl.started {
		fl.doneCheck <- true
	}
}
