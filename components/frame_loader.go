package components

import (
	"streming_server/protocol/rtp"
	"streming_server/video"
)

type FrameLoader struct {
	frameSync      *video.FrameSync
	privateChannel chan *rtp.Packet
	started        bool
	doneCheck      chan bool
}

func NewFrameLoader(frameSync *video.FrameSync, privateChannel chan *rtp.Packet) *FrameLoader {
	return &FrameLoader{
		frameSync:      frameSync,
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
				fl.frameSync.AddFrame(packet.Payload, packet.Header.SequenceNumber)
			}
		}
	}()
}

func (fl *FrameLoader) Stop() {
	if fl.started {
		fl.doneCheck <- true
	}
}
