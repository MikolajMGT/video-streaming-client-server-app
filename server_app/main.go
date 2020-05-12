package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"streming_server/components"
	"streming_server/protocol/rtp"
	"streming_server/protocol/rtsp/state"
	"sync"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("[ERROR] incorrect number of arguments, please provide server port")
	}

	port := os.Args[1]
	log.Println("[RTSP] server started")

	mainChannel := make(chan *rtp.Packet)
	serverMap := new(sync.Map)

	sigs := make(chan os.Signal, 1)
	done := make(chan bool)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	listener, err := net.Listen("tcp", fmt.Sprint(":", port))
	if err != nil {
		log.Println("[ERROR] error while opening connection:", err)
	}

	go runDataDisposer(serverMap, mainChannel, done)

	go func(serverMap *sync.Map, listener net.Listener) {
		for {
			clientConnection, err := listener.Accept()
			if err != nil {
				log.Println("[ERROR] error while connecting with client:", err)
			}
			go func(serverMap *sync.Map, clientConnection net.Conn) {
				log.Printf("[RTSP] received new connection from %v", clientConnection.RemoteAddr().String())
				privateChannel := make(chan *rtp.Packet)
				srv := components.NewServer(clientConnection, mainChannel, privateChannel)
				serverMap.LoadOrStore(srv, privateChannel)
				srv.Start()
			}(serverMap, clientConnection)
		}
	}(serverMap, listener)

	<-sigs
	freeResources(serverMap)
	done <- true
	log.Println("[RTSP] Server closed")
}

func runDataDisposer(serverMap *sync.Map, mainChannel chan *rtp.Packet, done chan bool) {
	for {
		select {
		case packet := <-mainChannel:
			serverMap.Range(
				func(k, v interface{}) bool {
					srv := k.(*components.RtspServer)
					privateChan := v.(chan *rtp.Packet)
					if srv.State == state.Playing {
						privateChan <- packet
					} else if srv.State == state.Detached {
						srv.CloseConnection()
						serverMap.Delete(srv)
					}
					return true
				},
			)
		case <-done:
			return
		}
	}
}

func freeResources(serverMap *sync.Map) {
	serverMap.Range(
		func(k, v interface{}) bool {
			srv := k.(*components.RtspServer)
			srv.CloseConnection()
			return true
		},
	)
}
