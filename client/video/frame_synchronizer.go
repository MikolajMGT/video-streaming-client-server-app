package video

import (
	"fmt"
	"gopkg.in/karalabe/cookiejar.v2/collections/deque"
)

type FrameSync struct {
	FramesQueue *deque.Deque
	//LastImage     *image.Image
	CurrentSeqNum int
}

func NewFrameSync() *FrameSync {
	return &FrameSync{
		FramesQueue:   deque.New(),
		CurrentSeqNum: 1,
	}
}

func (fs *FrameSync) AddFrame(image []byte, sequentialNumber int) {
	if sequentialNumber > fs.CurrentSeqNum {
		fs.FramesQueue.PushRight(image)
	}
	//if sequentialNumber < fs.CurrentSeqNum {
	//	fs.FramesQueue.PushRight(fs.LastImage)
	//} else if sequentialNumber > fs.CurrentSeqNum {
	//	for i := fs.CurrentSeqNum; i < sequentialNumber; i++ {
	//		fs.FramesQueue.PushRight(fs.LastImage)
	//	}
	//	fs.FramesQueue.PushRight(image)
	//} else {
	//	fs.FramesQueue.PushRight(image)
	//}
}

func (fs *FrameSync) NextFrame() []byte {
	fs.CurrentSeqNum++
	//lastInQueue := fs.FramesQueue.Right()
	//fs.LastImage = lastInQueue.(*image.Image)
	fmt.Println("Size", fs.FramesQueue.Size())
	return fs.FramesQueue.PopLeft().([]byte)
}

func (fs *FrameSync) Empty() bool {
	return fs.FramesQueue.Empty()
}
