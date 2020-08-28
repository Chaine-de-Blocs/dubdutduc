package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"gitlab.com/gomidi/midi/writer"
)

const (
	C int32 = iota
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
	Rest
)

const (
	Fb = E
	Cb = B
)

type Mode uint8

const (
	Ionian Mode = iota
	Dorian
	Phrygian
	Lydian
	Mixolydian
	Aeolian
	Locrian
)

type NoteDuration uint32

const (
	Minim       NoteDuration = 2
	CrochtetDot              = 3
	Crochtet                 = 4
	Quaver                   = 8
	Semiquaver               = 16
)

type TimeSignature struct {
	Numerator   uint8
	Denominator uint8
}

func NewTimeSignature(hash string) *TimeSignature {
	hashLen := len(hash)
	hashTrimmedLen := len(strings.TrimLeft(hash, "0"))

	leftZeros := hashLen - hashTrimmedLen

	numerator := uint8((leftZeros % 3) + 2)
	return &TimeSignature{
		Numerator:   numerator,
		Denominator: 4,
	}
}

func (ts *TimeSignature) GetTicksOfDuration(d NoteDuration, wr *writer.SMF) float64 {
	resolution := float64(wr.MetricTicks.Resolution())
	switch d {
	case Minim:
		return (float64(wr.MetricTicks.Ticks4th()) * 2) / resolution
	case CrochtetDot:
		return (float64(wr.MetricTicks.Ticks4th()) + float64(wr.MetricTicks.Ticks8th())) / resolution
	case Crochtet:
		return float64(wr.MetricTicks.Ticks4th()) / resolution
	case Quaver:
		return float64(wr.MetricTicks.Ticks8th()) / resolution
	case Semiquaver:
		return float64(wr.MetricTicks.Ticks16th()) / resolution
	}
	return 0.
}
func (ts *TimeSignature) MetricMeasureDuration() float64 {
	return float64(ts.Numerator) / float64(ts.Denominator)
}

type Note struct {
	Note     int32
	Velocity int32
	Duration NoteDuration
	Tone     int32
}

func (n *Note) Play(wr *writer.SMF) {
	if n.Note == Rest {
		wr.Silence(int8(wr.Channel()), true)
	} else {
		writer.NoteOn(wr, uint8(n.GetNoteTone()), uint8(n.Velocity))
	}

	n.ApplyMeterDuration(wr)
	writer.NoteOff(wr, uint8(n.GetNoteTone()))
}

func (n *Note) ApplyMeterDuration(wr *writer.SMF) {
	switch n.Duration {
	case CrochtetDot:
		writer.Forward(wr, 0, 3, Quaver)
	default:
		writer.Forward(wr, 0, 1, uint32(n.Duration))
	}
}

type Chord struct {
	Notes []*Note
}

func (c *Chord) Play(wr *writer.SMF) {
	for _, n := range c.Notes {
		writer.NoteOn(wr, uint8(n.GetNoteTone()), uint8(n.Velocity))
	}

	tonic := c.Notes[0]
	tonic.ApplyMeterDuration(wr)

	for _, n := range c.Notes {

		writer.NoteOff(wr, uint8(n.GetNoteTone()))
	}
}

type Melody struct {
	Notes         []*Note
	Scale         int32
	Mode          Mode
	TimeSignature *TimeSignature
	Measures      uint8
	Phrases       map[uint8][]*Note
}

