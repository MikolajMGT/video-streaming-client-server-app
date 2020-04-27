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
	if len(os.Args) < 2 {
		log.Fatalln("[ERROR] incorrect number of arguments, please provide server port")
	}

	port := os.Args[1]
	log.Println("[RTSP] server started")

	mainChannel := make(chan *rtp.Packet)
	serverMap := new(sync.Map)
	listener, err := net.Listen("tcp", fmt.Sprint(":", port))
	if err != nil {
		log.Println("[ERROR] error while opening connection:", err)
	}

	go func(serverMap *sync.Map) {
		for {
			packet := <-mainChannel
			serverMap.Range(
				func(k, v interface{}) bool {
					srv := k.(*components.RtspServer)
					privateChan := v.(chan *rtp.Packet)
					// skip sending packet to yourself
					if !srv.IsStreaming {
						privateChan <- packet
					}
					return true
				},
			)
		}
	}(serverMap)

	for {
		clientConnection, err := listener.Accept()
		if err != nil {
			log.Println("[ERROR] error while connecting with client:", err)
		}
		go func(clientConnection net.Conn, serverMap *sync.Map) {
			log.Printf("[RTSP] received new connection from %v", clientConnection.RemoteAddr().String())
			privateChannel := make(chan *rtp.Packet)
			srv := components.NewServer(clientConnection, mainChannel, privateChannel)
			serverMap.LoadOrStore(srv, privateChannel)
			srv.Start()
		}(clientConnection, serverMap)
	}
}
