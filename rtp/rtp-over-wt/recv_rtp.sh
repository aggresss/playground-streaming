#!/usr/bin/env bash
# Recieve RTP stream from UDP network



if [[ $(sed --version 2>&1 | head -n1) =~ "GNU" ]]; then
    sed -i 's/5004/5104/g' test.sdp
    sed -i 's/5006/5106/g' test.sdp
else
    sed -i '' 's/5004/5104/g' test.sdp
    sed -i '' 's/5006/5106/g' test.sdp
fi

ffplay \
    -protocol_whitelist file,rtp,udp \
    -i test.sdp
