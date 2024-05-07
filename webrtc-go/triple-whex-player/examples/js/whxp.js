'use strict';

// WHEP
const subVideoVideo01 = document.querySelector('#subVideo01');
const subAudioCanvas01 = document.querySelector('#subAudio01');
const subVideoVideo02 = document.querySelector('#subVideo02');
const subAudioCanvas02 = document.querySelector('#subAudio02');
const subVideoVideo03 = document.querySelector('#subVideo03');
const subAudioCanvas03 = document.querySelector('#subAudio03');

const whepUrlTextarea01 = document.querySelector('#whepUrl01');
const whepUrlTextarea02 = document.querySelector('#whepUrl02');
const whepUrlTextarea03 = document.querySelector('#whepUrl03');

const whepStartButton = document.querySelector('#whepStart');
const whepStopButton = document.querySelector('#whepStop');

// Global Variable
var whepClient01 = null;
var whepClient02 = null;
var whepClient03 = null;

// Initialize
window.onload = () => {
  //var baseUrl = window.location.protocol + '//' + window.location.hostname + ':8082';
  var baseUrl = 'http://whep.traitx.cn:8080';
  whepUrlTextarea01.value = baseUrl + '/live/test01.whep';
  whepUrlTextarea02.value = baseUrl + '/live/test02.whep';
  whepUrlTextarea03.value = baseUrl + '/live/test03.whep';
  whepStartButton.addEventListener('click', whepStart);
  whepStopButton.addEventListener('click', whepStop);
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

    this.peerConnection.addTransceiver("audio", {
      direction: "recvonly",
    });
    this.peerConnection.addTransceiver("video", {
      direction: "recvonly",
    });

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

function whepStart() {
  whepClient01 = new WHEPClient(whepUrlTextarea01.value, "", subAudioCanvas01, subVideoVideo01);
  whepClient02 = new WHEPClient(whepUrlTextarea02.value, "", subAudioCanvas02, subVideoVideo02);
  whepClient03 = new WHEPClient(whepUrlTextarea03.value, "", subAudioCanvas03, subVideoVideo03);
}

function whepStop() {
  whepClient01.disconnectStream();
  whepClient02.disconnectStream();
  whepClient03.disconnectStream();
}
