

```
ffmpeg -i $MEDIA_FILE -an -c:v libx264 -s 1280X720 -r 25 -bsf:v h264_mp4toannexb -b:v 2M -max_delay 0 -bf 0 -g 96 -keyint_min 96 -sc_threshold 0 output.h264
ffmpeg -i $MEDIA_FILE -c:a libopus -page_duration 20000 -vn output.ogg
```