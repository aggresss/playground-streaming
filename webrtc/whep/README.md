## WHEP Demo from disk file

### Create H264 Annex-B file named output.h264 and/or output.ogg that contains a Opus track

```
ffmpeg -i $MEDIA_FILE -an -c:v libx264 -s 1280X720 -r 30 -bsf:v h264_mp4toannexb -b:v 2M -max_delay 0 -bf 0 -g 120 -keyint_min 120 -sc_threshold 0 output.h264
ffmpeg -i $MEDIA_FILE -c:a libopus -page_duration 20000 -vn output.ogg
```
