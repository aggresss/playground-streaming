
- https://peach.blender.org/
- https://download.blender.org/peach/bigbuckbunny_movies/

```
curl -OL https://download.blender.org/peach/bigbuckbunny_movies/big_buck_bunny_720p_h264.mov

ffmpeg -i big_buck_bunny_720p_h264.mov \
    -dn -sn \
    -c:v libx264 -s 1280X720 -r 30 -bsf:v h264_mp4toannexb -b:v 2M -max_delay 0 -bf 0 -g 120 -keyint_min 120 -sc_threshold 0 \
    -c:a aac -b:a 192K -ac 2 \
    -f mp4 output.mp4 -y

ffmpeg -i output.mp4 \
    -vf "drawtext=fontsize=160:text='%{pts\:hms}:fontsize=96:fontcolor=white:box=1:x=10:y=h-th-10:boxcolor=black@0.5'" \
    -c:v libx264 -s 1280X720 -r 30 -bsf:v h264_mp4toannexb -b:v 2M -max_delay 0 -bf 0 -g 120 -keyint_min 120 -sc_threshold 0 \
    -c:a aac -b:a 192K -ac 2 \
    -f mp4 big_buck_bunny_720p_h264_aac_2m_pts.mp4 -y

curl -OL https://download.blender.org/peach/bigbuckbunny_movies/big_buck_bunny_1080p_h264.mov

ffmpeg -i big_buck_bunny_1080p_h264.mov \
    -dn -sn \
    -c:v libx264 -s 1920X1080 -r 30 -bsf:v h264_mp4toannexb -b:v 10M -max_delay 0 -bf 0 -g 120 -keyint_min 120 -sc_threshold 0 \
    -c:a aac -b:a 192K -ac 2 \
    -f mp4 output.mp4 -y

ffmpeg -i output.mp4 \
    -vf "drawtext=fontsize=160:text='%{pts\:hms}:fontsize=96:fontcolor=white:box=1:x=10:y=h-th-10:boxcolor=black@0.5'" \
    -c:v libx264 -s 1920X1080 -r 30 -bsf:v h264_mp4toannexb -b:v 10M -max_delay 0 -bf 0 -g 120 -keyint_min 120 -sc_threshold 0 \
    -c:a aac -b:a 192K -ac 2 \
    -f mp4 big_buck_bunny_1080p_h264_aac_10m_pts.mp4 -y


ffmpeg -i big_buck_bunny_720p_h264_aac_pts.mp4 -frames:v 10 output_frame_%03d.jpg
```