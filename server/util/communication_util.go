package util

import (
	"bufio"
	"fmt"
	"log"
	"strings"
)

func FormatHeader(sequentialNumber int) string {
	return fmt.Sprintf(
		"RTSP/1.0 200 OK\r\nCSeq: %v\r\n", sequentialNumber,
	)
}

func PrepareDescribeResponse(sequentialNumber int, rtspDestinationPort string, mjpegType int,
	sessionId string, videoFileName string,
) string {

	control := fmt.Sprintf(
		"v=0\r\nm=video %v RTP/AVP %v\r\na=control:streamid=%v\r\na=mimetypestring;\"video/MJPEG\"\r\n",
		rtspDestinationPort, mjpegType, sessionId,
	)
	content := fmt.Sprintf("Content-Base: %v\r\nContent-Type: application/sdp\r\nContent-Length: %v\r\n",
		videoFileName, len(control),
	)
	return fmt.Sprint(FormatHeader(sequentialNumber), content, control)
}

func PrepareSetupResponse(sequentialNumber int, framePeriod int, serverAddress string) string {
	content := fmt.Sprintf("Frame-Period: %v\r\nServer-Address: %v\r\n", framePeriod, serverAddress)
	return fmt.Sprint(FormatHeader(sequentialNumber), content)
}

func ReadRequestElements(bufferedReader *bufio.Reader) []string {
	request := ""
	for i := 0; i < 3; i++ {
		requestLineBytes, _, err := bufferedReader.ReadLine()
		if err != nil {
			log.Println("Client disconnected.")
			return make([]string, 0)
		}
		requestLine := string(requestLineBytes)
		request += requestLine + " "
		log.Println("\t[RTSP message]", requestLine)
	}
	return strings.Split(request, " ")
}
