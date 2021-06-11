#!/usr/bin/env bash
# Send local file as RTP stream to UDP network
# Inspired by the prectice of https://github.com/versatica/mediasoup-demo/blob/v3/broadcasters/ffmpeg.sh

function show_usage()
{
	echo
	echo "USAGE"
	echo "-----"
	echo
	echo "  IP_ADDR=127.0.0.1 MEDIA_FILE=./test.mkv ./send_rtp.sh"
	echo
	echo "  where:"
	echo "  - IP_ADDR is the IP of the RTP send target"
	echo "  - MEDIA_FILE is the path to a audio+video file (such as a .mkv file)"
	echo
	echo "REQUIREMENTS"
	echo "------------"
	echo
	echo "  - ffmpeg: stream audio and video (https://www.ffmpeg.org)"
	echo
}

echo

if [ -z "${IP_ADDR}" ] ; then
	>&2 echo "ERROR: missing IP_ADDR environment variable"
	show_usage
	exit 1
fi

if [ -z "${MEDIA_FILE}" ] ; then
	>&2 echo "ERROR: missing MEDIA_FILE environment variable"
	show_usage
	exit 1
fi

if [ "$(command -v ffmpeg)" == "" ] ; then
	>&2 echo "ERROR: ffmpeg command not found, must install FFmpeg"
	show_usage
	exit 1
fi

AUDIO_SSRC=1111
AUDIO_PT=100
AUDIO_PORT=5004
VIDEO_SSRC=2222
VIDEO_PT=101
VIDEO_PORT=5006

ffmpeg \
    -re \
    -v info \
    -stream_loop -1 \
    -i ${MEDIA_FILE} \
    -map 0:a:0 \
    -c:a copy \
    -map 0:v:0 \
    -c:v copy \
    -f tee \
    "[select=a:f=rtp:ssrc=${AUDIO_SSRC}:payload_type=${AUDIO_PT}]rtp://${IP_ADDR}:${AUDIO_PORT} \
     | \
     [select=v:f=rtp:ssrc=${VIDEO_SSRC}:payload_type=${VIDEO_PT}]rtp://${IP_ADDR}:${VIDEO_PORT}"
