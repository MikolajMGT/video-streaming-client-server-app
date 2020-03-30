package main

import (
	"os"
	"streming_server/server"
)

func main() {
	port := os.Args[1]
	srv := server.NewRtspServer(port)

	srv.Start()
}
