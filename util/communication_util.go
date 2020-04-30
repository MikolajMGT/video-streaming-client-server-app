package util

import (
	"bufio"
	"errors"
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

func PrepareSetupResponse(sequentialNumber int, serverPort string) string {
	content := fmt.Sprintf("Transport: server_port=%v\r\n", serverPort)
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

func ParseParameter(text string, parameterName string) ([]string, error) {
	transportOptions := strings.Split(text, ";")
	for _, option := range transportOptions {
		if strings.HasPrefix(option, parameterName) {
			value := strings.Split(option, "=")[1]
			ports := strings.Split(value, ",")
			return ports, nil
		}
	}
	return make([]string, 0), errors.New("unable to parse parameter")
}
