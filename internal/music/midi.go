package music

import (
	"bytes"
	"encoding/binary"
	"sort"
)

type midiEvent struct {
	tick  int
	order int // events on the same tick: meta, program, note off, note on
	data  []byte
}

// MIDI encodes the piece as a format 1 Standard MIDI File: a conductor track plus melody, harmony, bass and drums.
func (p *Piece) MIDI() []byte {
	tempo := 60_000_000 / p.BPM
	conductor := []midiEvent{
		{0, 0, meta(0x03, []byte("Block "+p.Hash))},
		{0, 0, meta(0x51, []byte{byte(tempo >> 16), byte(tempo >> 8), byte(tempo)})},
		{0, 0, meta(0x58, []byte{byte(p.Beats), 2, 24, 8})},
	}
	tracks := [][]byte{encodeTrack(conductor)}

	for i, t := range p.Tracks() {
		ch := byte(i)
		if t.Instrument.Drum {
			ch = 9
		}
		events := []midiEvent{
			{0, 0, meta(0x03, []byte(t.Instrument.Name))},
		}
		if !t.Instrument.Drum {
			events = append(events, midiEvent{0, 1, []byte{0xC0 | ch, byte(t.Instrument.Program)}})
		}
		for _, n := range t.Notes {
			events = append(events,
				midiEvent{n.Start, 3, []byte{0x90 | ch, byte(n.Pitch), byte(n.Velocity)}},
				midiEvent{n.Start + n.Duration, 2, []byte{0x80 | ch, byte(n.Pitch), 0}},
			)
		}
		tracks = append(tracks, encodeTrack(events))
	}

	var buf bytes.Buffer
	buf.WriteString("MThd")
	binary.Write(&buf, binary.BigEndian, uint32(6))
	binary.Write(&buf, binary.BigEndian, uint16(1))
	binary.Write(&buf, binary.BigEndian, uint16(len(tracks)))
	binary.Write(&buf, binary.BigEndian, uint16(PPQ))
	for _, t := range tracks {
		buf.WriteString("MTrk")
		binary.Write(&buf, binary.BigEndian, uint32(len(t)))
		buf.Write(t)
	}
	return buf.Bytes()
}

func encodeTrack(events []midiEvent) []byte {
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].tick != events[j].tick {
			return events[i].tick < events[j].tick
		}
		return events[i].order < events[j].order
	})
	var b bytes.Buffer
	last := 0
	for _, e := range events {
		b.Write(varLen(e.tick - last))
		b.Write(e.data)
		last = e.tick
	}
	b.Write(varLen(0))
	b.Write([]byte{0xFF, 0x2F, 0x00})
	return b.Bytes()
}

func meta(kind byte, data []byte) []byte {
	out := append([]byte{0xFF, kind}, varLen(len(data))...)
	return append(out, data...)
}

func varLen(n int) []byte {
	out := []byte{byte(n & 0x7F)}
	for n >>= 7; n > 0; n >>= 7 {
		out = append([]byte{byte(n&0x7F) | 0x80}, out...)
	}
	return out
}
