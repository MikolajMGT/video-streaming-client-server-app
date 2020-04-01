package main

import (
	"os"
	"streming_server/client/client"
)

func main() {
	serverAddress := os.Args[1]
	serverPort := os.Args[2]
	videoFileName := os.Args[3]

	client.NewRtspClient(serverAddress, serverPort, videoFileName)
}
