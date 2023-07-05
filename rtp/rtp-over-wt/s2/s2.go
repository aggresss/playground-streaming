package main

import (
	"bufio"
	"context"
	"encoding/binary"
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

		reader := bufio.NewReader(stream)
		for {
			buf := make([]byte, MAXPAYLOAD)
			header := make([]byte, 2)
			var bytesRead, n int
			var err error

			for bytesRead < 2 {
				if n, err = reader.Read(header[bytesRead:2]); err != nil {
					fmt.Printf("read stream header faild: %v\n", err)
					return
				}
				bytesRead += n
			}

			length := int(binary.BigEndian.Uint16(header))

			if length > cap(buf) {
				fmt.Printf("buf cap limit: %v\n", err)
				return
			}

			bytesRead = 0
			for bytesRead < length {
				if n, err = reader.Read(buf[bytesRead:length]); err != nil {
					fmt.Printf("read stream body faild: %v\n", err)
					return
				}
				bytesRead += n
			}

			if isAudioStream {
				_, err = audioUDPConn.Write(buf)
			} else {
				_, err = videoUDPConn.Write(buf)
			}
			if err != nil {
				fmt.Printf("write data failed: %v\n", err)
				return
			}
		}
	})

	fmt.Printf("listening on %s\n", s.H3.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
