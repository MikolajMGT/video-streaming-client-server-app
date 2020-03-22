package main

import (
	"fmt"
	"streming_server/protocol/rtp"
)

func main() {
	header := rtp.NewRtpHeader(1, 2, 3)
	fmt.Println(header)

	//arr := header.TransformToByteArray()
	//fmt.Println(arr)

	arr2 := []byte{128, 1, 0, 2, 0, 0, 0, 3, 0, 0, 39, 15}
	h2, _ := rtp.NewRtpHeaderFromBytes(arr2)
	fmt.Println(h2)
	h2.Log()
}
