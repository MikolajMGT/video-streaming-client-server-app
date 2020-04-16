package video

import (
	"github.com/enriquebris/goconcurrentqueue"
)

type FrameSync struct {
	FramesQueue   *goconcurrentqueue.FIFO
	FramePeriod   int
	CurrentSeqNum int
}

func NewFrameSync() *FrameSync {
	return &FrameSync{
		FramesQueue:   goconcurrentqueue.NewFIFO(),
		FramePeriod:   33,
		CurrentSeqNum: 1,
	}
}

func (fs *FrameSync) AddFrame(image []byte, sequentialNumber int) {
	if sequentialNumber > fs.CurrentSeqNum {
		_ = fs.FramesQueue.Enqueue(image)
	}
}

func (fs *FrameSync) NextFrame() []byte {
	fs.CurrentSeqNum++
	data, _ := fs.FramesQueue.DequeueOrWaitForNextElement()
	return data.([]byte)
}

func (fs *FrameSync) Empty() bool {
	return fs.FramesQueue.GetLen() == 0
}
