#!/usr/bin/env bash
# Recieve RTP stream from UDP network

ffplay \
    -protocol_whitelist file,rtp,udp \
    -i media.sdp
