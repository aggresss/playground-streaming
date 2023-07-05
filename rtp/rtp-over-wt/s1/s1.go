package main

import (
	"context"
	"crypto/tls"
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
)

func main() {

	// Step 01: Create a Webtransport client and audio/video stream

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

	audioStream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		fmt.Println("OpenStreamSync failed:", err.Error())
		return
	}
	defer audioStream.Close()

	videoStream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		fmt.Println("OpenStreamSync failed:", err.Error())
		return
	}
	defer videoStream.Close()

	// Step 02: Listen UDP packet and forward packet to Webtransport

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
	fmt.Printf("listening on %s for audio\n", udpAudio.LocalAddr().String())

	go func() {

	}()

	go func() {

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
