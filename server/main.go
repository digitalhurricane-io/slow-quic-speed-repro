package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/logging"
	"github.com/quic-go/quic-go/qlog"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

func Run(port int) {
	udpConn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: port})
	if err != nil {
		log.Fatal("failed to listen on udp")
	}

	// use a sha-256 hash of key as sha-256 is the required 32 bytes long
	statelessResetKey := quic.StatelessResetKey(sha256.Sum256([]byte("a314kjdsaf903245jlsfhww")))

	tr := quic.Transport{
		Conn:              udpConn,
		StatelessResetKey: &statelessResetKey,
	}

	config := &quic.Config{
		MaxIncomingStreams: 200,
		KeepAlivePeriod:    2 * time.Second,
		MaxIdleTimeout:     10 * time.Second,
		Allow0RTT:          true,
		Tracer: func(ctx context.Context, p logging.Perspective, connID quic.ConnectionID) *logging.ConnectionTracer {
			role := "server"
			if p == logging.PerspectiveClient {
				role = "client"
			}
			filename := fmt.Sprintf("./qlog/log_%x_%s_%d.qlog", connID, role, time.Now().UnixMicro())
			f, err := os.Create(filename)
			if err != nil {
				log.Println(fmt.Errorf("failed to create file for qlog: %w", err))
				return nil
			}
			return qlog.NewConnectionTracer(f, p, connID)
		},
	}

	listener, err := tr.Listen(generateTLSConfig(), config)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to start quic listener: %w", err))
	}

	for {
		var conn quic.Connection

		conn, err = listener.Accept(context.Background())
		if err != nil {
			log.Println(fmt.Errorf("failed to get conn from quic listener: %w", err))
			return
		}

		log.Println("peer conn accepted")

		go handleConn(conn)
	}
}

func handleConn(conn quic.Connection) {
	for {
		s, err := conn.AcceptStream(context.Background())
		if err != nil {
			log.Println(fmt.Errorf("failed to accept stream: %w", err))
			return
		}

		go readFromStream(s)
	}
}

func readFromStream(s quic.Stream) {
	var buf = make([]byte, 32_000)
	start := time.Now()
	for {

		n, err := s.Read(buf)
		if err != nil {
			log.Println("failed to read from stream: %w", err)
			return
		}

		log.Printf("read %d bytes from stream. ms between reads: %d", n, time.Since(start).Milliseconds())

		start = time.Now()
	}
}

// GenerateTLSConfig Setup a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"z"},
	}
}
