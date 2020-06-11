package video

import (
	"github.com/kyroy/priority-queue"
)

type FrameSync struct {
	FramesQueue   *pq.PriorityQueue
	FramePeriod   int
	CurrentSeqNum int
}

func NewFrameSync() *FrameSync {
	return &FrameSync{
		FramesQueue:   pq.NewPriorityQueue(),
		FramePeriod:   33,
		CurrentSeqNum: 0,
	}
}

func (fs *FrameSync) AddFrame(image []byte, sequentialNumber int) {
	if sequentialNumber > fs.CurrentSeqNum {
		fs.FramesQueue.Insert(image, float64(sequentialNumber))
	}
}

func (fs *FrameSync) NextFrame() []byte {
	fs.CurrentSeqNum++
	data := fs.FramesQueue.PopLowest()
	return data.([]byte)
}

func (fs *FrameSync) Empty() bool {
	return fs.FramesQueue.Len() == 0
}
