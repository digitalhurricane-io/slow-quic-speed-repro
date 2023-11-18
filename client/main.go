package client

import (
	"fmt"
	"log"
	"net"
	"time"
)

func Start(addr string) {

	conn, err := net.Dial("udp", addr)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	var numPacketsToSend = 10_000

	message := make([]byte, 1439) // most quic packets were sent with this size

	start := time.Now()

	var packetCount int
	for i := 0; i < numPacketsToSend; i++ {
		// Send a packet
		_, err = conn.Write(message)
		if err != nil {
			log.Println("Stopping packet sending: ", err)
			return
		}
		packetCount++
	}

	fmt.Printf("Took %d ms to send %d packets\n", time.Since(start).Milliseconds(), numPacketsToSend)
}
