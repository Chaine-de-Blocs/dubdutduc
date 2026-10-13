package music

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"
)

var blockHashes = []string{
	"00000000000000000003efccdd987dd6d93ba18327eef8fd4b46d0de863eb14c",
	"000000000000000000051f8864b8eddf483e7d2b941d626ecea1de70fa0bf551",
	"0000000000000000000e760a04fc958a0631d47490b5f111d0d6aca418b9df17",
	"00000000000000000011f9866ca32fbbbb3cfba26af498dcd98c0f013a920021",
	"00000000000000000013f43456fe2e94a0760eaf779912e0fa37dfb64fe4ccdc",
}

func sampleHashes() []string {
	hashes := append([]string{}, blockHashes...)
	for i := range 500 {
		sum := sha256.Sum256(fmt.Appendf(nil, "block-%d", i))
		hashes = append(hashes, hex.EncodeToString(sum[:]))
	}
	return hashes
}

func TestComposeIsDeterministic(t *testing.T) {
	for _, h := range blockHashes {
		if !reflect.DeepEqual(Compose(h), Compose(h)) {
			t.Fatalf("hash %s produced two different pieces", h)
		}
		if !bytes.Equal(Compose(h).MIDI(), Compose(h).MIDI()) {
			t.Fatalf("hash %s produced two different MIDI files", h)
		}
	}
	if reflect.DeepEqual(Compose(blockHashes[0]).Melody, Compose(blockHashes[1]).Melody) {
		t.Fatal("different hashes produced the same melody")
	}
}

func TestRhythmsFillTheBar(t *testing.T) {
	for beats, opts := range bodyRhythms {
		for _, o := range append(opts, phraseEndRhythms[beats]...) {
			if o.v.units() != beats*4 {
				t.Errorf("%d/4 rhythm %v lasts %d sixteenths", beats, o.v, o.v.units())
			}
		}
	}
	for _, set := range []map[int][]weighted[rhythm]{sustainedComping, pulsedComping} {
		for beats, opts := range set {
			for _, o := range opts {
				if o.v.units() != beats*4 {
					t.Errorf("%d/4 comping %v lasts %d sixteenths", beats, o.v, o.v.units())
				}
			}
		}
	}
	for beats, opts := range bassPatterns {
		for _, o := range opts {
			n := 0
			for _, h := range o.v {
				n += h.units
			}
			if n != beats*4 {
				t.Errorf("%d/4 bass pattern %v lasts %d sixteenths", beats, o.v, n)
			}
		}
	}
}

func TestPieceStructure(t *testing.T) {
	for _, h := range sampleHashes() {
		p := Compose(h)
		bar := p.BarTicks()
		if len(p.Bars) != 16 {
			t.Fatalf("%s: %d bars", h, len(p.Bars))
		}

		for _, tr := range p.Tracks() {
			for _, n := range tr.Notes {
				if n.Start/bar != (n.Start+n.Duration-1)/bar {
					t.Fatalf("%s: %s note at %d crosses a barline", h, tr.Instrument.Name, n.Start)
				}
				if n.Start+n.Duration > p.Length() {
					t.Fatalf("%s: %s note ends after the piece", h, tr.Instrument.Name)
				}
			}
		}

		end := 0
		for i, n := range p.Melody.Notes {
			if !p.Scale.Contains(n.Pitch) {
				t.Fatalf("%s: melody pitch %d is outside %s", h, n.Pitch, p.Key())
			}
			if n.Start < end {
				t.Fatalf("%s: melody note %d overlaps the previous note", h, i)
			}
			end = n.Start + n.Duration
		}

		for i, b := range p.Bars {
			found := false
			for _, n := range p.Melody.Notes {
				if n.Start == i*bar {
					found = true
					if !b.Chord.Has(n.Pitch) {
						t.Fatalf("%s: bar %d downbeat %s is not in %s", h, i+1, PitchName(n.Pitch), b.Chord.Name())
					}
				}
			}
			if !found {
				t.Fatalf("%s: bar %d has no downbeat", h, i+1)
			}
		}

		last := p.Melody.Notes[len(p.Melody.Notes)-1]
		if mod(last.Pitch-p.Scale.Root, 12) != 0 {
			t.Fatalf("%s: melody ends on %s in %s", h, PitchName(last.Pitch), p.Key())
		}
		if p.Bars[15].Chord.Degree != 0 {
			t.Fatalf("%s: final chord is %s", h, p.Bars[15].Chord.Roman())
		}
		for _, b := range p.Bars {
			if b.Chord.Quality == Diminished && p.Scale.Mode != Locrian {
				t.Fatalf("%s: diminished chord %s", h, b.Chord.Name())
			}
			// Locrian's tonic is diminished; other Locrian ° chords should stay rare (avoid III°).
			if p.Scale.Mode == Locrian && b.Chord.Quality == Diminished && b.Chord.Degree != 0 {
				t.Fatalf("%s: non-tonic diminished chord %s in Locrian", h, b.Chord.Name())
			}
		}
	}
}

func TestMIDIHeader(t *testing.T) {
	data := Compose(blockHashes[0]).MIDI()
	if string(data[:4]) != "MThd" {
		t.Fatalf("missing MThd header")
	}
	format := binary.BigEndian.Uint16(data[8:10])
	tracks := binary.BigEndian.Uint16(data[10:12])
	division := binary.BigEndian.Uint16(data[12:14])
	if format != 1 || tracks != 5 || division != PPQ {
		t.Fatalf("header format=%d tracks=%d division=%d", format, tracks, division)
	}
	if n := bytes.Count(data, []byte("MTrk")); n != 5 {
		t.Fatalf("found %d MTrk chunks", n)
	}
}

func TestDrumTrack(t *testing.T) {
	for _, h := range sampleHashes()[:50] {
		p := Compose(h)
		if !p.Drums.Instrument.Drum || len(p.Drums.Notes) == 0 {
			t.Fatalf("%s: missing drums", h)
		}
		for _, n := range p.Drums.Notes {
			switch n.Pitch {
			case drumKick, drumSnare, drumHatClosed, drumHatOpen:
			default:
				t.Fatalf("%s: unexpected drum pitch %d", h, n.Pitch)
			}
			if n.Start/p.BarTicks() != (n.Start+n.Duration-1)/p.BarTicks() {
				t.Fatalf("%s: drum note crosses a barline", h)
			}
		}
		if p.Melody.Instrument.Program < 80 {
			t.Fatalf("%s: melody program %d is not a chip lead", h, p.Melody.Instrument.Program)
		}
	}
}
