package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

const (
	HOST             = "localhost"
	WEBTRANSPORTPORT = "5059"
	AUDIOUDPPORT     = "5014"
	VIDEOUDPPORT     = "5016"
	MAXPAYLOAD       = 1500
)

var hasAudioStream bool

func main() {

	// Step 01: Create UPD dial

	udpAudio, err := net.ResolveUDPAddr("udp", ":"+AUDIOUDPPORT)
	if err != nil {
		println("Resolve UDPAddr failed:", err.Error())
		os.Exit(1)
	}
	audioUDPConn, err := net.DialUDP("udp", nil, udpAudio)
	if err != nil {
		println("dial audio udp failed:", err.Error())
		os.Exit(1)
	}
	defer audioUDPConn.Close()

	udpVideo, err := net.ResolveUDPAddr("udp", ":"+VIDEOUDPPORT)
	if err != nil {
		println("Resolve UDPAddr failed:", err.Error())
		os.Exit(1)
	}
	videoUDPConn, err := net.DialUDP("udp", nil, udpVideo)
	if err != nil {
		println("dial video udp failed:", err.Error())
		os.Exit(1)
	}
	defer videoUDPConn.Close()

	// Step 02: Setup WebTransport Server and Accept stream and forward to UDP

	tlsConf, err := GetTLSConf(time.Now(), time.Now().Add(10*24*time.Hour))
	if err != nil {
		log.Fatal(err)
	}

	wmux := http.NewServeMux()

	s := webtransport.Server{
		H3: http3.Server{
			TLSConfig: tlsConf,
			Addr:      ":" + WEBTRANSPORTPORT,
			Handler:   wmux,
		},
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	defer s.Close()

	wmux.HandleFunc("/streaming", func(w http.ResponseWriter, r *http.Request) {
		conn, err := s.Upgrade(w, r)
		if err != nil {
			log.Printf("upgrading failed: %s", err)
			w.WriteHeader(500)
			return
		}

		stream, err := conn.AcceptStream(context.Background())
		if err != nil {
			log.Fatalf("failed to accept directional stream: %v", err)
		}
		defer stream.Close()
		fmt.Printf("accept new stream, remote: %s, streamID: %x\n", conn.RemoteAddr().String(), stream.StreamID())

		isAudioStream := false
		if !hasAudioStream {
			hasAudioStream = true
			isAudioStream = true
		}

		for {

			// bytes, err := reader.ReadBytes(byte('\n'))
			// if err != nil {
			// 	if err != io.EOF {
			// 		fmt.Println("failed to read data, err:", err)
			// 	}
			// 	return
			// }
			// fmt.Printf("request: %s", bytes)

			// stream.Write(bytes)
			// fmt.Printf("response: %s", bytes)
		}
	})

	fmt.Printf("listening on %s\n", s.H3.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
