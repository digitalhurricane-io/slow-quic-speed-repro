package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/logging"
	"github.com/quic-go/quic-go/qlog"
	"log"
	"os"
	"runtime/pprof"
	"time"
)

func Start(addr, logDir string) {

	logFile, err := os.Create(fmt.Sprintf("%s/client_app.log", logDir))
	if err != nil {
		log.Fatal("failed to create log file: ", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	profFile, err := os.Create(fmt.Sprintf("%s/client_cpu_profile.prof", logDir))
	if err != nil {
		log.Fatal("failed to create prof file: ", err)
	}
	defer profFile.Close()

	if err := pprof.StartCPUProfile(profFile); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

	// with these values, android stream times out with:
	// err writing to test stream: timeout: no recent network activity
	//quicConfig := &quic.Config{
	//	KeepAlivePeriod: 2 * time.Second,
	//	MaxIdleTimeout:  10 * time.Second,
	//}

	quicConfig := &quic.Config{
		KeepAlivePeriod: 5 * time.Second,
		MaxIdleTimeout:  20 * time.Second,
		Tracer: func(ctx context.Context, p logging.Perspective, connID quic.ConnectionID) *logging.ConnectionTracer {
			role := "server"
			if p == logging.PerspectiveClient {
				role = "client"
			}
			filename := fmt.Sprintf("%s/log_%x_%s_%d.qlog", logDir, connID, role, time.Now().UnixMicro())
			f, err := os.Create(filename)
			if err != nil {
				log.Println(fmt.Errorf("failed to create file for qlog: %w", err))
				return nil
			}
			return qlog.NewConnectionTracer(f, p, connID)
		},
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"z"}, // must match server
	}

	time.Sleep(time.Second)

	log.Println("connecting to server")

	quicConn, err := quic.DialAddr(context.Background(), addr, tlsConfig, quicConfig)
	if err != nil {
		log.Println(fmt.Errorf("failed to dial server: %w", err))
		//continue
		return
	}

	stream, err := quicConn.OpenStream()
	if err != nil {
		log.Println(fmt.Errorf("failed to open stream: %w", err))
		//continue
		return
	}

	veryStart := time.Now()
	buf := make([]byte, 32_000)
	start := time.Now()
	for i := 0; i < 100; i++ {

		n, err := stream.Write(buf)
		if err != nil {
			log.Println(fmt.Errorf("err writing to test stream: %w", err))
			return
		}

		log.Printf("wrote %d bytes to stream. ms between writes: %d\n", n, time.Since(start).Milliseconds())
		start = time.Now()
	}

	log.Printf("DONE. ms taken to complete: %d", time.Since(veryStart).Milliseconds())

	err = stream.Close()
	if err != nil {
		log.Println(fmt.Errorf("err closing stream: %w", err))
	}

	err = quicConn.CloseWithError(0, "going away")
	if err != nil {
		log.Println(fmt.Errorf("err closing conn: %w", err))
	}
}
