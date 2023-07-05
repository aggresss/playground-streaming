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
	AUDIOUDPPORT     = "5104"
	VIDEOUDPPORT     = "5106"
	MAXPAYLOAD       = 1500
)

func main() {

	// Step 01: Create UPD dial

	udpAudio, err := net.ResolveUDPAddr("udp", HOST+":"+AUDIOUDPPORT)
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

	udpVideo, err := net.ResolveUDPAddr("udp", HOST+":"+VIDEOUDPPORT)
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

		var audioStream, videoStream webtransport.Stream

		for {
			stream, err := conn.AcceptStream(context.Background())
			if err != nil {
				fmt.Println("failed to accept directional stream:", err.Error())
			}
			defer stream.Close()

			fmt.Printf("accept new stream, remote: %s, streamID: %x\n", conn.RemoteAddr().String(), stream.StreamID())

			if audioStream == nil {
				audioStream = stream
				go forwardStreamtoUDP(audioStream, audioUDPConn)
			} else if videoStream == nil {
				videoStream = stream
				go forwardStreamtoUDP(videoStream, videoUDPConn)
			} else {
				continue
			}
		}
	})

	fmt.Printf("listening on %s\n", s.H3.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}

func forwardStreamtoUDP(stream webtransport.Stream, udpconn *net.UDPConn) {
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

		_, err = udpconn.Write(buf)
		if err != nil {
			fmt.Printf("write data failed: %v\n", err)
		}
	}
}
