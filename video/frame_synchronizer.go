package video

import (
	"github.com/enriquebris/goconcurrentqueue"
	"log"
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
		CurrentSeqNum: 0,
	}
}

func (fs *FrameSync) AddFrame(image []byte, sequentialNumber int) {
	if sequentialNumber > fs.CurrentSeqNum {
		err := fs.FramesQueue.Enqueue(image)
		if err != nil {
			log.Fatalln("[ERROR] cannot add frame to queue:", err)
		}
	}
}

func (fs *FrameSync) NextFrame() []byte {
	fs.CurrentSeqNum++
	data, err := fs.FramesQueue.DequeueOrWaitForNextElement()
	if err != nil {
		log.Fatalln("[ERROR] cannot get frame from queue:", err)
	}
	return data.([]byte)
}

func (fs *FrameSync) Empty() bool {
	return fs.FramesQueue.GetLen() == 0
}
