
- https://peach.blender.org/
- https://download.blender.org/peach/bigbuckbunny_movies/

```
curl -OL https://download.blender.org/peach/bigbuckbunny_movies/big_buck_bunny_720p_h264.mov

ffmpeg -i big_buck_bunny_720p_h264.mov \
    -vf "drawtext=fontsize=160:text='%{pts\:hms}:fontsize=96:fontcolor=white:box=1:x=10:y=h-th-10:boxcolor=black@0.5'" \
    -c:v libx264 -s 1280X720 -r 30 -bsf:v h264_mp4toannexb -b:v 2M -max_delay 0 -bf 0 -g 120 -keyint_min 120 -sc_threshold 0 \
    -c:a aac -b:a 192K -ac 2 \
    -f mp4 output.mp4 -y
```