func NewMelody(hash string) *Melody {
	runes := []rune{'a', 'b', 'c', 'd', 'e', 'f', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}

	type classifier struct {
		char  rune
		count int
	}

	trimmedHash := strings.TrimLeft(hash, "0")

	classifiers := make([]classifier, 0)
	for _, c := range runes {
		r := regexp.MustCompile(string(c))
		matches := r.FindAllStringIndex(trimmedHash, -1)
		classifiers = append(classifiers, classifier{
			char:  c,
			count: len(matches),
		})
	}

	sort.Slice(classifiers, func(i, j int) bool {
		return classifiers[i].count > classifiers[j].count
	})

	var scale int32
	var mode Mode

loop_scale:
	for _, c := range classifiers {
		switch c.char {
		case '4':
			scale = Ab
			break loop_scale
		case 'a':
			scale = A
			break loop_scale
		case '5':
			scale = Bb
			break loop_scale
		case 'b':
			scale = B
			break loop_scale
		case 'c':
			scale = C
			break loop_scale
		case '6':
			scale = Db
			break loop_scale
		case 'd':
			scale = D
			break loop_scale
		case '7':
			scale = Eb
			break loop_scale
		case 'e':
			scale = E
			break loop_scale
		case 'f':
			scale = F
			break loop_scale
		case '8':
			scale = Gb
			break loop_scale
		case '9':
			scale = G
			break loop_scale
		}
	}

loop_mode:
	for _, c := range classifiers {
		switch c.char {
		case '0':
			mode = Ionian
			break loop_mode
		case '1':
			mode = Dorian
			break loop_mode
		case '2':
			mode = Phrygian
			break loop_mode
		case '3':
			mode = Lydian
			break loop_mode
		case '4':
			mode = Mixolydian
			break loop_mode
		case '5':
			mode = Aeolian
			break loop_mode
		case '6':
			mode = Locrian
			break loop_mode
		}
	}

	ts := NewTimeSignature(hash)
	melody := &Melody{
		Notes:         make([]*Note, 0),
		Mode:          mode,
		Scale:         scale,
		TimeSignature: ts,
	}

	notePerPhrase := 0
	var noteDurationForPhrase NoteDuration
	var prevNote *Note
	for indexH, c := range trimmedHash {
		var note Note

		note.Velocity = 100

		note.Tone = int32(5)

		var indexOfC int
		for i, classifier := range classifiers {
			if classifier.char == c {
				indexOfC = i
				break
			}
		}

		if notePerPhrase == 0 {
		loop_duration:
			for _, dc := range trimmedHash[indexH+1:] {
				notePerPhrase++
				switch dc {
				case '0':
					noteDurationForPhrase = Minim
					break loop_duration
				case '1':
					noteDurationForPhrase = CrochtetDot
					break loop_duration
				case '2':
					noteDurationForPhrase = Crochtet
					break loop_duration
				case '3':
					fallthrough
				case '4':
					noteDurationForPhrase = Quaver
					break loop_duration
				case '5':
					noteDurationForPhrase = Semiquaver
					break loop_duration
				case '6':
					noteDurationForPhrase = Semiquaver
					break loop_duration
				}
			}
		}

		switch indexOfC {
		case 0:
			fallthrough
		case 10:
			note.Note = melody.Quinte()
		case 1:
			fallthrough
		case 7:
			fallthrough
		case 11:
			note.Note = melody.Third()
		case 2:
			fallthrough
		case 9:
			note.Note = melody.Tonic()
		case 3:
			fallthrough
		case 8:
			note.Note = melody.Seventh()
		case 4:
			note.Note = melody.Quarte()
		case 5:
			fallthrough
		case 12:
			note.Note = melody.Second()
		case 6:
			fallthrough
		case 15:
			note.Note = melody.Sixte()
		}

		if prevNote != nil && prevNote.Note != Rest {
			toneInterval := prevNote.Note - note.Note
			if toneInterval < -13 {
				note.Tone = note.Tone + 1
			} else if toneInterval > 13 {
				note.Tone = note.Tone - 1
			}
		}

		note.Duration = noteDurationForPhrase

		melody.Notes = append(melody.Notes, &note)

		prevNote = &note
		notePerPhrase--
	}

	return melody
}

func (m *Melody) Tonic() int32 {
	return m.Scale
}
func (m *Melody) Second() int32 {
	switch m.Mode {
	case Phrygian:
		fallthrough
	case Locrian:
		return m.Scale + 1
	}
	return m.Scale + 2
}
func (m *Melody) Third() int32 {
	switch m.Mode {
	case Ionian:
		fallthrough
	case Lydian:
		fallthrough
	case Mixolydian:
		return m.Scale + 4
	}
	return m.Scale + 3

}
func (m *Melody) Quarte() int32 {
	switch m.Mode {
	case Lydian:
		return m.Scale + 6
	}
	return m.Scale + 5
}
func (m *Melody) Quinte() int32 {
	switch m.Mode {
	case Locrian:
		return m.Scale + 6
	}
	return m.Scale + 7
}
func (m *Melody) Sixte() int32 {
	switch m.Mode {
	case Phrygian:
		fallthrough
	case Aeolian:
		fallthrough
	case Locrian:
		return m.Scale + 8
	}
	return m.Scale + 9
}
func (m *Melody) Seventh() int32 {
	switch m.Mode {
	case Ionian:
		fallthrough
	case Lydian:
		return m.Scale + 11
	}
	return m.Scale + 10
}

