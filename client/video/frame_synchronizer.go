package video

type FrameSync struct {
	FramesMap     map[int][]byte
	CurrentSeqNum int
	LastSeqNum    int
}

func NewFrameSync() *FrameSync {
	return &FrameSync{
		FramesMap:     make(map[int][]byte, 0),
		CurrentSeqNum: 1,
		LastSeqNum:    0,
	}
}

func (fs *FrameSync) AddFrame(image []byte, sequentialNumber int) {
	if sequentialNumber > fs.CurrentSeqNum {
		fs.FramesMap[sequentialNumber] = image
	}
}

func (fs *FrameSync) NextFrame() []byte {
	if fs.LastSeqNum > 0 {
		delete(fs.FramesMap, fs.LastSeqNum)
	}

	img := fs.FramesMap[fs.CurrentSeqNum]
	for img == nil {
		fs.CurrentSeqNum++
		img = fs.FramesMap[fs.CurrentSeqNum]
	}
	fs.LastSeqNum = fs.CurrentSeqNum
	fs.CurrentSeqNum++

	return img
}

func (fs *FrameSync) Empty() bool {
	return len(fs.FramesMap) == 0
}
