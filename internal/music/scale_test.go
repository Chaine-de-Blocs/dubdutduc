package music

import "testing"

func TestModes(t *testing.T) {
	tests := []struct {
		scale Scale
		want  [7]int
	}{
		{Scale{C, Ionian}, [7]int{C, D, E, F, G, A, B}},
		{Scale{D, Dorian}, [7]int{D, E, F, G, A, B, C + 12}},
		{Scale{E, Phrygian}, [7]int{E, F, G, A, B, C + 12, D + 12}},
		{Scale{F, Lydian}, [7]int{F, G, A, B, C + 12, D + 12, E + 12}},
		{Scale{G, Mixolydian}, [7]int{G, A, B, C + 12, D + 12, E + 12, F + 12}},
		{Scale{A, Aeolian}, [7]int{A, B, C + 12, D + 12, E + 12, F + 12, G + 12}},
		{Scale{B, Locrian}, [7]int{B, C + 12, D + 12, E + 12, F + 12, G + 12, A + 12}},
	}
	for _, tt := range tests {
		s := tt.scale
		got := [7]int{s.Tonic(), s.Second(), s.Third(), s.Fourth(), s.Fifth(), s.Sixth(), s.Seventh()}
		if got != tt.want {
			t.Errorf("%s %s degrees = %v, want %v", PitchName(s.Root), s.Mode, got, tt.want)
		}
	}
}

func TestChordNames(t *testing.T) {
	c := Scale{C, Ionian}
	tests := []struct {
		chord Chord
		name  string
		roman string
	}{
		{c.Chord(0, false), "C", "I"},
		{c.Chord(1, false), "Dm", "ii"},
		{c.Chord(4, true), "G7", "V7"},
		{c.Chord(6, false), "B°", "vii°"},
		{c.Chord(3, true), "Fmaj7", "IV7"},
	}
	for _, tt := range tests {
		if tt.chord.Name() != tt.name || tt.chord.Roman() != tt.roman {
			t.Errorf("chord = %s (%s), want %s (%s)", tt.chord.Name(), tt.chord.Roman(), tt.name, tt.roman)
		}
	}
}
