package music

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"math/rand/v2"
	"slices"
	"strings"
)

const (
	PPQ        = 480
	Sixteenth  = PPQ / 4
	phraseBars = 4

	chordLow    = 47
	chordHigh   = 64
	chordCenter = 55
	bassLow     = 34
	bassHigh    = 48
)

type Note struct {
	Pitch    int
	Start    int // ticks from the start of the piece
	Duration int // ticks
	Velocity int
}

type Track struct {
	Instrument Instrument
	Notes      []Note
}

type Bar struct {
	Chord   Chord
	Voicing []int
}

type Piece struct {
	Hash    string
	Scale   Scale
	BPM     int
	Beats   int // quarter-note beats per bar
	Bars    []Bar
	Melody  Track
	Harmony Track
	Bass    Track
	Drums   Track
}

func (p *Piece) BarTicks() int { return p.Beats * PPQ }

func (p *Piece) Length() int { return len(p.Bars) * p.BarTicks() }

func (p *Piece) Tonic() string { return p.Scale.Name(p.Scale.Root) }

func (p *Piece) Key() string { return p.Tonic() + " " + p.Scale.Mode.String() }

func (p *Piece) Tracks() []*Track {
	return []*Track{&p.Melody, &p.Harmony, &p.Bass, &p.Drums}
}

// Compose writes a 16-bar piece for hash. The same hash always yields the same piece.
func Compose(hash string) *Piece {
	hash = strings.ToLower(strings.TrimSpace(hash))
	r := rand.New(rand.NewChaCha8(seed(hash)))

	p := &Piece{Hash: hash}
	p.Scale = Scale{Root: r.IntN(12), Mode: choose(r, modeWeights)}
	p.Beats = choose(r, meterWeights)
	p.BPM = pick(r, tempos)
	ens := pick(r, ensembles)
	p.Melody.Instrument, p.Harmony.Instrument, p.Bass.Instrument = ens.lead, ens.chords, ens.bass
	p.Drums.Instrument = drums

	c := &composer{r: r, p: p, barUnits: p.Beats * 4, barTicks: p.BarTicks()}
	c.harmonize()
	c.melody()
	c.comp()
	c.bassline()
	c.drumline()
	return p
}

func seed(hash string) [32]byte {
	if b, err := hex.DecodeString(hash); err == nil && len(b) == 32 {
		return [32]byte(b)
	}
	return sha256.Sum256([]byte(hash))
}

type composer struct {
	r        *rand.Rand
	p        *Piece
	barUnits int
	barTicks int
}

// harmonize lays out four phrases: A A B cadence.
func (c *composer) harmonize() {
	h := harmonies[c.p.Scale.Mode]
	a := pick(c.r, h.open)
	degrees := slices.Concat(a, a, pick(c.r, h.middle), pick(c.r, h.cadence))

	var prev []int
	for i, d := range degrees {
		chord := c.p.Scale.Chord(d, false)
		if d == 4 && i+1 < len(degrees) && degrees[i+1] == 0 {
			if seventh := c.p.Scale.Chord(d, true); seventh.Dominant() {
				chord = seventh
			}
		}
		v := voice(chord.Tones, prev)
		c.p.Bars = append(c.p.Bars, Bar{Chord: chord, Voicing: v})
		prev = v
	}
}

// voice picks the close-position voicing nearest to prev inside the chord register.
func voice(tones, prev []int) []int {
	var best []int
	bestCost := math.MaxInt
	for inv := range tones {
		for bottom := chordLow; bottom <= chordHigh; bottom++ {
			if mod(bottom, 12) != tones[inv] {
				continue
			}
			v := []int{bottom}
			for k := 1; k < len(tones); k++ {
				pc := tones[(inv+k)%len(tones)]
				next := v[k-1] + 1
				for mod(next, 12) != pc {
					next++
				}
				v = append(v, next)
			}
			if v[len(v)-1] > chordHigh {
				continue
			}
			if cost := voicingCost(v, prev); cost < bestCost {
				best, bestCost = v, cost
			}
		}
	}
	return best
}

func voicingCost(v, prev []int) int {
	sum := 0
	for _, p := range v {
		sum += p
	}
	drift := abs(sum/len(v) - chordCenter)
	if prev == nil {
		return drift
	}
	return motion(v, prev) + motion(prev, v) + drift/2
}

