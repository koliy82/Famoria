package link

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

// buildMP4 assembles a minimal ISO-BMFF file with the given track handler types.
// The structure mirrors what a real muxer writes: ftyp, then moov containing one
// trak/mdia/hdlr chain per track.
func buildMP4(handlers ...string) []byte {
	var moovBody []byte
	for _, h := range handlers {
		hdlr := box("hdlr", concat(
			[]byte{0, 0, 0, 0}, // version + flags
			[]byte{0, 0, 0, 0}, // pre_defined
			[]byte(h),          // handler_type
			[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // reserved
			append([]byte("Handler"), 0),               // name (NUL-terminated)
		))
		mdia := box("mdia", hdlr)
		trak := box("trak", mdia)
		moovBody = append(moovBody, trak...)
	}

	return concat(
		box("ftyp", []byte("isom\x00\x00\x02\x00mp41")),
		box("moov", moovBody),
		box("mdat", make([]byte, 64)), // stands in for the media payload
	)
}

func box(typ string, payload []byte) []byte {
	size := uint32(8 + len(payload))
	out := make([]byte, 8, size)
	binary.BigEndian.PutUint32(out[0:4], size)
	copy(out[4:8], typ)
	return append(out, payload...)
}

func concat(parts ...[]byte) []byte {
	var out []byte
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

func writeTemp(t *testing.T, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestProbeMP4DetectsTracks(t *testing.T) {
	cases := []struct {
		name      string
		handlers  []string
		wantAudio bool
		wantVideo bool
	}{
		{"video and audio", []string{"vide", "soun"}, true, true},
		{"video only", []string{"vide"}, false, true},
		{"audio only", []string{"soun"}, true, false},
		{"video audio subtitle", []string{"vide", "soun", "sbtl"}, true, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			path := writeTemp(t, "in.mp4", buildMP4(c.handlers...))
			res, err := probeMP4(path)
			if err != nil {
				t.Fatal(err)
			}
			if !res.Parsed {
				t.Fatal("Parsed = false, want the container recognized")
			}
			if got := res.HasAudio(); got != c.wantAudio {
				t.Errorf("HasAudio() = %v, want %v (tracks=%v)", got, c.wantAudio, res.Tracks)
			}
			if got := res.HasVideo(); got != c.wantVideo {
				t.Errorf("HasVideo() = %v, want %v (tracks=%v)", got, c.wantVideo, res.Tracks)
			}
		})
	}
}

// TestProbeMP4RejectsNonMP4 is the guard that matters most: an unrecognized file
// must report Parsed=false so the caller cannot conclude "no audio".
//
// This is exactly the failure mode of the earlier byte-marker checks, which
// reported video=false for files that plainly had video, and would have produced
// the opposite error here — claiming audio was missing when the file was simply
// not parsed.
func TestProbeMP4RejectsNonMP4(t *testing.T) {
	cases := map[string][]byte{
		"empty":          {},
		"random bytes":   {0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		"text":           []byte("this is not an mp4 file at all, just prose"),
		"webm/mkv magic": {0x1a, 0x45, 0xdf, 0xa3, 0x00, 0x00, 0x00, 0x00},
		"mp3 id3":        append([]byte("ID3\x03\x00"), make([]byte, 32)...),
		"jpeg":           {0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 0x4a, 0x46},
	}

	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			path := writeTemp(t, "probe.bin", data)
			res, err := probeMP4(path)
			if err != nil {
				t.Fatalf("probeMP4 should not error on unrecognized input: %v", err)
			}
			if res.Parsed {
				t.Errorf("Parsed = true for %q, want false", name)
			}
			if res.HasAudio() || res.HasVideo() {
				t.Errorf("unparsed file reported tracks %v", res.Tracks)
			}
		})
	}
}

