package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"streming_server/server/server"
)

func main() {
	port := os.Args[1]

	log.Println("[RTSP] Server started")
	listener, _ := net.Listen("tcp", fmt.Sprint(":", port))

	for {
		clientConnection, _ := listener.Accept()
		go func(clientConnection net.Conn) {
			log.Printf("[RTSP] received new connection from %v", clientConnection.RemoteAddr().String())
			srv := server.NewRtspServer(clientConnection)
			srv.Start()
		}(clientConnection)
	}
}
