package components

import (
	"image/jpeg"
	"log"
	"streming_server/util"
	"streming_server/video"
	"time"
)

const DefaultCongestionInterval = 400

type CongestionController struct {
	ticker              *time.Ticker
	rtpSender           *RtpSender
	rtcpReceiver        *RtcpReceiver
	frameSync           *video.FrameSync
	qualityAdjuster     *video.QualityAdjuster
	interval            time.Duration
	doneCheck           chan bool
	prevCongestionLevel int
}

func NewCongestionController(rtcpReceiver *RtcpReceiver, frameSync *video.FrameSync) *CongestionController {
	return &CongestionController{
		rtcpReceiver:        rtcpReceiver,
		frameSync:           frameSync,
		qualityAdjuster:     video.NewQualityAdjuster(),
		interval:            DefaultCongestionInterval * time.Millisecond,
		doneCheck:           make(chan bool),
		prevCongestionLevel: util.NoCongestion,
	}
}

func (cc *CongestionController) SetRtpSender(rtpSender *RtpSender) {
	cc.rtpSender = rtpSender
}

func (cc *CongestionController) adjustSendRate() {
	if cc.prevCongestionLevel != cc.rtcpReceiver.congestionLevel {
		sendDelay := cc.frameSync.FramePeriod +
			cc.rtcpReceiver.congestionLevel*int(float64(cc.frameSync.FramePeriod)*0.1)
		cc.rtpSender.UpdateInterval(time.Duration(sendDelay) * time.Millisecond)
		cc.prevCongestionLevel = cc.rtcpReceiver.congestionLevel
		log.Println("[CC] send delay has been changed to", sendDelay)
	}
}

func (cc *CongestionController) AdjustCompressionQuality(frameBuffer []byte, imageLength int) {
	if cc.rtcpReceiver.congestionLevel > util.NoCongestion {
		lowerQuality := jpeg.DefaultQuality -
			int(jpeg.DefaultQuality*0.15*float64(cc.rtcpReceiver.congestionLevel))
		cc.qualityAdjuster.ChangeCompressionQuality(lowerQuality)
		frameBytes := cc.qualityAdjuster.Compress(frameBuffer[0:imageLength])
		copy(frameBuffer, frameBytes)
		log.Println("[CC] quality changed to", lowerQuality)
	}
}

func (cc *CongestionController) Start() {
	cc.ticker = time.NewTicker(cc.interval)

	go func() {
		for {
			select {
			case <-cc.doneCheck:
				return
			case <-cc.ticker.C:
				cc.adjustSendRate()
			}
		}
	}()
}

func (cc *CongestionController) Stop() {
	cc.doneCheck <- true
	cc.ticker.Stop()
}
