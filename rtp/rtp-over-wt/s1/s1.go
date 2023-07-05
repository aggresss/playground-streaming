package main

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

const (
	HOST             = "localhost"
	WEBTRANSPORTPORT = "5059"
	AUDIOUDPPORT     = "5004"
	VIDEOUDPPORT     = "5006"
	MAXPAYLOAD       = 1500
)

func main() {

	// Step 01: Create a WebTransport client and audio/video stream

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
	}

	quicConfig := &quic.Config{}

	d := webtransport.Dialer{
		RoundTripper: &http3.RoundTripper{
			TLSClientConfig: tlsConf,
			QuicConfig:      quicConfig,
		},
	}

	resp, conn, err := d.Dial(context.Background(), fmt.Sprintf("https://%s:%s/streaming", HOST, WEBTRANSPORTPORT), nil)
	if err != nil {
		fmt.Println("Dial failed:", err.Error())
		os.Exit(1)
	}

	defer conn.CloseWithError(0, "")

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Dial response abnormal")
		os.Exit(1)
	}

	streamAudio, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		fmt.Println("OpenStreamSync failed:", err.Error())
		os.Exit(1)
	}
	defer streamAudio.Close()

	streamVideo, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		fmt.Println("OpenStreamSync failed:", err.Error())
		os.Exit(1)
	}
	defer streamVideo.Close()

	// Step 02: Listen UDP packet and forward packet to WebTransport

	udpAudio, err := net.ListenPacket("udp", ":"+AUDIOUDPPORT)
	if err != nil {
		log.Fatal(err)
	}
	defer udpAudio.Close()
	fmt.Printf("listening on %s for audio\n", udpAudio.LocalAddr().String())

	udpVideo, err := net.ListenPacket("udp", ":"+VIDEOUDPPORT)
	if err != nil {
		log.Fatal(err)
	}
	defer udpVideo.Close()
	fmt.Printf("listening on %s for video\n", udpVideo.LocalAddr().String())

	go func() {
		for {
			buf := make([]byte, MAXPAYLOAD)
			_, _, err := udpAudio.ReadFrom(buf)
			if err != nil {
				fmt.Printf("recv udp audio faild:", err)
				return
			}

			bufCopy := make([]byte, 2+len(buf))
			binary.BigEndian.PutUint16(bufCopy, uint16(len(buf)))
			copy(bufCopy[2:], buf)

			_, err = streamAudio.Write(bufCopy)
			if err != nil {
				fmt.Println("write audio data failed:", err.Error())
			}
		}
	}()

	go func() {
		for {
			buf := make([]byte, MAXPAYLOAD)
			_, _, err := udpVideo.ReadFrom(buf)
			if err != nil {
				fmt.Printf("recv udp video faild:", err)
				return
			}

			bufCopy := make([]byte, 2+len(buf))
			binary.BigEndian.PutUint16(bufCopy, uint16(len(buf)))
			copy(bufCopy[2:], buf)

			_, err = streamVideo.Write(bufCopy)
			if err != nil {
				fmt.Println("write video data failed:", err.Error())
			}
		}
	}()

	// Processing

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	for {
		select {
		case s := <-ch:
			fmt.Println(s)
			os.Exit(0)
		}
	}
}
