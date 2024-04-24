package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/cc"
	"github.com/pion/interceptor/pkg/flexfec"
	"github.com/pion/interceptor/pkg/nack"
	"github.com/pion/interceptor/pkg/twcc"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/h264reader"
	"github.com/pion/webrtc/v3/pkg/media/oggreader"
)

const (
	HTTP_ADDR           = ":8082"
	CANDIDATE           = "127.0.0.1"
	ICE_UDP_PORT        = 15060
	ICE_TCP_PORT        = 15060
	AUDIO_FILE_NAME     = "output.ogg"
	VIDEO_FILE_NAME     = "output.h264"
	OGG_PAGE_DURATION   = time.Millisecond * 20
	H264_FRAME_DURATION = time.Millisecond * 41
)

type whepHandler struct {
	httpAddr          string
	candidates        []string
	iceUdpPort        int
	iceTcpPort        int
	audioFileName     string
	videoFileName     string
	oggPageDuration   time.Duration
	h264FrameDuration time.Duration

	locker         sync.RWMutex
	mapWhepClients map[string]*webrtc.PeerConnection
	api            *webrtc.API
}

func (h *whepHandler) Init() error {
	h.mapWhepClients = make(map[string]*webrtc.PeerConnection)

	if _, err := os.Stat(h.audioFileName); err != nil {
		return err
	}
	if _, err := os.Stat(h.videoFileName); err != nil {
		return err
	}

	settingsEngine := webrtc.SettingEngine{}
	settingsEngine.SetNAT1To1IPs(h.candidates, webrtc.ICECandidateTypeHost)
	udplistener, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IP{0, 0, 0, 0},
		Port: h.iceUdpPort,
	})
	if err != nil {
		return err
	}
	settingsEngine.SetICEUDPMux(webrtc.NewICEUDPMux(nil, udplistener))
	tcplistener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.IP{0, 0, 0, 0},
		Port: h.iceTcpPort,
	})
	if err != nil {
		return err
	}
	settingsEngine.SetICETCPMux(webrtc.NewICETCPMux(nil, tcplistener, 20))
	settingsEngine.SetNetworkTypes([]webrtc.NetworkType{webrtc.NetworkTypeTCP4})
	settingsEngine.SetLite(true)
	// nack
	settingsEngine.SetTrackLocalRtx(true)
	// flexfec
	settingsEngine.SetTrackLocalFlexfec(true)

	mediaEngine := &webrtc.MediaEngine{}
	if err := mediaEngine.RegisterCodec(
		webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:    webrtc.MimeTypeH264,
				ClockRate:   90000,
				SDPFmtpLine: "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f",
			},
			PayloadType: 96,
		},
		webrtc.RTPCodecTypeVideo); err != nil {
		return err
	}
	if err := mediaEngine.RegisterCodec(
		webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:    "video/rtx",
				ClockRate:   90000,
				SDPFmtpLine: "apt=96",
			},
			PayloadType: 97,
		},
		webrtc.RTPCodecTypeVideo); err != nil {
		return err
	}
	if err := mediaEngine.RegisterCodec(
		webrtc.RTPCodecParameters{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:    "video/flexfec-03",
				ClockRate:   90000,
				SDPFmtpLine: "repair-window=10000000",
			},
			PayloadType: 49,
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
	if err := registerDefaultInterceptors(mediaEngine, interceptorRegistry); err != nil {
		return err
	}

	h.api = webrtc.NewAPI(
		webrtc.WithSettingEngine(settingsEngine),
		webrtc.WithMediaEngine(mediaEngine),
		webrtc.WithInterceptorRegistry(interceptorRegistry))

	return nil
}

func registerDefaultInterceptors(mediaEngine *webrtc.MediaEngine, interceptorRegistry *interceptor.Registry) error {
	// ConfigureNack
	generator, err := nack.NewGeneratorInterceptor(
		nack.GeneratorSize(512),
		nack.GeneratorSkipLastN(0),
		nack.GeneratorMaxNacksPerPacket(0),
		nack.GeneratorInterval(time.Millisecond*100),
	)
	if err != nil {
		return err
	}

	responder, err := nack.NewResponderInterceptor(
		// nack.ResponderSize(1024),
		nack.ResponderSize(2048),
	)
	if err != nil {
		return err
	}

	mediaEngine.RegisterFeedback(webrtc.RTCPFeedback{Type: "nack"}, webrtc.RTPCodecTypeVideo)
	mediaEngine.RegisterFeedback(webrtc.RTCPFeedback{Type: "nack", Parameter: "pli"}, webrtc.RTPCodecTypeVideo)
	interceptorRegistry.Add(responder)
	interceptorRegistry.Add(generator)

	// ConfigureFlexFEC
	flexFec, err := flexfec.NewFecInterceptor()
	if err != nil {
		return err
	}
	interceptorRegistry.Add(flexFec)

	// ConfigureRTCPReports
	if err := webrtc.ConfigureRTCPReports(interceptorRegistry); err != nil {
		return err
	}

	// ConfigureTWCCSender
	if err := webrtc.ConfigureTWCCSender(mediaEngine, interceptorRegistry); err != nil {
		return err
	}

	// ConfigureCC
	cc, err := cc.NewInterceptor(nil)
	if err != nil {
		return err
	}
	interceptorRegistry.Add(cc)

	tf, err := twcc.NewHeaderExtensionInterceptor()
	if err == nil {
		interceptorRegistry.Add(tf)
	}

	return nil
}

