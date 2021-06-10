package main

import (
	"fmt"

	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v3"
)

func main() {

	// Create a MediaEngine object to configure the supported codec
	m := &webrtc.MediaEngine{}

	// Setup the codecs you want to use.
	// Only support H264 and OPUS
	for _, codec := range []webrtc.RTPCodecParameters{
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     webrtc.MimeTypeOpus,
				ClockRate:    48000,
				Channels:     2,
				SDPFmtpLine:  "minptime=10;useinbandfec=1",
				RTCPFeedback: nil,
			},
			PayloadType: 111,
		},
	} {
		if err := m.RegisterCodec(codec, webrtc.RTPCodecTypeAudio); err != nil {
			panic(err)
		}
	}

	videoRTCPFeedback := []webrtc.RTCPFeedback{
		{Type: webrtc.TypeRTCPFBGoogREMB, Parameter: ""},
		{Type: webrtc.TypeRTCPFBNACK, Parameter: ""},
		{Type: webrtc.TypeRTCPFBNACK, Parameter: "pli"},
		{Type: webrtc.TypeRTCPFBCCM, Parameter: "fir"},
	}

	for _, codec := range []webrtc.RTPCodecParameters{
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     webrtc.MimeTypeH264,
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f",
				RTCPFeedback: videoRTCPFeedback,
			},
			PayloadType: 102,
		},
		{
			RTPCodecCapability: webrtc.RTPCodecCapability{
				MimeType:     "video/rtx",
				ClockRate:    90000,
				Channels:     0,
				SDPFmtpLine:  "apt=102",
				RTCPFeedback: nil,
			},
			PayloadType: 121,
		},
	} {
		if err := m.RegisterCodec(codec, webrtc.RTPCodecTypeVideo); err != nil {
			panic(err)
		}
	}

	for _, extension := range []string{
		"urn:ietf:params:rtp-hdrext:sdes:mid",
		"urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id",
		"urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id",
		"http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time",
		"urn:ietf:params:rtp-hdrext:ssrc-audio-level",
		"urn:ietf:params:rtp-hdrext:toffset",
	} {
		if err := m.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: extension}, webrtc.RTPCodecTypeAudio); err != nil {
			panic(err)
		}
	}

	for _, extension := range []string{
		"urn:ietf:params:rtp-hdrext:sdes:mid",
		"urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id",
		"urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id",
		"http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time",
		"urn:3gpp:video-orientation",
		"urn:ietf:params:rtp-hdrext:toffset",
	} {
		if err := m.RegisterHeaderExtension(webrtc.RTPHeaderExtensionCapability{URI: extension}, webrtc.RTPCodecTypeVideo); err != nil {
			panic(err)
		}
	}

	// Create a InterceptorRegistry. This is the user configurable RTP/RTCP Pipeline.
	// This provides NACKs, RTCP Reports and other features. If you use `webrtc.NewPeerConnection`
	// this is enabled by default. If you are manually managing You MUST create a InterceptorRegistry
	// for each PeerConnection.
	i := &interceptor.Registry{}

	// Use the default set of Interceptors
	if err := webrtc.RegisterDefaultInterceptors(m, i); err != nil {
		panic(err)
	}

	// Prepare the configuration
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlan,
	}

	// Create a new RTCPeerConnection
	peerConnection, err := webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithInterceptorRegistry(i)).NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	audioTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "rtctool-audio", "auth-audio")
	peerConnection.AddTrack(audioTrack)
	videoTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "rtctool-video", "auth-video")
	peerConnection.AddTrack(videoTrack)

	// Create an offer to send to the other process
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", offer)
}
