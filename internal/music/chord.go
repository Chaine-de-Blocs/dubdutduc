package music

import "strings"

type Quality int

const (
	Major Quality = iota
	Minor
	Diminished
	Augmented
)

type Chord struct {
	Degree  int
	Tones   []int // pitch classes, root first
	Quality Quality
	Seventh bool
	scale   Scale
}

// Chord builds the diatonic triad (or seventh chord) stacked in thirds on degree.
func (s Scale) Chord(degree int, seventh bool) Chord {
	n := 3
	if seventh {
		n = 4
	}
	tones := make([]int, n)
	for i := range tones {
		tones[i] = mod(s.Degree(degree+2*i), 12)
	}
	third, fifth := mod(tones[1]-tones[0], 12), mod(tones[2]-tones[0], 12)
	q := Major
	switch {
	case third == 3 && fifth == 7:
		q = Minor
	case third == 3 && fifth == 6:
		q = Diminished
	case third == 4 && fifth == 8:
		q = Augmented
	}
	return Chord{Degree: mod(degree, 7), Tones: tones, Quality: q, Seventh: seventh, scale: s}
}

func (c Chord) Root() int { return c.Tones[0] }

func (c Chord) Fifth() int { return c.Tones[2] }

func (c Chord) Has(pitch int) bool {
	pc := mod(pitch, 12)
	for _, t := range c.Tones {
		if t == pc {
			return true
		}
	}
	return false
}

func (c Chord) Dominant() bool {
	return c.Quality == Major && len(c.Tones) == 4 && mod(c.Tones[3]-c.Tones[0], 12) == 10
}

func (c Chord) Name() string {
	name := c.scale.Name(c.Root())
	switch c.Quality {
	case Minor:
		name += "m"
	case Diminished:
		name += "°"
	case Augmented:
		name += "+"
	}
	if c.Seventh {
		if c.Quality == Major && !c.Dominant() {
			name += "maj"
		}
		name += "7"
	}
	return name
}

var numerals = [7]string{"I", "II", "III", "IV", "V", "VI", "VII"}

func (c Chord) Roman() string {
	n := numerals[c.Degree]
	switch c.Quality {
	case Minor:
		n = strings.ToLower(n)
	case Diminished:
		n = strings.ToLower(n) + "°"
	case Augmented:
		n += "+"
	}
	if c.Seventh {
		n += "7"
	}
	return n
}