func motion(from, to []int) int {
	total := 0
	for _, f := range from {
		nearest := math.MaxInt
		for _, t := range to {
			nearest = min(nearest, abs(f-t))
		}
		total += nearest
	}
	return total
}

func (c *composer) melody() {
	body := bodyRhythms[c.p.Beats]
	end := phraseEndRhythms[c.p.Beats]
	a := choose(c.r, body)
	b := c.other(body, a)
	x := choose(c.r, body)
	y := c.other(body, x)
	rhythms := []rhythm{
		a, b, a, choose(c.r, end),
		a, b, a, choose(c.r, end),
		x, y, x, choose(c.r, end),
		a, b, y, finalRhythms[c.p.Beats],
	}

	m := newMelodist(c.r, c.p.Scale)
	for bar, r := range rhythms {
		// The second phrase restates the first phrase over the same chords before a new ending.
		if bar >= phraseBars && bar < 2*phraseBars-1 {
			for _, n := range c.melodyNotesIn(bar - phraseBars) {
				n.Start += phraseBars * c.barTicks
				m.accept(n.Pitch)
				c.p.Melody.Notes = append(c.p.Melody.Notes, n)
			}
			continue
		}
		c.melodyBar(m, bar, r)
	}
}

func (c *composer) other(opts []weighted[rhythm], not rhythm) rhythm {
	for range 8 {
		if r := choose(c.r, opts); !slices.Equal(r, not) {
			return r
		}
	}
	return not
}

func (c *composer) melodyNotesIn(bar int) []Note {
	var out []Note
	for _, n := range c.p.Melody.Notes {
		if n.Start/c.barTicks == bar {
			out = append(out, n)
		}
	}
	return out
}

func (c *composer) melodyBar(m *melodist, bar int, r rhythm) {
	chord := c.p.Bars[bar].Chord
	final := bar == len(c.p.Bars)-1
	lastNote := -1
	for i, b := range r {
		if !b.rest {
			lastNote = i
		}
	}

	onset := 0
	for i, b := range r {
		if b.rest {
			onset += b.units
			continue
		}
		var p int
		switch {
		case final && i == lastNote:
			p = m.tonic()
		case onset == 0:
			p = m.chordTone(chord, m.r.Float64() < 0.4)
		case b.units >= 6:
			p = m.chordTone(chord, true)
		case c.strong(onset) && m.r.Float64() < 0.7:
			p = m.chordTone(chord, false)
		case i == lastNote && onset+b.units == c.barUnits && !final:
			p = m.approach(chord, c.p.Bars[bar+1].Chord)
		default:
			p = m.walk(chord)
		}
		m.accept(p)

		dur := b.units * Sixteenth
		if !final {
			dur = dur * 85 / 100
		}
		c.p.Melody.Notes = append(c.p.Melody.Notes, Note{
			Pitch:    p,
			Start:    bar*c.barTicks + onset*Sixteenth,
			Duration: dur,
			Velocity: c.velocity(82, 16, bar, onset),
		})
		onset += b.units
	}
}

func (c *composer) strong(onset int) bool {
	return c.p.Beats == 4 && onset == 8
}

func (c *composer) comp() {
	if c.p.Harmony.Instrument.ID == "chip-arp" {
		c.arp()
		return
	}
	opts := pulsedComping[c.p.Beats]
	style := choose(c.r, opts)
	last := len(c.p.Bars) - 1

	for bar, b := range c.p.Bars {
		r := style
		if bar == last {
			r = finalRhythms[c.p.Beats]
		}
		onset := 0
		for _, hit := range r {
			if !hit.rest {
				dur := hit.units * Sixteenth * 7 / 10
				if bar == last {
					dur = hit.units * Sixteenth
				}
				vel := c.velocity(58, 10, bar, onset)
				for _, p := range b.Voicing {
					c.p.Harmony.Notes = append(c.p.Harmony.Notes, Note{
						Pitch:    p,
						Start:    bar*c.barTicks + onset*Sixteenth,
						Duration: dur,
						Velocity: vel,
					})
				}
			}
			onset += hit.units
		}
	}
}