type ChordAlteration struct {
	Aug        bool
	Dim        bool
	Dom        bool
	FifthAug   bool
	Seven      bool
	SevenMaj   bool
	Ninth      bool
	NinthMin   bool
	NinthAug   bool
	Thirteenth bool
	Eleven     bool
	Sus        bool
	SusFour    bool
	Reverse    uint8
}

type Degree uint8

const (
	I = iota
	II
	III
	IV
	V
	VI
	VII
)

func (m *Melody) BuildChord(wr *writer.SMF, d Degree, alt *ChordAlteration, duration NoteDuration) {
	var chord Chord

	var f, t, q Note
	f.Tone, t.Tone, q.Tone = 3, 3, 3
	f.Duration, t.Duration, q.Duration = duration, duration, duration
	switch d {
	case I:
		f.Note = m.Tonic()
		t.Note = m.Third()
		q.Note = m.Quinte()
	case II:
		f.Note = m.Second()
		t.Note = m.Quarte()
		q.Note = m.Sixte()
	case III:
		f.Note = m.Third()
		t.Note = m.Quinte()
		q.Note = m.Seventh()
	case IV:
		f.Note = m.Quarte()
		t.Note = m.Sixte()
		q.Note = m.Tonic()
		q.Tone++
	case V:
		f.Note = m.Quinte()
		t.Note = m.Seventh()
		q.Note = m.Second()
		q.Tone++
	case VI:
		f.Note = m.Sixte()
		t.Note = m.Tonic()
		t.Tone++
		q.Note = m.Third()
		q.Tone++
	case VII:
		f.Note = m.Seventh()
		t.Note = m.Second()
		t.Tone++
		q.Note = m.Quarte()
		q.Tone++
	}

	f.Velocity, t.Velocity, q.Velocity = 100, 100, 100

	chord.Notes = append(chord.Notes, &f)
	chord.Notes = append(chord.Notes, &t)
	chord.Notes = append(chord.Notes, &q)

	chord.Play(wr)
}

func (m *Melody) BuildMelody(wr *writer.SMF) {
	m.Phrases = make(map[uint8][]*Note, 0)
	relativePosition, nextRelativePosition := 0., 0.
	for _, n := range m.Notes {
		measure, _ := m.CurrentMeasure(wr)

		relativePosition = nextRelativePosition
		nextRelativePosition += m.TimeSignature.GetTicksOfDuration(n.Duration, wr)

		fmt.Println("Measure", measure, "Metric", nextRelativePosition, "Pos", nextRelativePosition)

		// let's groove
		if nextRelativePosition > float64(m.TimeSignature.Numerator) {
			remainingTicks := float64(m.TimeSignature.Numerator) - relativePosition
			grooveRest := &Note{
				Duration: Quaver,
				Note:     Rest,
				Velocity: 100,
				Tone:     5,
			}
			switch remainingTicks {
			case 0.25:
				n.Duration = Semiquaver
			case 0.5:
				n.Duration = Quaver
			case 0.75:
				grooveRest.Duration = Quaver
				m.Phrases[measure] = append(m.Phrases[measure], grooveRest)
				grooveRest.Play(wr)

				n.Duration = Semiquaver
			case 1:
				n.Duration = Crochtet
			case 1.5:
				grooveRest.Duration = Quaver
				m.Phrases[measure] = append(m.Phrases[measure], grooveRest)
				grooveRest.Play(wr)

				n.Duration = Crochtet
			}
			nextRelativePosition = 0
		}

		if nextRelativePosition == relativePosition {
			nextRelativePosition = 0
		}

		n.Play(wr)

		m.Phrases[measure] = append(m.Phrases[measure], n)
	}
	m.Measures, _ = m.CurrentMeasure(wr)
}

func (m *Melody) Silence(wr *writer.SMF, d NoteDuration) {
	n := &Note{
		Duration: d,
		Note:     Rest,
	}
	n.Play(wr)
}

