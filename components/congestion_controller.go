package components

import (
	"image/jpeg"
	"log"
	"streming_server/util"
	"streming_server/video"
	"time"
)

const DefaultCongestionInterval = 400

// TODO adjust to client streaming mechanism (temporarily disabled due to lack of compatibility)
type CongestionController struct {
	Ticker              *time.Ticker
	RtpSender           *RtpSender
	rtcpReceiver        *RtcpReceiver
	FrameSync           *video.FrameSync
	QualityAdjuster     *video.QualityAdjuster
	Interval            time.Duration
	doneCheck           chan bool
	prevCongestionLevel int
}

func NewCongestionController(rtcpReceiver *RtcpReceiver, frameSync *video.FrameSync) *CongestionController {
	return &CongestionController{
		rtcpReceiver:        rtcpReceiver,
		FrameSync:           frameSync,
		QualityAdjuster:     video.NewQualityAdjuster(),
		Interval:            DefaultCongestionInterval * time.Millisecond,
		doneCheck:           make(chan bool),
		prevCongestionLevel: util.NoCongestion,
	}
}

func (con *CongestionController) adjustSendRate() {
	if con.prevCongestionLevel != con.rtcpReceiver.CongestionLevel {
		sendDelay := con.FrameSync.FramePeriod +
			con.rtcpReceiver.CongestionLevel*int(float64(con.FrameSync.FramePeriod)*0.1)
		con.RtpSender.UpdateInterval(time.Duration(sendDelay) * time.Millisecond)
		con.prevCongestionLevel = con.rtcpReceiver.CongestionLevel
		log.Println("[congestion] send delay has been changed to ", sendDelay)
	}
}

func (con *CongestionController) AdjustCompressionQuality(frameBuffer []byte, imageLength int) {
	if con.rtcpReceiver.CongestionLevel > util.NoCongestion {
		lowerQuality := jpeg.DefaultQuality - int(jpeg.DefaultQuality*0.15*float64(con.rtcpReceiver.CongestionLevel))
		con.QualityAdjuster.ChangeCompressionQuality(lowerQuality)
		frameBytes := con.QualityAdjuster.Compress(frameBuffer[0:imageLength])
		copy(frameBuffer, frameBytes)
		log.Println("Quality changed to", lowerQuality)
	}
}

func (con *CongestionController) Start() {
	con.Ticker = time.NewTicker(con.Interval)

	go func() {
		for {
			select {
			case <-con.doneCheck:
				return
			case <-con.Ticker.C:
				con.adjustSendRate()
			}
		}
	}()
}

func (con *CongestionController) Stop() {
	con.doneCheck <- true
	con.Ticker.Stop()
}