func (h *whepHandler) createWhepClient(path, offerStr string) (string, error) {
	h.locker.Lock()
	defer h.locker.Unlock()
	if _, ok := h.mapWhepClients[path]; ok {
		return "", errors.New("whep client already exist")
	}
	pc, err := h.api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return "", err
	}
	videoTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
	if err != nil {
		return "", err
	}
	videoRtpSender, err := pc.AddTrack(videoTrack)
	if err != nil {
		return "", err
	}
	audioTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "pion")
	if err != nil {
		return "", err
	}
	audioRtpSender, err := pc.AddTrack(audioTrack)
	if err != nil {
		return "", err
	}
	iceConnectedCtx, iceConnectedCtxCancel := context.WithCancel(context.Background())
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := videoRtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()
	go func() {
		file, err := os.Open(h.videoFileName)
		if err != nil {
			panic(err)
		}
		defer func() {
			file.Close()
		}()
		h264, err := h264reader.NewReader(file)
		if err != nil {
			panic(err)
		}
		<-iceConnectedCtx.Done()
		ticker := time.NewTicker(h.h264FrameDuration)
		for ; true; <-ticker.C {
			nal, err := h264.NextNAL()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Println(err)
				return
			}
			if err = videoTrack.WriteSample(media.Sample{Data: nal.Data, Duration: h.h264FrameDuration}); err != nil {
				return
			}
		}
	}()
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := audioRtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()
	go func() {
		file, err := os.Open(h.audioFileName)
		if err != nil {
			panic(err)
		}
		defer func() {
			file.Close()
		}()
		ogg, _, err := oggreader.NewWith(file)
		if err != nil {
			panic(err)
		}
		<-iceConnectedCtx.Done()
		var lastGranule uint64
		ticker := time.NewTicker(h.oggPageDuration)
		for ; true; <-ticker.C {
			pageData, pageHeader, err := ogg.ParseNextPage()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Println(err)
				return
			}
			sampleCount := float64(pageHeader.GranulePosition - lastGranule)
			lastGranule = pageHeader.GranulePosition
			sampleDuration := time.Duration((sampleCount/48000)*1000) * time.Millisecond
			if err = audioTrack.WriteSample(media.Sample{Data: pageData, Duration: sampleDuration}); err != nil {
				return
			}
		}
	}()
	pc.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Println("pc state change:", connectionState.String())
		switch connectionState {
		case webrtc.ICEConnectionStateConnected:
			iceConnectedCtxCancel()
		case webrtc.ICEConnectionStateDisconnected, webrtc.ICEConnectionStateFailed:
			h.deleteWhepClient(path)
		}
	})
	if err := pc.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offerStr,
	}); err != nil {
		return "", err
	}
	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		return "", err
	}
	gatherComplete := webrtc.GatheringCompletePromise(pc)
	if err = pc.SetLocalDescription(answer); err != nil {
		return "", err
	}
	<-gatherComplete
	h.mapWhepClients[path] = pc
	return pc.LocalDescription().SDP, nil
}

func (h *whepHandler) deleteWhepClient(path string) error {
	h.locker.Lock()
	defer h.locker.Unlock()
	pc, ok := h.mapWhepClients[path]
	if !ok {
		return errors.New("whep client not exist")
	}
	pc.Close()
	delete(h.mapWhepClients, path)
	return nil
}

func (h *whepHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if originHdr := r.Header.Get("Origin"); originHdr != "" {
		w.Header().Set("Access-Control-Allow-Origin", originHdr)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
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
		answer, err := h.createWhepClient(r.URL.Path, string(offer))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Location", strings.Join([]string{scheme, r.Host, r.URL.Path}, ""))
		w.Header().Set("Content-Type", "application/sdp")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(answer))
		return
	case http.MethodDelete:
		if err := h.deleteWhepClient(r.URL.Path); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	case http.MethodOptions:
		if reqMethodHdr := r.Header.Get("Access-Control-Request-Method"); reqMethodHdr != "" {
			w.Header().Set("Access-Control-Allow-Methods", reqMethodHdr)
		}
		if reqHeadersHdr := r.Header.Get("Access-Control-Request-Headers"); reqHeadersHdr != "" {
			w.Header().Set("Access-Control-Allow-Headers", reqHeadersHdr)
		}
		w.WriteHeader(http.StatusNoContent)
		return
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func main() {
	candidates := []string{os.Getenv("CANDIDATE")}
	if candidates[0] == "" {
		candidates[0] = CANDIDATE
	}
	h := &whepHandler{
		httpAddr:          HTTP_ADDR,
		candidates:        candidates,
		iceUdpPort:        ICE_UDP_PORT,
		iceTcpPort:        ICE_TCP_PORT,
		audioFileName:     AUDIO_FILE_NAME,
		videoFileName:     VIDEO_FILE_NAME,
		oggPageDuration:   OGG_PAGE_DURATION,
		h264FrameDuration: H264_FRAME_DURATION,
	}
	if err := h.Init(); err != nil {
		log.Fatal(err)
	}
	log.Println("whep demo running", h.httpAddr)
	log.Fatal(http.ListenAndServe(h.httpAddr, h))
}
