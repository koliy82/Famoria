package link

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
)

// probeResult describes the tracks found inside a media file.
type probeResult struct {
	// Tracks holds each track's handler type, e.g. "vide" or "soun".
	Tracks []string
	// Parsed reports whether the file was understood as an ISO-BMFF (MP4)
	// container. When false the track list is meaningless — the file may be a
	// different container, and the caller must not conclude "no audio".
	Parsed bool
}

// HasAudio reports whether a sound track was found.
func (p probeResult) HasAudio() bool { return p.has("soun") }

// HasVideo reports whether a video track was found.
func (p probeResult) HasVideo() bool { return p.has("vide") }

func (p probeResult) has(handler string) bool {
	for _, t := range p.Tracks {
		if t == handler {
			return true
		}
	}
	return false
}

func (p probeResult) String() string {
	if !p.Parsed {
		return "unparsed"
	}
	if len(p.Tracks) == 0 {
		return "no tracks"
	}
	return strings.Join(p.Tracks, ",")
}

// boxes whose children must be searched recursively to reach hdlr. Everything
// else is treated as a leaf and skipped, which keeps the walk cheap: a 25 MB file
// is mostly mdat, and that box is never descended into.
var containerBoxes = map[string]bool{
	"moov": true, "trak": true, "mdia": true, "minf": true,
	"stbl": true, "edts": true, "udta": true, "mvex": true,
}

// probeMP4 inspects an MP4/MOV file and reports which tracks it contains.
//
// This exists because "the video plays without sound" has several causes that
// look identical from the chat, and the byte-level checks tried earlier are not
// trustworthy: the string "soun" appears in plenty of metadata that has nothing
// to do with an actual audio track. Walking moov/trak/mdia/hdlr reads the
// container's own track table, which is the same information a player uses.
//
// The result is deliberately conservative: when the file cannot be parsed as
// ISO-BMFF, Parsed is false and no conclusion about audio may be drawn. Callers
// must treat that as "unknown", not as "missing".
func probeMP4(path string) (probeResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return probeResult{}, err
	}
	defer f.Close()

	var res probeResult
	if err := walkBoxes(f, 0, fileSize(path), 0, &res); err != nil {
		return res, err
	}
	return res, nil
}

func fileSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return fi.Size()
}

// maxDepth bounds recursion so a malformed file cannot exhaust the stack.
const maxDepth = 8

// walkBoxes scans boxes in [offset, end) and records the handler type of every
// hdlr box found, descending into container boxes.
func walkBoxes(r io.ReaderAt, offset, end int64, depth int, res *probeResult) error {
	if depth > maxDepth {
		return nil
	}

	pos := offset
	for pos+8 <= end {
		var header [8]byte
		if _, err := r.ReadAt(header[:], pos); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		size := int64(binary.BigEndian.Uint32(header[0:4]))
		typ := string(header[4:8])
		bodyOffset := pos + 8

		// A size of 1 means a 64-bit largesize follows; 0 means "until EOF".
		// Both are legal but rare for moov, and neither is worth special-casing
		// here beyond not misreading the box extent.
		switch size {
		case 1:
			var large [8]byte
			if _, err := r.ReadAt(large[:], bodyOffset); err != nil {
				return nil
			}
			size = int64(binary.BigEndian.Uint64(large[:]))
			bodyOffset += 8
		case 0:
			size = end - pos
		}

		if size < 8 || pos+size > end {
			// Truncated or corrupt box: stop rather than chase a bogus offset.
			return nil
		}

		switch {
		case typ == "hdlr":
			if handler, ok := readHandlerType(r, bodyOffset); ok {
				res.Parsed = true
				res.Tracks = append(res.Tracks, handler)
			}
		case containerBoxes[typ]:
			res.Parsed = true
			if err := walkBoxes(r, bodyOffset, pos+size, depth+1, res); err != nil {
				return err
			}
		case typ == "ftyp":
			// Confirms this is an ISO-BMFF file even if it has no readable tracks.
			res.Parsed = true
		}

		pos += size
	}
	return nil
}

// readHandlerType extracts the handler_type field from an hdlr box body.
//
// Layout: version+flags (4 bytes), pre_defined (4), handler_type (4).
func readHandlerType(r io.ReaderAt, bodyOffset int64) (string, bool) {
	var buf [12]byte
	if _, err := r.ReadAt(buf[:], bodyOffset); err != nil {
		return "", false
	}
	handler := string(buf[8:12])
	// Reject anything that is not a plausible FourCC, so a misaligned read cannot
	// be mistaken for a real track.
	for _, c := range handler {
		if c < 0x20 || c > 0x7e {
			return "", false
		}
	}
	return handler, true
}

// describeMediaFile returns a short human-readable summary of a media file for
// logging, e.g. "vide,soun (25.3 MB)". It never fails: probing is diagnostic and
// must not be able to break a working upload.
func describeMediaFile(path string) string {
	fi, err := os.Stat(path)
	if err != nil {
		return "unknown"
	}
	res, err := probeMP4(path)
	if err != nil || !res.Parsed {
		return fmt.Sprintf("%.1f MB, tracks unparsed", float64(fi.Size())/1048576)
	}
	return fmt.Sprintf("tracks=[%s] %.1f MB", res, float64(fi.Size())/1048576)
}
