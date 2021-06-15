package main

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/interceptor"
	"github.com/pion/sdp"
	"github.com/pion/webrtc/v3"
)

func pion_init() {

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
		SDPSemantics:  webrtc.SDPSemanticsUnifiedPlan,
		RTCPMuxPolicy: webrtc.RTCPMuxPolicyRequire,
	}

	// Create a new RTCPeerConnection
	pc, err := webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithInterceptorRegistry(i)).NewPeerConnection(config)
	if err != nil {
		panic(err)
	}
	defer pc.Close()

	for _, codec := range []webrtc.RTPCodecCapability{
		{MimeType: webrtc.MimeTypeOpus},
		{MimeType: webrtc.MimeTypeH264},
	} {
		track, err := webrtc.NewTrackLocalStaticSample(codec, uuid.NewString(), uuid.NewString())
		if err != nil {
			panic(err)
		}
		_, err = pc.AddTrack(track)
		if err != nil {
			panic(err)
		}
	}

	// Create an offer to send to the other process
	offer, err := pc.CreateOffer(nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", offer)

	sd := &sdp.SessionDescription{}
	sd.Unmarshal(offer.SDP)

	fmt.Printf("%#v\n", sd)
	for _, m := range sd.MediaDescriptions {
		fmt.Printf("%#v\n", m)
	}
}

func sdp_gen() {
	sd := sdp.NewJSEPSessionDescription(false)
	sd.WithFingerprint("sha-256", "8E:84:0C:91:60:EF:F6:3D:FC:AB:1C:44:74:BD:95:26:40:01:B5:89:27:44:FE:E1:83:86:60:DD:3B:86:BC:4D")

	mAudio := sdp.NewJSEPMediaDescription("audio", []string{})
	mAudio.WithPropertyAttribute("sendonly")
	mAudio.WithMediaSource(3705385319, "f4dd9165-eea3-dd4c-8c5c-b71d5bcf388c", "stream-label", "label")
	sd.WithMedia(mAudio)

	sdp := sd.Marshal()
	fmt.Println(sdp)
}

func main() {
	// pion_init()
	sdp_gen()
}
