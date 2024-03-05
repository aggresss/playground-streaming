package main

import (
	"time"

	"github.com/pion/ice/v2"
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/flexfec"
	"github.com/pion/interceptor/pkg/nack"
	"github.com/pion/interceptor/pkg/pacer"
	"github.com/pion/webrtc/v3"
)

type TransportParams struct {
	Configuration      webrtc.Configuration
	ICEUDPMux          ice.UDPMux
	ICETCPMux          ice.TCPMux
	ICELite            bool
	ICEProtocolPolicy  webrtc.ICEProtocolPolicy
	NAT1To1IPs         []string
	EnabledAudioCodecs []webrtc.RTPCodecParameters
	EnabledVideoCodecs []webrtc.RTPCodecParameters
	EnableFlexFEC      bool
	IsSendSide         bool
}

func createPeerConnection(params *TransportParams) (pc *webrtc.PeerConnection, err error) {
	// SettingsEngine
	settingsEngine := webrtc.SettingEngine{}
	if len(params.NAT1To1IPs) > 0 {
		settingsEngine.SetNAT1To1IPs(params.NAT1To1IPs, webrtc.ICECandidateTypeHost)
	}
	if params.ICEUDPMux != nil {
		settingsEngine.SetICEUDPMux(params.ICEUDPMux)
	}
	if params.ICETCPMux != nil {
		settingsEngine.SetICETCPMux(params.ICETCPMux)
	}
	settingsEngine.SetNetworkTypes([]webrtc.NetworkType{
		webrtc.NetworkTypeUDP4,
		webrtc.NetworkTypeTCP4,
	})
	settingsEngine.SetLite(params.ICELite)
	settingsEngine.SetTrackLocalRtx(true)
	settingsEngine.SetICEProtocolPolicy(params.ICEProtocolPolicy)
	if params.ICEProtocolPolicy != webrtc.ICEProtocolPolicyPreferTCP {
		settingsEngine.SetTrackLocalFlexfec(params.EnableFlexFEC)
	}
	// MediaEngine
	mediaEngine := &webrtc.MediaEngine{}
	for _, codec := range params.EnabledAudioCodecs {
		if err := mediaEngine.RegisterCodec(codec, webrtc.RTPCodecTypeAudio); err != nil {
			return nil, err
		}
	}
	for _, codec := range params.EnabledVideoCodecs {
		if err := mediaEngine.RegisterCodec(codec, webrtc.RTPCodecTypeVideo); err != nil {
			return nil, err
		}
	}
	// InterceptorRegistry
	interceptorRegistry := &interceptor.Registry{}
	// Configure Pacer
	if params.IsSendSide {
		pacer, err := pacer.NewInterceptor()
		if err != nil {
			return nil, err
		}
		interceptorRegistry.Add(pacer)
	}
	// Configure FlexFEC
	if params.EnableFlexFEC && params.IsSendSide {
		flexFec, err := flexfec.NewFecInterceptor()
		if err != nil {
			return nil, err
		}
		interceptorRegistry.Add(flexFec)
	}
	// Configure Nack
	mediaEngine.RegisterFeedback(webrtc.RTCPFeedback{Type: "nack"}, webrtc.RTPCodecTypeVideo)
	mediaEngine.RegisterFeedback(webrtc.RTCPFeedback{Type: "nack", Parameter: "pli"}, webrtc.RTPCodecTypeVideo)
	if params.IsSendSide {
		responder, err := nack.NewResponderInterceptor(
			nack.ResponderSize(1024),
		)
		if err != nil {
			return nil, err
		}
		interceptorRegistry.Add(responder)
	} else {
		generator, err := nack.NewGeneratorInterceptor(
			nack.GeneratorSize(512),
			nack.GeneratorSkipLastN(0),
			nack.GeneratorMaxNacksPerPacket(0),
			nack.GeneratorInterval(time.Millisecond*40),
		)
		if err != nil {
			return nil, err
		}
		interceptorRegistry.Add(generator)
	}
	// Configure RTCP Reports
	if err := webrtc.ConfigureRTCPReports(interceptorRegistry); err != nil {
		return nil, err
	}
	// Configure TWCC Sender
	if params.IsSendSide {
		if err := webrtc.ConfigureTWCCSender(mediaEngine, interceptorRegistry); err != nil {
			return nil, err
		}
	}

	return webrtc.NewAPI(
		webrtc.WithSettingEngine(settingsEngine),
		webrtc.WithMediaEngine(mediaEngine),
		webrtc.WithInterceptorRegistry(interceptorRegistry)).NewPeerConnection(params.Configuration)
}
