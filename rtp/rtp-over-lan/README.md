## rtp over lan

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