// arp plays chord tones as a short chip arpeggio that fills each bar.
func (c *composer) arp() {
	last := len(c.p.Bars) - 1
	for bar, b := range c.p.Bars {
		tones := b.Voicing
		if len(tones) == 0 {
			continue
		}
		step := 2
		if c.r.Float64() < 0.35 {
			step = 1
		}
		up := c.r.Float64() < 0.7
		n := c.barUnits / step
		if bar == last {
			c.p.Harmony.Notes = append(c.p.Harmony.Notes, Note{
				Pitch:    tones[0],
				Start:    bar * c.barTicks,
				Duration: c.barTicks,
				Velocity: c.velocity(62, 8, bar, 0),
			})
			continue
		}
		for i := 0; i < n; i++ {
			idx := i % len(tones)
			if !up {
				idx = len(tones) - 1 - idx
			}
			c.p.Harmony.Notes = append(c.p.Harmony.Notes, Note{
				Pitch:    tones[idx],
				Start:    bar*c.barTicks + i*step*Sixteenth,
				Duration: step * Sixteenth * 8 / 10,
				Velocity: c.velocity(60, 10, bar, i*step),
			})
		}
	}
}

func (c *composer) bassline() {
	pattern := choose(c.r, bassPatterns[c.p.Beats])
	last := len(c.p.Bars) - 1
	prevRoot := 0

	for bar, b := range c.p.Bars {
		root := nearestInRange(b.Chord.Root(), prevRoot, bassLow, bassHigh)
		fifth := root + mod(b.Chord.Fifth()-b.Chord.Root(), 12)
		prevRoot = root

		hits := pattern
		if bar == last {
			hits = []bassHit{{c.barUnits, bassRoot}}
		}
		onset := 0
		for _, h := range hits {
			if h.tone != bassRest {
				p := root
				if h.tone == bassFifth {
					p = fifth
				}
				dur := h.units * Sixteenth
				if bar != last {
					dur = dur * 92 / 100
				}
				c.p.Bass.Notes = append(c.p.Bass.Notes, Note{
					Pitch:    p,
					Start:    bar*c.barTicks + onset*Sixteenth,
					Duration: dur,
					Velocity: c.velocity(78, 8, bar, onset),
				})
			}
			onset += h.units
		}
	}
}

func (c *composer) drumline() {
	pattern := choose(c.r, drumPatterns[c.p.Beats])
	fill := choose(c.r, drumFills[c.p.Beats])
	last := len(c.p.Bars) - 1
	for bar := range c.p.Bars {
		hits := pattern
		switch {
		case bar == last:
			hits = drumEndings[c.p.Beats]
		case (bar+1)%phraseBars == 0:
			hits = fill
		}
		for _, h := range hits {
			c.p.Drums.Notes = append(c.p.Drums.Notes, Note{
				Pitch:    h.pitch,
				Start:    bar*c.barTicks + h.onset*Sixteenth,
				Duration: h.units * Sixteenth,
				Velocity: max(1, min(127, h.vel+c.r.IntN(5)-2)),
			})
		}
	}
}

// nearestInRange places pitch class pc in [lo, hi] as close as possible to prev (or near lo when prev is 0).
func nearestInRange(pc, prev, lo, hi int) int {
	if prev == 0 {
		prev = lo + 2
	}
	best := -1
	for p := lo; p <= hi; p++ {
		if mod(p, 12) == pc && (best < 0 || abs(p-prev) < abs(best-prev)) {
			best = p
		}
	}
	return best
}

// velocity swells toward the middle of each phrase and accents downbeats.
func (c *composer) velocity(base, depth float64, bar, onset int) int {
	pos := float64((bar%phraseBars)*c.barUnits+onset) / float64(phraseBars*c.barUnits)
	v := base + depth*math.Sin(math.Pi*pos)
	switch {
	case onset == 0:
		v += 6
	case onset%2 == 1:
		v -= 6
	}
	v += float64(c.r.IntN(5) - 2)
	return max(1, min(127, int(math.Round(v))))
}

type melodist struct {
	r       *rand.Rand
	root    int
	pitches []int // scale pitches available to the melody, ascending
	center  int   // index into pitches
	prev    int
	lastDir int
	leap    int // direction of an unresolved leap, 0 when none
	repeats int
}

