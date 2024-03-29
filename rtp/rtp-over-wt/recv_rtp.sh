#!/usr/bin/env bash
# Recieve RTP stream from UDP network

SDP_FILE="test.sdp"

if [[ $(sed --version 2>&1 | head -n1) =~ "GNU" ]]; then
    sed -i 's/5004/5014/g' ${SDP_FILE}
    sed -i 's/5006/5016/g' ${SDP_FILE}
else
    sed -i '' 's/5004/5014/g' ${SDP_FILE}
    sed -i '' 's/5006/5016/g' ${SDP_FILE}
fi

ffplay -loglevel debug \
    -protocol_whitelist file,rtp,udp \
    -i ${SDP_FILE}
