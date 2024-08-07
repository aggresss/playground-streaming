'use strict';

// Options
const audioCodecPreferencesSelect = document.querySelector('#audioCodecPreferences');
const videoCodecPreferencesSelect = document.querySelector('#videoCodecPreferences');
// WHIP
const pubVideoVideo = document.querySelector('#pubVideo');
const pubAudioCanvas = document.querySelector('#pubAudio');
const whipUrlTextarea = document.querySelector('#whipUrl');
const whipTokenTextarea = document.querySelector('#whipToken');
const whipStartButton = document.querySelector('#whipStart');
const whipStopButton = document.querySelector('#whipStop');
// WHEP
const subVideoVideo = document.querySelector('#subVideo');
const subAudioCanvas = document.querySelector('#subAudio');
const whepUrlTextarea = document.querySelector('#whepUrl');
const whepTokenTextarea = document.querySelector('#whepToken');
const whepStartButton = document.querySelector('#whepStart');
const whepStopButton = document.querySelector('#whepStop');

// Global Variable
var whipClient = null;
var whepClient = null;

const supportsSetCodecPreferences = window.RTCRtpTransceiver && 'setCodecPreferences' in window.RTCRtpTransceiver.prototype;

// Initialize
window.onload = () => {
  whipUrlTextarea.value = window.location.protocol + '//' + window.location.hostname + ':8082/live/livestream.whip';
  whipStartButton.addEventListener('click', whipStart);
  whipStopButton.addEventListener('click', whipStop);
  whepUrlTextarea.value = window.location.protocol + '//' + window.location.hostname + ':8082/live/livestream.whep';
  whepStartButton.addEventListener('click', whepStart);
  whepStopButton.addEventListener('click', whepStop);
  // Prefernce
  if (supportsSetCodecPreferences) {
    const { codecs } = RTCRtpSender.getCapabilities('audio');
    codecs.forEach(codec => {
      if (['audio/CN', 'audio/telephone-event'].includes(codec.mimeType)) {
        return;
      }
      const option = document.createElement('option');
      option.value = (codec.mimeType + ' ' + codec.clockRate + ' ' +
        (codec.sdpFmtpLine || '')).trim();
      option.innerText = option.value;
      audioCodecPreferencesSelect.appendChild(option);
    });
    audioCodecPreferencesSelect.disabled = false;
  }
  if (supportsSetCodecPreferences) {
    const { codecs } = RTCRtpSender.getCapabilities('video');
    codecs.forEach(codec => {
      if (['video/red', 'video/ulpfec', 'video/rtx'].includes(codec.mimeType)) {
        return;
      }
      const option = document.createElement('option');
      option.value = (codec.mimeType + ' ' + (codec.sdpFmtpLine || '')).trim();
      option.innerText = option.value;
      videoCodecPreferencesSelect.appendChild(option);
    });
    videoCodecPreferencesSelect.disabled = false;
  }
}

class WHIPClient {
  constructor(endpoint, token, audioElement, videoElement) {
    this.endpoint = endpoint;
    this.token = token;
    this.audioElement = audioElement;
    this.videoElement = videoElement;

    this.peerConnection = new RTCPeerConnection({
      iceServers: [
        {
          urls: 'stun:stun.l.google.com:19302'
        }
      ],
      bundlePolicy: 'max-bundle',
      rtcpMuxPolicy: "require",
      iceTransportPolicy: "all"
    });

    console.log('whip peer connection created.')

    this.peerConnection.addEventListener('negotiationneeded', async ev => {
      console.log('Connection negotiation starting');
      await negotiateConnectionWithClientOffer(this.peerConnection, this.endpoint, this.token);
      console.log('Connection negotiation ended');
    });

    this.accessLocalMediaSources().catch(console.error);
  }

  async accessLocalMediaSources() {
    return navigator.mediaDevices.getUserMedia({ video: true, audio: true }).then(stream => {
      stream.getTracks().forEach(track => {
        const transceiver = this.peerConnection.addTransceiver(track, {
          direction: 'sendonly',
        });
        if (!transceiver.sender.track) {
          return
        }
        let ms = new MediaStream([transceiver.sender.track]);
        switch (track.kind) {
          case 'audio':
            if (audioCodecPreferencesSelect.value !== '') {
              const [mimeType, clockRate, sdpFmtpLine] = audioCodecPreferencesSelect.value.split(' ');
              const { codecs } = RTCRtpSender.getCapabilities('audio');
              const selectedCodecIndex = codecs.findIndex(c => c.mimeType === mimeType && c.clockRate === parseInt(clockRate, 10) && c.sdpFmtpLine === sdpFmtpLine);
              const selectedCodec = codecs[selectedCodecIndex];
              codecs.splice(selectedCodecIndex, 1);
              codecs.unshift(selectedCodec);
              transceiver.setCodecPreferences(codecs);
            }
            this.streamVisualizer = new StreamVisualizer(ms, this.audioElement);
            this.streamVisualizer.start();
            break;
          case 'video':
            if (videoCodecPreferencesSelect.value !== '') {
              const [mimeType, sdpFmtpLine] = videoCodecPreferencesSelect.value.split(' ');
              const { codecs } = RTCRtpSender.getCapabilities('video');
              const selectedCodecIndex = codecs.findIndex(c => c.mimeType === mimeType && c.sdpFmtpLine === sdpFmtpLine);
              const selectedCodec = codecs[selectedCodecIndex];
              codecs.splice(selectedCodecIndex, 1);
              codecs.unshift(selectedCodec);
              transceiver.setCodecPreferences(codecs);
            }
            transceiver.sender.track.applyConstraints({
              width: 1280,
              height: 720,
            });
            this.videoElement.srcObject = ms;
            break;
          default:
            break;
        }
      });
      return stream;
    });
  }