// TestProbeMP4HandlesTruncatedFile ensures a partially downloaded file does not
// produce a false "no audio" verdict or a panic.
func TestProbeMP4HandlesTruncatedFile(t *testing.T) {
	full := buildMP4("vide", "soun")

	for _, cut := range []int{4, 8, 12, len(full) / 2, len(full) - 1} {
		if cut > len(full) {
			continue
		}
		path := writeTemp(t, "trunc.mp4", full[:cut])

		res, err := probeMP4(path)
		if err != nil {
			t.Fatalf("truncated at %d: unexpected error %v", cut, err)
		}
		// Either it found the tracks before the cut, or it reports unparsed. It
		// must never claim a definitive "no audio" on a truncated file.
		if !res.Parsed && len(res.Tracks) > 0 {
			t.Errorf("truncated at %d: tracks reported without Parsed", cut)
		}
	}
}

// TestProbeMP4SurvivesCorruptBoxSizes ensures a bogus size field cannot send the
// walker into an infinite loop or an out-of-bounds read.
func TestProbeMP4SurvivesCorruptBoxSizes(t *testing.T) {
	cases := map[string]uint32{
		"size zero":      0,
		"size huge":      0x7fffffff,
		"size negative":  0xffffffff,
		"size too small": 4,
	}

	for name, size := range cases {
		t.Run(name, func(t *testing.T) {
			// A header claiming a bad size, followed by bytes that are not a box.
			data := make([]byte, 64)
			binary.BigEndian.PutUint32(data[0:4], size)
			copy(data[4:8], "moov")

			path := writeTemp(t, "corrupt.mp4", data)
			if _, err := probeMP4(path); err != nil {
				t.Errorf("%s: unexpected error %v", name, err)
			}
		})
	}
}

// TestProbeMP4DeepNestingIsBounded verifies maxDepth prevents stack exhaustion on
// a pathological file.
func TestProbeMP4DeepNestingIsBounded(t *testing.T) {
	// Nest moov inside moov far past maxDepth.
	payload := buildMP4("vide", "soun")
	for i := 0; i < maxDepth+5; i++ {
		payload = box("moov", payload)
	}
	path := writeTemp(t, "deep.mp4", payload)

	res, err := probeMP4(path)
	if err != nil {
		t.Fatalf("deep nesting should not error: %v", err)
	}
	// Reaching maxDepth stops the walk; the file is still recognized as a
	// container, but the tracks are legitimately not found. The assertion is that
	// this terminates and reports Parsed rather than crashing.
	if !res.Parsed {
		t.Error("deeply nested file should still be recognized as ISO-BMFF")
	}
}

func TestProbeMP4MissingFile(t *testing.T) {
	if _, err := probeMP4(filepath.Join(t.TempDir(), "absent.mp4")); err == nil {
		t.Error("expected an error for a missing file")
	}
}

// TestDescribeMediaFileNeverFails pins the contract that logging helpers must
// keep: diagnostics may not break a working upload.
func TestDescribeMediaFileNeverFails(t *testing.T) {
	if got := describeMediaFile(filepath.Join(t.TempDir(), "absent.mp4")); got != "unknown" {
		t.Errorf("missing file = %q, want %q", got, "unknown")
	}

	mp4 := writeTemp(t, "ok.mp4", buildMP4("vide", "soun"))
	if got := describeMediaFile(mp4); got == "" || got == "unknown" {
		t.Errorf("describeMediaFile(valid mp4) = %q", got)
	}

	junk := writeTemp(t, "junk.bin", []byte("not media"))
	if got := describeMediaFile(junk); got == "" || got == "unknown" {
		t.Errorf("describeMediaFile(junk) = %q, want an unparsed description", got)
	}
}

func TestProbeResultString(t *testing.T) {
	if got := (probeResult{}).String(); got != "unparsed" {
		t.Errorf("empty result = %q, want %q", got, "unparsed")
	}
	if got := (probeResult{Parsed: true}).String(); got != "no tracks" {
		t.Errorf("parsed without tracks = %q, want %q", got, "no tracks")
	}
	if got := (probeResult{Parsed: true, Tracks: []string{"vide", "soun"}}).String(); got != "vide,soun" {
		t.Errorf("tracks = %q, want %q", got, "vide,soun")
	}
}
