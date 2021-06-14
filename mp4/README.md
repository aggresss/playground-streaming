

## root

 Here is a list of the root level atoms contained in MP files:

- ftyp: Contains the file type, description and the common data structures used.
- pdin: Contains progressive video loading/downloading information.
- moov: Container for all the movie metadata.
- moof: Container with video fragments.
- mfra: The container with random access to video fragment
- mdat: Data container for media.
- stts: sample-to-time table.
- stsc: sample-to-chunk table.
- stsz: sample sizes (framing)
- meta: The container with the metadata information.

| L0 | L1 | L2 | L3 | L4 | L5 |Ma|Description
|----|----|----|----|----|----|---|---
|ftyp|    |    |    |    |    |*|file type and compatibility
|pdin|    |    |    |    |    | |progressive download information
|moov|    |    |    |    |    |*|container for all the metadata
|    |mvhd|    |    |    |    |*|movie header, overall declarations
|    |trak|    |    |    |    |*|container for an individual track or stream
|    |    |tkhd|    |    |    |*|track header, overall information about the track
|    |    |tref|    |    |    | |track reference container
|    |    |edts|    |    |    | |edit list container
|    |    |    |elst|    |    | |an edit list
|    |    |mdia|    |    |    |*|container for the media information in a track
|    |    |    |mdhd|    |    |*|mediaheader, overall infomation about the media
|    |    |    |hdlr|    |    |*|handler, declares the media(handler) type
|    |    |    |minf|    |    |*|media information container
|    |    |    |    |vmhd|    | |video media header, overall information(video track only)
|    |    |    |    |smhd|    | |audio media header, overall information(audio track only)
|    |    |    |    |hmhd|    | |hint media header, overall information(hint track only)
|    |    |    |    |nmhd|    | |Null media header, overall information(some tracks only)
|    |    |    |    |dinf|    |*|data information box, container
|    |    |    |    |    |dref|*|data reference box, decleares source of media data in track
|    |    |    |    |stbl|    |*|sample table box, container for the time/space map
|    |    |    |    |    |stsd|*|sample descriptions(codec types, initialization etc.)
|    |    |    |    |    |stts|*|(decoding) time to sample
|    |    |    |    |    |ctts| |(composition) time to sample
|    |    |    |    |    |stsc|*|sample to chunk, partial data offset information
|    |    |    |    |    |stsz| |sample sizes(framing)
|    |    |    |    |    |stz2| |compact sample sizes(framing)
|    |    |    |    |    |stco|*|chunk offset, partial data-offset information
|    |    |    |    |    |co64| |64-bit chunk offset
|    |    |    |    |    |stss| |sync sample table(random access points)
|    |    |    |    |    |stsh| |shadow sync sample table
|    |    |    |    |    |padb| |sample padding bits
|    |    |    |    |    |stdp| |sample degradation priority
|    |    |    |    |    |sdtp| |independent and disposable samples
|    |    |    |    |    |sbgp| |sample to group
|    |    |    |    |    |sgpd| |sample group description
|    |    |    |    |    |subs| |sub-sample information
|    |mvex|    |    |    |    | |movie extends box
|    |    |mehd|    |    |    | |movie extends header box
|    |    |trex|    |    |    |*|track extends defaults
|    |    |ipmc|    |    |    | |IPMP control box
|moof|    |    |    |    |    |*|movie fragment
|    |mfhd|    |    |    |    |*|movie fragment header
|    |traf|    |    |    |    | |track fragment
|    |    |tfhd|    |    |    |*|track fragment header
|    |    |trun|    |    |    | |track fragment run
|    |    |sdtp|    |    |    | |independent and disposable samples
|    |    |sbgp|    |    |    | |sample to group
|    |    |subs|    |    |    | |sub-sample information
|mfra|    |    |    |    |    | |movie fragment random access
|    |tfra|    |    |    |    | |track fragment random access
|    |mfro|    |    |    |    |*|movie fragment random access offset
|mdat|    |    |    |    |    | |media data container
|free|    |    |    |    |    | |free space
|skip|    |    |    |    |    | |free space
|    |udta|    |    |    |    | |user-data
|    |cprt|    |    |    |    | |copyright etc.
|meta|    |    |    |    |    | |metadata
|    |hdlr|    |    |    |    |*|handler, declares the metadata (handler) type
|    |dinf|    |    |    |    | |data information box, container
|    |    |dref|    |    |    | |data reference box, declares source of metadata items
|    |ipmc|    |    |    |    | |IPMP Control Box
|    |imoc|    |    |    |    | |item location
|    |ipro|    |    |    |    | |item protection
|    |    |sinf|    |    |    | |protection scheme information box
|    |    |    |frma|    |    | |original format box
|    |    |    |imif|    |    | |IPMP Information box
|    |    |    |schm|    |    | |scheme type box
|    |    |    |schi|    |    | |scheme information box
|    |iinf|    |    |    |    | |item information
|    |xml |    |    |    |    | |XML container
|    |bxml|    |    |    |    | |binary XML container
|    |pitm|    |    |    |    | |primary item reference
|    |fiin|    |    |    |    | |file delivery item information
|    |    |paen|    |    |    | |partition entry
|    |    |    |fpar|    |    | |file partition
|    |    |    |fecr|    |    | |FEC reservoir
|    |    |segr|    |    |    | |file delivery session group
|    |    |gitn|    |    |    | |group id to name
|    |    |tsel|    |    |    | |track selection
|meco|    |    |    |    |    | |additional metadata container
|    |mere|    |    |    |    | |metabox relation

## Reference

- https://blog.csdn.net/tx3344/article/details/8463375
- https://docs.fileformat.com/video/mp4/
- https://github.com/abema/go-mp4
