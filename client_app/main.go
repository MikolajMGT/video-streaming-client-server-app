package main

import (
	"log"
	"os"
	"streming_server/components"
)

func main() {

	if len(os.Args) < 3 {
		log.Fatalln("[ERROR] incorrect number of arguments, provide server address and port")
	}

	serverAddress := os.Args[1]
	serverPort := os.Args[2]
	videoFileName := "livestream"

	components.NewClient(serverAddress, serverPort, videoFileName)
}