func countLinkedDuration(notes []*Note, d NoteDuration) int {
	next := 0
	for _, nextNote := range notes {
		if nextNote.Duration == d && nextNote.Note != Rest {
			next++
		} else {
			return next
		}
	}
	return next
}
func (m *Melody) BuildHarmony(wr *writer.SMF) {
	for i := uint8(1); i <= m.Measures; i++ {
		if _, ok := m.Phrases[i]; !ok {
			continue
		}

		n := m.Phrases[i][0]
		d := Degree(I)
		switch n.Note {
		case m.Second():
			d = II
		case m.Third():
			d = III
		case m.Quarte():
			d = IV
		case m.Quinte():
			d = V
		case m.Sixte():
			d = VI
		case m.Seventh():
			d = VII
		}

		nextMinim, nextQuaver, nextSemiquaver, nextCrochtet, nextCrochtetDot := 0, 0, 0, 0, 0
		resetCountersExcept := func(d NoteDuration) {
			switch d {
			case Minim:
				nextCrochtet, nextQuaver, nextSemiquaver, nextCrochtetDot = 0, 0, 0, 0
			case Crochtet:
				nextMinim, nextQuaver, nextSemiquaver, nextCrochtetDot = 0, 0, 0, 0
			case Quaver:
				nextMinim, nextCrochtet, nextSemiquaver, nextCrochtetDot = 0, 0, 0, 0
			case Semiquaver:
				nextMinim, nextQuaver, nextCrochtet, nextCrochtetDot = 0, 0, 0, 0
			case CrochtetDot:
				nextMinim, nextQuaver, nextSemiquaver, nextCrochtet = 0, 0, 0, 0
			default:
				nextMinim, nextCrochtetDot, nextCrochtet, nextQuaver, nextSemiquaver = 0, 0, 0, 0, 0
			}
		}
		for j, note := range m.Phrases[i] {
			if note.Note == Rest {
				resetCountersExcept(99) // reset all
				switch note.Duration {
				case Quaver:
					m.BuildChord(wr, d, nil, Quaver)
				case Crochtet:
					fallthrough
				default:
					note.Play(wr)
				}
				continue
			}

			switch note.Duration {
			case Minim:
				// m.BuildChord(wr, d, nil, Minim)
				// resetCountersExcept(Minim)
				// continue
				if nextMinim > 0 {
					continue
				}
				nextMinim = countLinkedDuration(m.Phrases[i][j:], Minim)
				switch nextMinim {
				case 1:
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, Crochtet)
					m.Silence(wr, Quaver)
				case 2:
					m.BuildChord(wr, d, nil, Minim)
					m.Silence(wr, Crochtet)
					m.BuildChord(wr, d, nil, Crochtet)
				case 3: // out of measures but why not it's music yo, let happen what happens
					// in the world of happy happenings
					m.BuildChord(wr, d, nil, Minim)
					m.Silence(wr, Minim)
					m.BuildChord(wr, d, nil, Crochtet)
					m.BuildChord(wr, d, nil, Crochtet)
				}
			case CrochtetDot:
				// m.BuildChord(wr, d, nil, CrochtetDot)
				// resetCountersExcept(CrochtetDot)
				// continue
				if nextCrochtetDot > 0 {
					continue
				}
				nextCrochtetDot = countLinkedDuration(m.Phrases[i][j:], CrochtetDot)
				switch nextCrochtetDot {
				case 2:
					m.Silence(wr, CrochtetDot)
					m.BuildChord(wr, d, nil, Crochtet)
					m.Silence(wr, Quaver)
				case 3:
					m.BuildChord(wr, d, nil, CrochtetDot)
					m.BuildChord(wr, d, nil, Quaver)
					m.Silence(wr, Crochtet)
					m.BuildChord(wr, d, nil, CrochtetDot)
				case 4:
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, CrochtetDot)
					m.Silence(wr, CrochtetDot)
					m.BuildChord(wr, d, nil, Crochtet)
				case 5:
					m.BuildChord(wr, d, nil, CrochtetDot)
					m.BuildChord(wr, d, nil, CrochtetDot)
					m.Silence(wr, CrochtetDot)
					m.BuildChord(wr, d, nil, CrochtetDot)
					m.Silence(wr, CrochtetDot)
				default:
					for nc := 1; nc <= nextCrochtetDot; nc++ {
						m.BuildChord(wr, d, nil, CrochtetDot)
					}
				}
			case Crochtet:
				// m.BuildChord(wr, d, nil, Crochtet)
				// resetCountersExcept(Crochtet)
				// continue
				if nextCrochtet > 0 {
					continue
				}
				nextCrochtet = countLinkedDuration(m.Phrases[i][j:], Crochtet)
				switch nextCrochtet {
				case 2: // Minim
					m.BuildChord(wr, d, nil, Minim)
				case 3:
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, Crochtet)
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, Crochtet)
				case 4: // 2 Minim
					m.BuildChord(wr, d, nil, Minim)
					m.Silence(wr, Crochtet)
					m.BuildChord(wr, d, nil, Crochtet)
				case 5: // 5 Crochet
					m.BuildChord(wr, d, nil, Minim)
					m.Silence(wr, Crochtet)
					m.BuildChord(wr, d, nil, Crochtet)
					m.Silence(wr, Crochtet)
				default:
					m.BuildChord(wr, d, nil, Crochtet)
				}
			case Quaver:
				// m.BuildChord(wr, d, nil, Quaver)
				// resetCountersExcept(Quaver)
				// continue
				if nextQuaver > 0 {
					continue
				}
				nextQuaver = countLinkedDuration(m.Phrases[i][j:], Quaver)
				switch nextQuaver {
				case 1:
					m.Silence(wr, Quaver)
				case 2: // 1 Crochtet
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, Quaver)
				case 3:
					m.BuildChord(wr, d, nil, CrochtetDot)
				case 4: // 1 Minim
					m.Silence(wr, Crochtet)
					m.BuildChord(wr, d, nil, CrochtetDot)
				case 5:
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, Minim)
				case 6: // 3 Crochtet
					m.BuildChord(wr, d, nil, Crochtet)
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, Crochtet)
					m.Silence(wr, Quaver)
				case 7: // 2 Minim (or 4 Crochtet)
					m.BuildChord(wr, d, nil, Minim)
					m.Silence(wr, Crochtet)
					m.BuildChord(wr, d, nil, Crochtet)
				case 8:
					m.BuildChord(wr, d, nil, Minim)
					m.Silence(wr, Crochtet)
					m.BuildChord(wr, d, nil, CrochtetDot)
				case 9: // 2 Minim + 1 Crochtet
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, Crochtet)
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, Crochtet)
					m.BuildChord(wr, d, nil, Minim)
				case 10:
					m.BuildChord(wr, d, nil, Minim)
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, Crochtet)
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, Crochtet)
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, Quaver)
				}
			case Semiquaver:
				// m.BuildChord(wr, d, nil, Semiquaver)
				// resetCountersExcept(Semiquaver)
				// continue
				if nextSemiquaver > 0 {
					continue
				}
				nextSemiquaver = countLinkedDuration(m.Phrases[i][j:], Semiquaver)
				switch nextSemiquaver {
				case 1:
					m.Silence(wr, Semiquaver)
				case 2: // 1 Quaver
					m.Silence(wr, Quaver)
				case 3:
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, Semiquaver)
				case 4: // 1 Crochtet
					m.BuildChord(wr, d, nil, Crochtet)
				case 5:
					m.Silence(wr, Semiquaver)
					m.BuildChord(wr, d, nil, Crochtet)
				case 6: // 1 Crochtet + 1 Quaver
					m.Silence(wr, Crochtet)
					m.BuildChord(wr, d, nil, Quaver)
				case 7:
					m.Silence(wr, Crochtet)
					m.Silence(wr, Semiquaver)
					m.BuildChord(wr, d, nil, Quaver)
				case 8: // 2 Crochtet (ou 1 Minim)
					m.BuildChord(wr, d, nil, Crochtet)
					m.BuildChord(wr, d, nil, Crochtet)
				case 9:
					m.BuildChord(wr, d, nil, Crochtet)
					m.Silence(wr, Semiquaver)
					m.BuildChord(wr, d, nil, Crochtet)
				case 10: // 1 Minim + 1 Quaver
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, Minim)
				case 11:
					m.BuildChord(wr, d, nil, Minim)
					m.Silence(wr, Semiquaver)
					m.BuildChord(wr, d, nil, Quaver)
				case 12: // 1 Minim + 1 Crochtet
					m.BuildChord(wr, d, nil, Minim)
					m.Silence(wr, Crochtet)
				case 13:
					m.BuildChord(wr, d, nil, Minim)
					m.Silence(wr, Semiquaver)
					m.BuildChord(wr, d, nil, Crochtet)
				case 14: // 1 Minim + 1 Crochtet + 1 Quaver
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, Quaver)
					m.BuildChord(wr, d, nil, Quaver)
					m.BuildChord(wr, d, nil, Minim)
				case 15:
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, Quaver)
					m.Silence(wr, Semiquaver)
					m.BuildChord(wr, d, nil, Quaver)
					m.BuildChord(wr, d, nil, Minim)
				case 16: // 2 Minim (or 4 Crochtet)
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, Minim)
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, Crochtet)
				case 17:
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, Minim)
					m.Silence(wr, Quaver)
					m.Silence(wr, Semiquaver)
					m.BuildChord(wr, d, nil, Crochtet)
				case 18: // 2 Minim + 1 Quaver
					m.BuildChord(wr, d, nil, Minim)
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, Crochtet)
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, Quaver)
				case 19:
					m.Silence(wr, Semiquaver)
					m.BuildChord(wr, d, nil, Minim)
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, Crochtet)
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, Quaver)
				case 20: // 2 Minim + 1 Crochtet
					m.BuildChord(wr, d, nil, Minim)
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, Crochtet)
					m.Silence(wr, Quaver)
					m.BuildChord(wr, d, nil, Crochtet)
				}
			}
			resetCountersExcept(note.Duration)
		}
	}
}
func (m *Melody) CurrentMeasure(wr *writer.SMF) (uint8, uint64) {
	tickPosition := wr.Position() / uint64(wr.MetricTicks.Resolution())
	if tickPosition < uint64(m.TimeSignature.Numerator) {
		return 1, tickPosition
	}
	return uint8(tickPosition/uint64(m.TimeSignature.Numerator)) + 1, tickPosition
}

