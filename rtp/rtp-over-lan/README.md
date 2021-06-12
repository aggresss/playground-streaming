## rtp over lan

SDP Example:

```
v=0
o=- 0 0 IN IP4 127.0.0.1
s=No Name
t=0 0
a=tool:libavformat 58.76.100
m=video 5004 RTP/AVP 100
c=IN IP4 127.0.0.1
a=rtpmap:100 H264/90000
a=fmtp:100 packetization-mode=1; sprop-parameter-sets=Z2QAH6yyAKALdCAAAAMAIAAABgHjBkk=,aOvMsiw=; profile-level-id=64001F
m=audio 5006 RTP/AVP 101
c=IN IP4 127.0.0.1
a=rtpmap:101 opus/48000/2
a=fmtp:101 sprop-stereo=1

```


对于 H.264 stream，SDP 中 sprop-parameter-sets 为 SPS 和 PPS 经过 Base64 编码后的数据，SDP示例：
```
sprop-parameter-sets=Z0IAKeNQFAe2AtwEBAaQeJEV,aM48gA==
```
第一部分从Base64解码到Base16：
```
67 42 00 29 E3 50 14 07 B6 02 DC 04 04 06 90 78 91 15
```
第二部分（逗号分隔）：
```
68 CE 3C 80
```

### Reference
- https://blog.csdn.net/zhoubotong2012/article/details/86711097
