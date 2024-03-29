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

	go forwardUDPtoStream(udpAudio, streamAudio)

	go forwardUDPtoStream(udpVideo, streamVideo)

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

func forwardUDPtoStream(udpconn net.PacketConn, stream webtransport.Stream) {
	for {
		buf := make([]byte, MAXPAYLOAD)
		n, _, err := udpconn.ReadFrom(buf)
		if err != nil {
			fmt.Printf("recv udp datagram faild:", err)
			return
		}

		bufCopy := make([]byte, 2+n)
		binary.BigEndian.PutUint16(bufCopy, uint16(n))
		copy(bufCopy[2:], buf)

		_, err = stream.Write(bufCopy)
		if err != nil {
			fmt.Println("write data to stream failed:", err.Error())
		}

		fmt.Printf("forward udp to stream: %04d\r", n)
	}
}