func (n *Note) GetNoteTone() int32 {
	return n.Note + (12 * n.Tone)
}

func main() {
	err := writer.WriteSMF("./t.mid", 2, func(wr *writer.SMF) error {
		//hash := "0ccccccccccc0" // 3/4 Minim
		//hash := "0cccccccccc1" // 3/4 CrochtetDot
		//hash := "0cccccccccc2" // 3/4 Crochtet
		//hash := "0cccccccccccccc5" // 3/4 Semiquaver
		//hash := "00ccccccccccc0" // 4/4 Minim
		//hash := "00ccccccccccc1" // 4/4 CrochtetDot
		//hash := "00ccccccccccc2" // 4/4 Crochtet
		//hash := "00ccccccccccc3" // 4/4 Quaver
		//hash := "00cccccccccccccccc3" // 4/4 Semiquaver
		//hash := "000ccccccccccc0" // 5/4 Minim
		//hash := "000cccccccccccccc1" // 5/4 CrochtetDot
		//hash := "000cccccccccccccc2" // 5/4 Crochtet
		//hash := "000ccccccccccccccccccc3" // 5/4 Quaver
		//hash := "000cccccccccccccccccccccccc5" // 5/4 Semiquaver

		hashes := []string{
			"00000000000000000003efccdd987dd6d93ba18327eef8fd4b46d0de863eb14c",
			"000000000000000000051f8864b8eddf483e7d2b941d626ecea1de70fa0bf551",
			"0000000000000000000e760a04fc958a0631d47490b5f111d0d6aca418b9df17",
			"00000000000000000011f9866ca32fbbbb3cfba26af498dcd98c0f013a920021",
			"00000000000000000013f43456fe2e94a0760eaf779912e0fa37dfb64fe4ccdc",
		}

		wr.SetChannel(1) // sets the channel for the next messages
		writer.TempoBPM(wr, 120)
		writer.TrackSequenceName(wr, "title")
		writer.Instrument(wr, "Lead")

		melodies := make([]*Melody, 0)

		for _, h := range hashes {
			melodies = append(melodies, NewMelody(h))
		}

		for _, m := range melodies {
			writer.Meter(wr, m.TimeSignature.Numerator, m.TimeSignature.Denominator)
			m.BuildMelody(wr)
		}
		writer.EndOfTrack(wr)

		wr.SetChannel(2)
		for _, m := range melodies {
			m.BuildHarmony(wr)
		}
		writer.EndOfTrack(wr)

		// wr.SetChannel(3)
		// writer.Instrument(wr, "Bass")
		// writer.EndOfTrack(wr)

		// wr.SetChannel(4)
		// writer.Instrument(wr, "Percussions")
		// writer.EndOfTrack(wr)

		return nil
	})

	if err != nil {
		fmt.Printf("could not write file, error: %s", err)
		return
	}
}