  async disconnectStream() {
    this.videoElement.srcObject = null;
    this.streamVisualizer.stop();
    this.streamVisualizer = null;

    var _a;
    const response = await fetch(this.endpoint, {
      method: 'DELETE',
      mode: 'cors',
    });
    this.peerConnection.close();
    (_a = this.localStream) === null || _a === void 0
      ? void 0
      : _a.getTracks().forEach(track => track.stop());
  }
}


class WHEPClient {
  constructor(endpoint, token, audioElement, videoElement) {
    this.endpoint = endpoint;
    this.token = token;
    this.audioElement = audioElement;
    this.videoElement = videoElement;
    this.ms = new MediaStream();

    this.peerConnection = new RTCPeerConnection({
      bundlePolicy: 'max-bundle',
      rtcpMuxPolicy: "require",
      iceTransportPolicy: "all"
    });

    console.log('whep peer connection created.')

    const audioPubTransceiver = this.peerConnection.addTransceiver("audio", {
      direction: "recvonly",
    });
    if (audioCodecPreferencesSelect.value !== '') {
      const [mimeType, clockRate, sdpFmtpLine] = audioCodecPreferencesSelect.value.split(' ');
      const { codecs } = RTCRtpSender.getCapabilities('audio');
      const selectedCodecIndex = codecs.findIndex(c => c.mimeType === mimeType && c.clockRate === parseInt(clockRate, 10) && c.sdpFmtpLine === sdpFmtpLine);
      const selectedCodec = codecs[selectedCodecIndex];
      codecs.splice(selectedCodecIndex, 1);
      codecs.unshift(selectedCodec);
      audioPubTransceiver.setCodecPreferences(codecs);
    }

    const videoPubTransceiver = this.peerConnection.addTransceiver("video", {
      direction: "recvonly",
    });
    if (videoCodecPreferencesSelect.value !== '') {
      const [mimeType, sdpFmtpLine] = videoCodecPreferencesSelect.value.split(' ');
      const { codecs } = RTCRtpSender.getCapabilities('video');
      const selectedCodecIndex = codecs.findIndex(c => c.mimeType === mimeType && c.sdpFmtpLine === sdpFmtpLine);
      const selectedCodec = codecs[selectedCodecIndex];
      codecs.splice(selectedCodecIndex, 1);
      codecs.unshift(selectedCodec);
      videoPubTransceiver.setCodecPreferences(codecs);
    }

    this.peerConnection.ontrack = (event) => {
      const track = event.track;
      switch (track.kind) {
        case "video":
          this.ms.addTrack(track);
          this.videoElement.srcObject = this.ms;
          break;
        case "audio":
          this.ms.addTrack(track);
          this.videoElement.srcObject = this.ms;
          this.streamVisualizer = new StreamVisualizer(this.ms, this.audioElement);
          this.streamVisualizer.start();
          break;
        default:
          console.log("got unknown track " + track);
      }
    };

    this.peerConnection.addEventListener("connectionstatechange", (ev) => {
      if (this.peerConnection.connectionState !== "connected") {
        return;
      }
      if (!this.videoElement.srcObject) {
        this.videoElement.srcObject = this.stream;
      }
    });

    this.peerConnection.addEventListener('negotiationneeded', async ev => {
      console.log('Connection negotiation starting');
      await negotiateConnectionWithClientOffer(this.peerConnection, this.endpoint, this.token);
      console.log('Connection negotiation ended');
    });
  }

  async disconnectStream() {
    this.videoElement.srcObject = null;
    this.streamVisualizer.stop();
    this.streamVisualizer = null;

    var _b;
    const response = await fetch(this.endpoint, {
      method: 'DELETE',
      mode: 'cors',
    });
    this.peerConnection.close();
    (_b = this.localStream) === null || _b === void 0
      ? void 0
      : _b.getTracks().forEach(track => track.stop());
  }
}

// Performs the actual SDP exchange.
async function negotiateConnectionWithClientOffer(peerConnection, endpoint, token) {
  const offer = await peerConnection.createOffer();
  console.log(`whxp client offer sdp:\n%c${offer.sdp}`, 'color:magenta');
  await peerConnection.setLocalDescription(offer);
  while (peerConnection.connectionState !== 'closed') {
    let response = await postSDPOffer(endpoint, token, offer.sdp);
    if (response.status === 201) {
      let answerSDP = await response.text();
      console.log(`whxp client answer sdp:\n%c${answerSDP}`, 'color:cyan');
      await peerConnection.setRemoteDescription(
        new RTCSessionDescription({ type: 'answer', sdp: answerSDP })
      );
      return response.headers.get('Location');
    } else if (response.status === 405) {
      console.error('Update the URL passed into the WHIP or WHEP client');
    } else {
      const errorMessage = await response.text();
      console.error(errorMessage);
    }

    await new Promise(r => setTimeout(r, 5000));
  }
}

async function postSDPOffer(endpoint, token, data) {
  return await fetch(endpoint, {
    method: 'POST',
    mode: 'cors',
    headers: {
      'Content-Type': 'application/sdp',
      'Authorization': 'Bearer ' + token,
    },
    body: data,
  });
}

function whipStart() {
  whipClient = new WHIPClient(whipUrlTextarea.value, whipTokenTextarea.value, pubAudioCanvas, pubVideoVideo);
}

function whipStop() {
  whipClient.disconnectStream();
}

function whepStart() {
  whepClient = new WHEPClient(whepUrlTextarea.value, whepTokenTextarea.value, subAudioCanvas, subVideoVideo);
}

function whepStop() {
  whepClient.disconnectStream();
}
