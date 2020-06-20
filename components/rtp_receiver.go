package components

import (
	"log"
	"net"
	"streming_server/protocol/rtp"
	"streming_server/ui"
	"streming_server/video"
	"strings"
	"time"
)

const DefaultRtpInterval = 1

type RtpReceiver struct {
	server            *RtspServer
	frameSync         *video.FrameSync
	view              *ui.View
	ticker            *time.Ticker
	interval          time.Duration
	udpCon            net.PacketConn
	highestRecvSeqNum int
	cumulativeLost    int
	recvPacketsNum    int
	totalBytes        int
	doneCheck         chan bool
	startTime         int64
	totalPlayTime     int64
	started           bool
	listeningPort     string
}

func NewRtpReceiver(frameSync *video.FrameSync, view *ui.View) *RtpReceiver {
	udpConn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		log.Fatalln("[RTP] cannot make rtp connection:", err)
	}
	listeningPort := strings.Split(udpConn.LocalAddr().String(), ":")[3]

	return &RtpReceiver{
		frameSync:     frameSync,
		view:          view,
		interval:      DefaultRtpInterval * time.Millisecond,
		udpCon:        udpConn,
		doneCheck:     make(chan bool),
		started:       false,
		listeningPort: listeningPort,
	}
}

func NewRtpReceiverWithServer(server *RtspServer, frameSync *video.FrameSync) *RtpReceiver {
	rtpReceiver := NewRtpReceiver(frameSync, nil)
	rtpReceiver.server = server
	return rtpReceiver
}

func (r *RtpReceiver) SetStartTime(startTime int64) {
	r.startTime = startTime
}

func (r *RtpReceiver) receive() {
	log.Println("[RTP] received packet")
	buf := make([]byte, 65507)
	packetLength, _, err := r.udpCon.ReadFrom(buf)

	if packetLength == 0 {
		return
	}

	if err != nil {
		log.Println("[RTP] error while reading packet:", err)
	}
	buf = buf[:packetLength]

	rtpPacket := rtp.NewPacketFromBytes(buf, packetLength)
	rtpPacket.Header.Log()

	//current unix time in milliseconds
	currentTime := time.Now().UnixNano() / int64(time.Millisecond)
	r.totalPlayTime += currentTime - r.startTime
	r.startTime = currentTime

	r.recvPacketsNum++
	if rtpPacket.Header.SequenceNumber > r.highestRecvSeqNum {
		r.highestRecvSeqNum = rtpPacket.Header.SequenceNumber
	}
	r.cumulativeLost = r.highestRecvSeqNum - r.recvPacketsNum

	dataRate := 0.0
	if r.totalPlayTime != 0 {
		dataRate = float64(r.totalBytes) / (float64(r.totalPlayTime) / 1000)
	}
	r.totalBytes += len(rtpPacket.Payload)

	r.view.UpdateStatistics(r.totalBytes, r.cumulativeLost, dataRate)
	r.frameSync.AddFrame(rtpPacket.Payload, rtpPacket.Header.SequenceNumber)
}

func (r *RtpReceiver) receiveAndForward() {
	log.Println("[RTP] received and forwarded rtp packet")
	buf := make([]byte, 65507)
	packetLength, _, err := r.udpCon.ReadFrom(buf)

	if packetLength == 0 {
		return
	}

	if err != nil {
		log.Println("[RTP] error while reading packet:", err)
	}
	buf = buf[:packetLength]

	rtpPacket := rtp.NewPacketFromBytes(buf, packetLength)
	rtpPacket.Header.Log()

	//current unix time in milliseconds
	currentTime := time.Now().UnixNano() / int64(time.Millisecond)
	r.totalPlayTime += currentTime - r.startTime
	r.startTime = currentTime

	r.recvPacketsNum++
	if rtpPacket.Header.SequenceNumber > r.highestRecvSeqNum {
		r.highestRecvSeqNum = rtpPacket.Header.SequenceNumber
	}
	r.cumulativeLost = r.highestRecvSeqNum - r.recvPacketsNum

	r.totalBytes += len(rtpPacket.Payload)
	r.server.mainChannel <- rtpPacket
}

func (r *RtpReceiver) Start() {
	r.started = true
	r.ticker = time.NewTicker(r.interval)
	r.doneCheck = make(chan bool)

	go func() {
		for {
			select {
			case <-r.doneCheck:
				return
			case <-r.ticker.C:
				if r.server != nil {
					r.receiveAndForward()
				} else {
					r.receive()
				}
			}
		}
	}()
}

func (r *RtpReceiver) Stop() {
	if r.started {
		close(r.doneCheck)
		r.ticker.Stop()
		r.started = false
	}
}

func (r *RtpReceiver) Close() {
	r.Stop()
	err := r.udpCon.Close()
	if err != nil {
		log.Println("[RTP] error while closing connection:", err)
	}
}
