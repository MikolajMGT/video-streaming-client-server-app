package video

import (
	"gopkg.in/karalabe/cookiejar.v2/collections/deque"
	"image"
)

type FrameSynchronizer struct {
	FramesQueue   *deque.Deque
	LastImage     *image.Image
	CurrentSeqNum int
}

func NewFrameSynchronizer() *FrameSynchronizer {
	return &FrameSynchronizer{
		FramesQueue:   new(deque.Deque),
		CurrentSeqNum: 1,
	}
}

func (fs *FrameSynchronizer) addFrame(image image.Image, sequentialNumber int) {
	if sequentialNumber < fs.CurrentSeqNum {
		fs.FramesQueue.PushRight(fs.LastImage)
	} else if sequentialNumber > fs.CurrentSeqNum {
		for i := fs.CurrentSeqNum; i < sequentialNumber; i++ {
			fs.FramesQueue.PushRight(fs.LastImage)
		}
		fs.FramesQueue.PushRight(image)
	} else {
		fs.FramesQueue.PushRight(image)
	}
}

func (fs *FrameSynchronizer) nextFrame() *image.Image {
	fs.CurrentSeqNum++
	lastInQueue := fs.FramesQueue.Right()
	fs.LastImage = lastInQueue.(*image.Image)
	return fs.FramesQueue.PopLeft().(*image.Image)
}
