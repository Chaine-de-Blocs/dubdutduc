package music

const (
	C = iota
	Db
	D
	Eb
	E
	F
	Gb
	G
	Ab
	A
	Bb
	B
)

var (
	pitchNames      = [12]string{"C", "D♭", "D", "E♭", "E", "F", "G♭", "G", "A♭", "A", "B♭", "B"}
	sharpPitchNames = [12]string{"C", "C♯", "D", "D♯", "E", "F", "F♯", "G", "G♯", "A", "A♯", "B"}
)

func PitchName(pitch int) string {
	return pitchNames[mod(pitch, 12)]
}

type Mode int

const (
	Ionian Mode = iota
	Dorian
	Phrygian
	Lydian
	Mixolydian
	Aeolian
	Locrian
)

var modeNames = [...]string{"Ionian", "Dorian", "Phrygian", "Lydian", "Mixolydian", "Aeolian", "Locrian"}

func (m Mode) String() string {
	return modeNames[m]
}

var modeSteps = [...][7]int{
	Ionian:     {0, 2, 4, 5, 7, 9, 11},
	Dorian:     {0, 2, 3, 5, 7, 9, 10},
	Phrygian:   {0, 1, 3, 5, 7, 8, 10},
	Lydian:     {0, 2, 4, 6, 7, 9, 11},
	Mixolydian: {0, 2, 4, 5, 7, 9, 10},
	Aeolian:    {0, 2, 3, 5, 7, 8, 10},
	Locrian:    {0, 1, 3, 5, 6, 8, 10},
}

type Scale struct {
	Root int
	Mode Mode
}

// Degree returns the pitch of the 0-based scale degree d, counted upward from Root.
// Degrees past the seventh continue into the next octave.
func (s Scale) Degree(d int) int {
	return s.Root + modeSteps[s.Mode][mod(d, 7)] + 12*floorDiv(d, 7)
}

func (s Scale) Tonic() int   { return s.Degree(0) }
func (s Scale) Second() int  { return s.Degree(1) }
func (s Scale) Third() int   { return s.Degree(2) }
func (s Scale) Fourth() int  { return s.Degree(3) }
func (s Scale) Fifth() int   { return s.Degree(4) }
func (s Scale) Sixth() int   { return s.Degree(5) }
func (s Scale) Seventh() int { return s.Degree(6) }

// sharps reports whether the key signature of the parent major scale uses sharps.
func (s Scale) sharps() bool {
	switch mod(s.Root-modeSteps[Ionian][s.Mode], 12) {
	case C, G, D, A, E, B, Gb:
		return true
	}
	return false
}

// Name spells pitch with the accidentals of this key.
func (s Scale) Name(pitch int) string {
	if s.sharps() {
		return sharpPitchNames[mod(pitch, 12)]
	}
	return PitchName(pitch)
}

func (s Scale) Contains(pitch int) bool {
	pc := mod(pitch-s.Root, 12)
	for _, step := range modeSteps[s.Mode] {
		if step == pc {
			return true
		}
	}
	return false
}

// Pitches returns every scale pitch in [lo, hi], ascending.
func (s Scale) Pitches(lo, hi int) []int {
	var out []int
	for p := lo; p <= hi; p++ {
		if s.Contains(p) {
			out = append(out, p)
		}
	}
	return out
}

func mod(a, n int) int {
	return ((a % n) + n) % n
}

func floorDiv(a, n int) int {
	if a >= 0 {
		return a / n
	}
	return -((-a + n - 1) / n)
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func sign(a int) int {
	switch {
	case a > 0:
		return 1
	case a < 0:
		return -1
	}
	return 0
}
