package server

import (
	"log"
	"net"
)

func Run(port int) {

	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("0.0.0.0"),
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	log.Println("UDP server listening on", addr.String())

	var packetCount int
	buf := make([]byte, 32000)

	for {
		// Read from UDP connection
		_, _, err = conn.ReadFrom(buf)
		if err != nil {
			log.Println(err)
			continue
		}
		packetCount++
		log.Printf("Received packet number %d\n", packetCount)
	}

}
