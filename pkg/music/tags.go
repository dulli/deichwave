package music

import (
	"bytes"
	"os"
	"strings"

	b64 "encoding/base64"
	"encoding/binary"

	"github.com/bogem/id3v2/v2"
	"github.com/jfreymuth/oggvorbis"
	"github.com/mewkiz/flac/meta"
)

// Metadata block body type.
const FLACPicture = 6

func tags_mp3(path string) (SongInfo, error) {
	// Load id3 tags of currently playing song
	tag, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		return SongInfo{}, err
	}
	defer tag.Close()

	sI := SongInfo{
		Artist:   tag.Artist(),
		Title:    tag.Title(),
		Playlist: "",
	}

	pictures := tag.GetFrames(tag.CommonID("Attached picture"))
	pic, ok := pictures[0].(id3v2.PictureFrame)
	if ok {
		sI.Picture = SongPicture{
			Data: pic.Picture,
			Mime: pic.MimeType,
		}
	} else {
		sI.Picture = SongPicture{}
	}
	return sI, nil
}

func tags_ogg(path string) (SongInfo, error) {
	// Load vorbis comments of currently playing song
	in, err := os.Open(path)
	if err != nil {
		return SongInfo{}, err
	}
	com, err := oggvorbis.GetCommentHeader(in)
	if err != nil {
		return SongInfo{}, err
	}

	// Get tag data
	tags := make(map[string]string)
	for _, val := range com.Comments {
		parts := strings.SplitN(val, "=", 2)
		key := parts[0]
		tag := parts[1]
		tags[key] = tag
	}

	sI := SongInfo{
		Artist:   tags["artist"],
		Title:    tags["title"],
		Playlist: "",
	}

	// Retrieve cover art
	data, err := b64.StdEncoding.DecodeString(tags["metadata_block_picture"])
	if err != nil {
		return sI, err
	}

	// Compute block size bytes
	sint := uint32(len(data))
	sbuf := bytes.NewBuffer([]byte{})
	if err := binary.Write(sbuf, binary.BigEndian, sint); err != nil {
		return sI, err
	}
	size := sbuf.Bytes()

	// Construct block header for parsing
	buff := new(bytes.Buffer)
	buff.WriteByte(FLACPicture + 1<<7)
	buff.Write(size[len(size)-3:])
	buff.Write(data)

	// Parse meta block
	block, err := meta.Parse(buff)
	if err != nil {
		return sI, err
	}

	pic, ok := block.Body.(*meta.Picture)
	if ok {
		sI.Picture = SongPicture{
			Data: pic.Data,
			Mime: pic.MIME,
		}
	} else {
		sI.Picture = SongPicture{}
	}
	return sI, nil
}
