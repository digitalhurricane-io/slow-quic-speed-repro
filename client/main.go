package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/quic-go/quic-go"
	"log"
	"time"
)

func Start(addr string) {

	// with these values, android stream times out with:
	// err writing to test stream: timeout: no recent network activity
	//quicConfig := &quic.Config{
	//	KeepAlivePeriod: 2 * time.Second,
	//	MaxIdleTimeout:  10 * time.Second,
	//}

	quicConfig := &quic.Config{
		KeepAlivePeriod: 5 * time.Second,
		MaxIdleTimeout:  20 * time.Second,
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"z"}, // must match server
	}

	for {
		time.Sleep(time.Second)

		log.Println("connecting to server")

		quicConn, err := quic.DialAddr(context.Background(), addr, tlsConfig, quicConfig)
		if err != nil {
			log.Println(fmt.Errorf("failed to dial server: %w", err))
			continue
		}

		stream, err := quicConn.OpenStream()
		if err != nil {
			log.Println(fmt.Errorf("failed to open stream: %w", err))
			continue
		}

		buf := make([]byte, 1_000)
		start := time.Now()
		for i := 0; i < 500; i++ {

			n, err := stream.Write(buf)
			if err != nil {
				log.Println(fmt.Errorf("err writing to test stream: %w", err))
				return
			}

			fmt.Printf("wrote %d bytes to stream. ms between writes: %d\n", n, time.Since(start).Milliseconds())
			start = time.Now()
		}
	}

}