func newMelodist(r *rand.Rand, s Scale) *melodist {
	base := 60 + s.Root
	if base < 62 {
		base += 12
	}
	m := &melodist{r: r, root: s.Root, pitches: s.Pitches(base-1, base+14), prev: base + 7}
	m.center = m.nearestIndex(base + 7)
	return m
}

func (m *melodist) nearestIndex(p int) int {
	i, _ := slices.BinarySearch(m.pitches, p)
	if i == len(m.pitches) {
		return i - 1
	}
	if i > 0 && p-m.pitches[i-1] < m.pitches[i]-p {
		return i - 1
	}
	return i
}

func (m *melodist) accept(p int) {
	d := p - m.prev
	m.leap = 0
	if abs(d) > 4 {
		m.leap = sign(d)
	}
	if d == 0 {
		m.repeats++
	} else {
		m.repeats = 0
		m.lastDir = sign(d)
	}
	m.prev = p
}

// chordTone returns the chord tone nearest the previous pitch, turning back after a leap.
func (m *melodist) chordTone(ch Chord, avoidRepeat bool) int {
	avoidRepeat = avoidRepeat || m.repeats > 0
	var best []int
	bestCost := math.MaxInt
	for _, p := range m.pitches {
		if !ch.Has(p) || (avoidRepeat && p == m.prev) {
			continue
		}
		cost := abs(p - m.prev)
		if m.leap != 0 && sign(p-m.prev) == m.leap {
			cost += 12
		}
		if cost < bestCost {
			best, bestCost = []int{p}, cost
		} else if cost == bestCost {
			best = append(best, p)
		}
	}
	if len(best) == 0 {
		return m.chordTone(ch, false)
	}
	return pick(m.r, best)
}

func (m *melodist) walk(ch Chord) int {
	i := m.nearestIndex(m.prev)
	last := len(m.pitches) - 1
	if m.leap != 0 {
		return m.pitches[max(0, min(last, i-m.leap))]
	}

	dir := m.direction(i)
	j := i + dir
	switch roll := m.r.Float64(); {
	case roll < 0.6:
	case roll < 0.72 && m.repeats == 0:
		j = i
	case roll < 0.9:
		j = i + 2*dir
	default:
		for k := 3; k <= 5; k++ {
			if idx := i + dir*k; idx >= 0 && idx <= last && ch.Has(m.pitches[idx]) {
				j = idx
				break
			}
		}
	}
	if j < 0 || j > last {
		j = i - (j - i)
	}
	return m.pitches[max(0, min(last, j))]
}

// direction keeps the current contour going, pulled back toward the middle of the range.
func (m *melodist) direction(i int) int {
	up := 0.5 + 0.15*float64(m.lastDir) + 0.06*float64(m.center-i)
	if m.r.Float64() < up {
		return 1
	}
	return -1
}

// approach returns a scale step next to the chord tone the next downbeat will land on.
func (m *melodist) approach(ch, next Chord) int {
	if m.leap != 0 {
		return m.walk(ch)
	}
	t := m.nearestIndex(m.chordTone(next, false))
	side := sign(m.prev - m.pitches[t])
	if side == 0 {
		side = pick(m.r, []int{-1, 1})
	}
	j := t + side
	if j < 0 || j >= len(m.pitches) || (m.pitches[j] == m.prev && m.repeats > 0) {
		j = t - side
	}
	return m.pitches[max(0, min(len(m.pitches)-1, j))]
}

func (m *melodist) tonic() int {
	best := -1
	for _, p := range m.pitches {
		if mod(p-m.root, 12) == 0 && (best < 0 || abs(p-m.prev) < abs(best-m.prev)) {
			best = p
		}
	}
	return best
}

func choose[T any](r *rand.Rand, opts []weighted[T]) T {
	total := 0
	for _, o := range opts {
		total += o.w
	}
	n := r.IntN(total)
	for _, o := range opts {
		if n < o.w {
			return o.v
		}
		n -= o.w
	}
	return opts[len(opts)-1].v
}

func pick[T any](r *rand.Rand, xs []T) T {
	return xs[r.IntN(len(xs))]
}
