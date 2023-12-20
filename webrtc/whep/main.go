package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v3"
)

const (
	HTTP_ADDR    = ":8082"
	WHEP_EXT     = ".whep"
	CANDIDATE    = "127.0.0.1"
	ICE_UDP_PORT = 5060
	ICE_TCP_PORT = 5060
)

type whepHandler struct {
	httpAddr   string
	allowExt   string
	candidates []string
	iceUdpPort int
	iceTcpPort int

	locker         sync.RWMutex
	mapWhepClients map[string]*webrtc.PeerConnection
	api            *webrtc.API
}

func (h *whepHandler) Init() error {
	h.mapWhepClients = make(map[string]*webrtc.PeerConnection)

	settingsEngine := webrtc.SettingEngine{}
	settingsEngine.SetNAT1To1IPs(h.candidates, webrtc.ICECandidateTypeHost)
	udplistener, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IP{0, 0, 0, 0},
		Port: h.iceUdpPort,
	})
	if err != nil {
		return err
	}
	iceUdpMux := webrtc.NewICEUDPMux(nil, udplistener)
	settingsEngine.SetICEUDPMux(iceUdpMux)
	tcplistener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.IP{0, 0, 0, 0},
		Port: h.iceTcpPort,
	})
	if err != nil {
		return err
	}
	iceTcpMux := webrtc.NewICETCPMux(nil, tcplistener, 20)
	settingsEngine.SetICETCPMux(iceTcpMux)
	settingsEngine.SetNetworkTypes([]webrtc.NetworkType{webrtc.NetworkTypeTCP4})

	mediaEngine := &webrtc.MediaEngine{}
	if err := mediaEngine.RegisterCodec(
		webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:    webrtc.MimeTypeH264,
				SDPFmtpLine: "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f",
				ClockRate:   90000,
			},
			PayloadType: 96,
		},
		webrtc.RTPCodecTypeVideo); err != nil {
		return err
	}
	if err = mediaEngine.RegisterCodec(
		webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:  webrtc.MimeTypeOpus,
				ClockRate: 48000,
			},
			PayloadType: 111,
		},
		webrtc.RTPCodecTypeAudio); err != nil {
		return err
	}

	interceptorRegistry := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(mediaEngine, interceptorRegistry); err != nil {
		return err
	}

	h.api = webrtc.NewAPI(
		webrtc.WithSettingEngine(settingsEngine),
		webrtc.WithMediaEngine(mediaEngine),
		webrtc.WithInterceptorRegistry(interceptorRegistry))

	return nil
}

func (h *whepHandler) createWhepClient(path string, offer []byte) (answer []byte, err error) {
	h.locker.Lock()
	defer h.locker.Unlock()
	return nil, nil
}

func (h *whepHandler) deleteWhepClient(path string) error {
	h.locker.Lock()
	defer h.locker.Unlock()
	return nil
}

func (h *whepHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if path.Ext(r.URL.Path) != h.allowExt {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodPost:
		scheme := "http://"
		if r.TLS != nil {
			scheme = "https://"
		}
		offer, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		answer, err := h.createWhepClient(r.URL.Path, offer)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Location", strings.Join([]string{scheme, r.Host, r.URL.Path}, ""))
		w.Header().Set("Content-Type", "application/sdp")
		w.WriteHeader(http.StatusCreated)
		w.Write(answer)
		return
	case http.MethodDelete:
		if err := h.deleteWhepClient(r.URL.Path); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	case http.MethodOptions:
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST,DELETE,OPTIONS")
		w.WriteHeader(http.StatusNoContent)
		return
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func main() {
	h := &whepHandler{
		httpAddr:   HTTP_ADDR,
		allowExt:   WHEP_EXT,
		candidates: []string{CANDIDATE},
		iceUdpPort: ICE_UDP_PORT,
		iceTcpPort: ICE_TCP_PORT,
	}
	if err := h.Init(); err != nil {
		log.Panicln(err)
	}
	log.Println("whep demo running", h.httpAddr)
	log.Fatal(http.ListenAndServe(h.httpAddr, h))
}
