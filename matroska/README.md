

- https://datatracker.ietf.org/doc/html/rfc8794

```
+-------------------------------------------------+
|  VINT_WIDTH  |   VINT_MARKER   |   VINT_DATA    |
+-------------------------------------------------+
```


## [Root Level](https://www.matroska.org/technical/diagram.html)

```
+-------------+
| EBML Header |
+---------------------------+
| Segment     | SeekHead    |
|             |-------------|
|             | Info        |
|             |-------------|
|             | Tracks      |
|             |-------------|
|             | Chapters    |
|             |-------------|
|             | Cluster     |
|             |-------------|
|             | Cues        |
|             |-------------|
|             | Attachments |
|             |-------------|
|             | Tags        |
+---------------------------+
```

建议每 5 秒或者 每 5MB 生成一个 Cluster

## [Matroska Design](https://matroska.org/technical/basics.html)

- Language Codes
- Physical Types
- Lacing

## [Block](https://matroska.org/technical/diagram.html)

```
+----------------------------------+
| Block | Portion of | Data Type   |
|       | a Block    |  - Bit Flag |
|       |--------------------------+
|       | Header     | TrackNumber |
|       |            |-------------|
|       |            | Timestamp   |
|       |            |-------------|
|       |            | Flags       |
|       |            |  - Gap      |
|       |            |  - Lacing   |
|       |            |  - Reserved |
|       |--------------------------|
|       | Optional   | FrameSize   |
|       |--------------------------|
|       | Data       | Frame       |
+----------------------------------+
```

```
Each Cluster MUST contain exactly one Timestamp Element. The Timestamp Element value MUST be stored once per Cluster. The Timestamp Element in the Cluster is relative to the entire Segment. The Timestamp Element SHOULD be the first Element in the Cluster.
```
每个 Cluster 必须包含 timstamp 信息，block header 中的 timestamp 为 其所在 cluster 的相对偏移时间(毫秒)。

```
There is only one reserved word for Element Size encoding, which is an Element Size encoded to all 1's. Such a coding indicates that the size of the Element is unknown, which is a special case that we believe will be useful for live streaming purposes. However, avoid using this reserved word unnecessarily, because it makes parsing slower and more difficult to implement.
```

```
 In the Matroska EBML Schema there are only two elements that allow unknown sizes, Segment and Cluster.
```

- https://github.com/ietf-wg-cellar/ebml-specification/issues/338