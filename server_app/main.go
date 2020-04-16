package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"streming_server/components"
	"streming_server/protocol/rtp"
	"sync"
)

func main() {
	port := os.Args[1]
	log.Println("[RTSP] Server started")

	mainChannel := make(chan *rtp.Packet)
	serverMap := new(sync.Map)
	listener, _ := net.Listen("tcp", fmt.Sprint(":", port))

	go func(serverMap *sync.Map) {
		for {
			packet := <-mainChannel
			serverMap.Range(
				func(k, v interface{}) bool {
					srv := k.(*components.RtspServer)
					privateChan := v.(chan *rtp.Packet)
					if !srv.IsStreaming {
						return true
					}
					privateChan <- packet
					return true
				},
			)
		}
	}(serverMap)

	for {
		clientConnection, _ := listener.Accept()
		go func(clientConnection net.Conn, serverMap *sync.Map) {
			log.Printf("[RTSP] received new connection from %v", clientConnection.RemoteAddr().String())
			privateChannel := make(chan *rtp.Packet)
			srv := components.NewRtspServer(clientConnection, mainChannel, privateChannel)
			serverMap.LoadOrStore(srv, privateChannel)
			srv.Start()
		}(clientConnection, serverMap)
	}
}
