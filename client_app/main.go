package main

import (
	"os"
	"streming_server/components"
)

func main() {
	serverAddress := os.Args[1]
	serverPort := os.Args[2]
	videoFileName := os.Args[3]

	components.NewRtspClient(serverAddress, serverPort, videoFileName)
}
