package music

type weighted[T any] struct {
	v T
	w int
}

// harmony lists 4-bar progressions as 0-based scale degrees.
type harmony struct {
	open    [][]int // phrases 1 and 2, start on the tonic
	middle  [][]int // phrase 3, contrast
	cadence [][]int // phrase 4, ends on the tonic
}

var harmonies = map[Mode]harmony{
	// Degree sets avoid each mode's non-tonic diminished triad
	// (Ionian 6, Dorian 5, Phrygian 4, Mixolydian 2, Aeolian 1, Locrian 2).
	// Locrian's tonic (0) is diminished by definition.
	Ionian: {
		open: [][]int{
			{0, 4, 5, 3}, {0, 5, 3, 4}, {0, 3, 5, 4}, {0, 3, 0, 4}, {0, 2, 3, 4}, {0, 5, 1, 4},
			{0, 4, 0, 5}, {0, 1, 4, 0}, {0, 5, 4, 0}, {0, 2, 5, 3}, {0, 3, 4, 5}, {0, 4, 3, 4},
			{0, 5, 0, 4}, {0, 1, 5, 4}, {0, 2, 0, 4}, {0, 3, 1, 4},
			// jazz: I–vi–ii–V, I–ii–V–I, I–IV–ii–V, rhythm-changes colors
			{0, 5, 1, 4}, {0, 1, 4, 0}, {0, 3, 1, 4}, {0, 4, 1, 4}, {0, 2, 5, 1},
			{0, 5, 1, 0}, {0, 1, 4, 5}, {0, 4, 5, 1}, {0, 2, 1, 4}, {0, 5, 4, 1},
		},
		middle: [][]int{
			{5, 3, 0, 4}, {3, 0, 1, 4}, {1, 4, 0, 5}, {3, 4, 2, 5}, {5, 2, 3, 4},
			{4, 5, 3, 4}, {2, 5, 1, 4}, {5, 4, 3, 0}, {1, 5, 3, 4}, {3, 5, 4, 0},
			{4, 3, 5, 4}, {2, 3, 4, 5},
			// jazz bridge-ish / circle motion
			{2, 5, 1, 4}, {1, 4, 5, 1}, {5, 1, 4, 5}, {3, 1, 4, 5}, {4, 1, 5, 1},
			{1, 4, 2, 5}, {5, 1, 2, 4},
		},
		cadence: [][]int{
			{3, 1, 4, 0}, {5, 3, 4, 0}, {5, 1, 4, 0}, {3, 0, 4, 0},
			{2, 3, 4, 0}, {4, 5, 4, 0}, {1, 5, 4, 0}, {5, 4, 3, 0},
			{3, 4, 5, 0}, {2, 5, 4, 0},
			// jazz turnarounds → I
			{1, 4, 1, 0}, {5, 1, 4, 0}, {2, 5, 1, 0}, {3, 1, 5, 0}, {1, 4, 5, 0},
			{4, 1, 4, 0}, {2, 1, 4, 0},
		},
	},
	Dorian: {
		open: [][]int{
			{0, 3, 0, 3}, {0, 3, 6, 0}, {0, 1, 2, 1}, {0, 6, 3, 0}, {0, 3, 2, 6}, {0, 6, 2, 3},
			{0, 4, 3, 0}, {0, 1, 3, 0}, {0, 3, 4, 0}, {0, 6, 0, 3}, {0, 2, 3, 6}, {0, 1, 6, 3},
			{0, 3, 6, 3}, {0, 4, 6, 0}, {0, 2, 6, 3}, {0, 6, 4, 3},
			// minor/modal jazz
			{0, 1, 3, 4}, {0, 1, 6, 0}, {0, 4, 1, 0}, {0, 2, 1, 3}, {0, 6, 1, 3},
			{0, 1, 4, 3}, {0, 3, 1, 6}, {0, 4, 3, 6},
		},
		middle: [][]int{
			{2, 3, 6, 1}, {3, 6, 2, 3}, {1, 2, 3, 6}, {6, 3, 1, 4},
			{3, 4, 6, 1}, {1, 3, 4, 6}, {6, 2, 3, 4}, {4, 3, 6, 0},
			{2, 6, 3, 1}, {3, 1, 6, 4},
			{1, 4, 3, 6}, {6, 1, 3, 4}, {2, 1, 6, 3}, {4, 1, 6, 3}, {1, 6, 4, 3},
		},
		cadence: [][]int{
			{2, 3, 6, 0}, {1, 3, 6, 0}, {4, 3, 6, 0}, {3, 4, 6, 0},
			{6, 3, 4, 0}, {1, 4, 6, 0}, {2, 6, 3, 0}, {3, 6, 4, 0},
			{6, 4, 3, 0}, {1, 6, 3, 0},
			{1, 4, 3, 0}, {6, 1, 3, 0}, {4, 1, 6, 0}, {2, 1, 6, 0}, {1, 3, 4, 0},
		},
	},
	Lydian: {
		open: [][]int{
			{0, 1, 0, 1}, {0, 1, 4, 0}, {0, 1, 2, 1}, {0, 5, 1, 0}, {0, 4, 5, 1}, {0, 1, 5, 1},
			{0, 2, 1, 0}, {0, 4, 1, 0}, {0, 1, 4, 1}, {0, 5, 4, 1}, {0, 2, 5, 1}, {0, 1, 0, 4},
			{0, 4, 0, 1}, {0, 5, 2, 1}, {0, 1, 5, 4}, {0, 2, 4, 1},
			// bright / floating jazz
			{0, 1, 2, 4}, {0, 2, 1, 5}, {0, 5, 1, 4}, {0, 1, 4, 5}, {0, 4, 1, 5},
			{0, 1, 5, 0}, {0, 2, 5, 4},
		},
		middle: [][]int{
			{5, 4, 1, 1}, {2, 5, 1, 4}, {5, 1, 2, 1}, {6, 5, 1, 4},
			{4, 1, 5, 1}, {2, 1, 5, 4}, {5, 2, 4, 1}, {1, 4, 5, 1},
			{6, 1, 4, 1}, {4, 5, 2, 1},
			{1, 2, 5, 4}, {2, 4, 1, 5}, {5, 1, 4, 2}, {6, 2, 1, 4},
		},
		cadence: [][]int{
			{5, 1, 4, 0}, {2, 1, 1, 0}, {5, 2, 1, 0}, {6, 5, 1, 0},
			{4, 1, 1, 0}, {1, 5, 1, 0}, {2, 5, 1, 0}, {5, 4, 1, 0},
			{6, 1, 1, 0}, {4, 5, 1, 0},
			{1, 4, 1, 0}, {2, 1, 4, 0}, {5, 1, 1, 0}, {1, 2, 1, 0},
		},
	},
	Mixolydian: {
		open: [][]int{
			{0, 6, 3, 0}, {0, 6, 3, 3}, {0, 4, 3, 0}, {0, 3, 6, 3}, {0, 5, 6, 0}, {0, 4, 5, 3},
			{0, 6, 0, 3}, {0, 3, 0, 6}, {0, 5, 3, 6}, {0, 4, 6, 0}, {0, 1, 6, 0}, {0, 3, 5, 6},
			{0, 6, 5, 3}, {0, 1, 3, 6}, {0, 4, 0, 6}, {0, 5, 0, 6},
			// dominant / blues-jazz
			{0, 6, 1, 0}, {0, 3, 6, 0}, {0, 1, 5, 6}, {0, 4, 1, 6}, {0, 5, 1, 6},
			{0, 6, 4, 3}, {0, 1, 6, 3}, {0, 3, 1, 6},
		},
		middle: [][]int{
			{5, 4, 3, 6}, {3, 5, 6, 4}, {1, 3, 6, 3}, {3, 0, 4, 6},
			{6, 3, 5, 4}, {4, 5, 6, 3}, {1, 5, 3, 6}, {5, 6, 3, 0},
			{3, 6, 4, 5}, {6, 5, 3, 4},
			{1, 6, 3, 5}, {6, 1, 3, 0}, {4, 1, 6, 3}, {5, 1, 6, 3}, {1, 3, 5, 6},
		},
		cadence: [][]int{
			{3, 6, 6, 0}, {3, 4, 6, 0}, {5, 3, 6, 0}, {1, 3, 6, 0},
			{6, 3, 6, 0}, {4, 5, 6, 0}, {5, 6, 3, 0}, {1, 6, 6, 0},
			{3, 5, 6, 0}, {6, 4, 6, 0},
			{1, 6, 3, 0}, {4, 6, 6, 0}, {5, 1, 6, 0}, {6, 5, 6, 0}, {1, 5, 6, 0},
		},
	},
	Aeolian: {
		open: [][]int{
			{0, 5, 2, 6}, {0, 3, 6, 2}, {0, 6, 5, 6}, {0, 5, 3, 6}, {0, 3, 5, 6}, {0, 4, 5, 3},
			{0, 6, 0, 5}, {0, 2, 5, 6}, {0, 3, 0, 6}, {0, 5, 6, 0}, {0, 2, 3, 6}, {0, 6, 3, 5},
			{0, 4, 6, 5}, {0, 5, 2, 3}, {0, 3, 2, 6}, {0, 6, 2, 5},
			// minor jazz / Andalusian-adjacent
			{0, 5, 6, 2}, {0, 3, 6, 5}, {0, 2, 6, 5}, {0, 5, 3, 2}, {0, 6, 5, 2},
			{0, 4, 3, 6}, {0, 2, 5, 3}, {0, 5, 6, 3},
		},
		middle: [][]int{
			{3, 0, 5, 6}, {5, 6, 2, 5}, {3, 4, 5, 6}, {5, 3, 0, 6},
			{2, 5, 6, 3}, {6, 5, 3, 4}, {4, 5, 2, 6}, {3, 6, 5, 2},
			{5, 2, 6, 3}, {2, 3, 5, 6},
			{6, 2, 5, 3}, {5, 6, 3, 2}, {2, 6, 5, 0}, {3, 5, 2, 6}, {4, 6, 5, 2},
		},
		cadence: [][]int{
			{5, 3, 6, 0}, {3, 5, 6, 0}, {2, 3, 6, 0}, {5, 6, 3, 0},
			{6, 5, 6, 0}, {3, 2, 6, 0}, {4, 5, 6, 0}, {2, 5, 6, 0},
			{5, 2, 6, 0}, {6, 3, 5, 0},
			{6, 5, 2, 0}, {3, 6, 5, 0}, {2, 6, 5, 0}, {5, 6, 2, 0}, {6, 2, 5, 0},
		},
	},
	Phrygian: {
		// Avoid degree 4 (v°). Flat-II (1) is the characteristic color.
		open: [][]int{
			{0, 1, 0, 1}, {0, 1, 5, 0}, {0, 5, 1, 0}, {0, 2, 1, 0}, {0, 1, 2, 1},
			{0, 5, 6, 1}, {0, 1, 6, 0}, {0, 3, 1, 0}, {0, 6, 5, 1}, {0, 1, 5, 1},
			{0, 2, 5, 1}, {0, 5, 0, 1}, {0, 1, 0, 5}, {0, 6, 1, 0}, {0, 3, 5, 1},
			{0, 1, 3, 0}, {0, 2, 6, 1}, {0, 5, 2, 1},
		},
		middle: [][]int{
			{1, 5, 0, 1}, {5, 1, 2, 1}, {2, 1, 5, 0}, {6, 5, 1, 0}, {1, 2, 5, 1},
			{3, 1, 5, 0}, {5, 6, 1, 0}, {1, 0, 5, 1}, {2, 5, 1, 0}, {6, 1, 5, 0},
			{1, 5, 6, 1}, {5, 2, 1, 0},
		},
		cadence: [][]int{
			{1, 5, 1, 0}, {5, 1, 1, 0}, {2, 1, 1, 0}, {6, 5, 1, 0},
			{1, 2, 1, 0}, {5, 6, 1, 0}, {3, 1, 1, 0}, {1, 6, 1, 0},
			{2, 5, 1, 0}, {6, 1, 1, 0}, {1, 5, 6, 0}, {5, 1, 6, 0},
		},
	},
	Locrian: {
		// Tonic (0) is i°; avoid degree 2 (III°) elsewhere. Lean on bII, iv, bVI, bVII.
		open: [][]int{
			{0, 1, 0, 1}, {0, 1, 5, 0}, {0, 5, 1, 0}, {0, 3, 1, 0}, {0, 1, 3, 0},
			{0, 6, 1, 0}, {0, 1, 6, 0}, {0, 5, 6, 1}, {0, 1, 5, 1}, {0, 3, 5, 1},
			{0, 5, 0, 1}, {0, 1, 0, 5}, {0, 6, 5, 1}, {0, 3, 6, 1}, {0, 1, 6, 5},
			{0, 4, 1, 0}, {0, 1, 4, 0}, {0, 5, 3, 1},
		},
		middle: [][]int{
			{1, 5, 0, 1}, {5, 1, 3, 0}, {1, 3, 5, 0}, {6, 5, 1, 0}, {1, 5, 6, 1},
			{3, 1, 5, 0}, {5, 6, 1, 0}, {1, 0, 5, 1}, {4, 1, 5, 0}, {6, 1, 5, 0},
			{1, 6, 5, 1}, {5, 3, 1, 0},
		},
		cadence: [][]int{
			{1, 5, 1, 0}, {5, 1, 1, 0}, {3, 1, 1, 0}, {6, 5, 1, 0},
			{1, 6, 1, 0}, {5, 6, 1, 0}, {1, 3, 1, 0}, {4, 1, 1, 0},
			{6, 1, 1, 0}, {1, 5, 6, 0}, {5, 1, 6, 0}, {3, 5, 1, 0},
		},
	},
}

var modeWeights = []weighted[Mode]{
	{Ionian, 16},
	{Dorian, 16},
	{Phrygian, 14},
	{Lydian, 12},
	{Mixolydian, 16},
	{Aeolian, 16},
	{Locrian, 10},
}

var meterWeights = []weighted[int]{
	{4, 3},
	{3, 1},
}

var tempos = []int{100, 112, 124, 136, 148, 160, 172, 184, 196, 208}

// beat is one rhythmic slot measured in sixteenth notes.
type beat struct {
	units int
	rest  bool
}

type rhythm []beat

// rh builds a rhythm from sixteenth counts; negative counts are rests.
func rh(units ...int) rhythm {
	out := make(rhythm, len(units))
	for i, u := range units {
		if u < 0 {
			out[i] = beat{units: -u, rest: true}
		} else {
			out[i] = beat{units: u}
		}
	}
	return out
}

func (r rhythm) units() int {
	n := 0
	for _, b := range r {
		n += b.units
	}
	return n
}

var bodyRhythms = map[int][]weighted[rhythm]{
	4: {
		{rh(4, 4, 4, 4), 6},
		{rh(4, 2, 2, 4, 4), 5},
		{rh(6, 2, 4, 4), 6},
		{rh(6, 2, 8), 4},
		{rh(4, 4, 8), 5},
		{rh(8, 4, 4), 4},
		{rh(2, 2, 2, 2, 4, 4), 3},
		{rh(4, 2, 2, 2, 2, 4), 3},
		{rh(6, 6, 4), 4},
		{rh(4, 2, 4, 2, 4), 3},
		{rh(3, 1, 4, 4, 4), 2},
		{rh(4, 4, 2, 2, 4), 4},
		{rh(2, 2, 4, 2, 2, 4), 3},
		{rh(8, 2, 2, 4), 3},
		{rh(4, 8, 4), 3},
		{rh(2, 6, 4, 4), 3},
		{rh(4, 4, 3, 1, 4), 2},
		{rh(2, 2, 2, 2, 2, 2, 4), 2},
		{rh(6, 2, 2, 2, 4), 3},
		{rh(4, 2, 6, 4), 3},
		{rh(2, 4, 2, 4, 4), 3},
		{rh(8, 8), 2},
		{rh(3, 1, 3, 1, 4, 4), 2},
		{rh(4, 3, 1, 2, 2, 4), 2},
		{rh(2, 2, 8, 4), 3},
		{rh(5, 3, 4, 4), 2},
		{rh(4, 1, 3, 4, 4), 2},
		{rh(2, 2, 2, 6, 4), 2},
		// jazz / swing-leaning melody cells (long-short, anticipations)
		{rh(3, 1, 3, 1, 3, 1, 4), 3},
		{rh(3, 1, 4, 3, 1, 4), 3},
		{rh(2, 1, 1, 2, 1, 1, 4, 4), 2},
		{rh(6, 1, 1, 4, 4), 3},
		{rh(3, 1, 2, 2, 3, 1, 4), 2},
		{rh(4, 3, 1, 4, 3, 1), 3},
		{rh(2, 2, 3, 1, 2, 2, 4), 2},
		{rh(3, 5, 3, 5), 2},
		{rh(1, 1, 2, 4, 1, 1, 2, 4), 2},
		{rh(4, 2, 1, 1, 4, 2, 2), 2},
	},
	3: {
		{rh(4, 4, 4), 6},
		{rh(8, 4), 4},
		{rh(4, 2, 2, 4), 5},
		{rh(6, 2, 4), 5},
		{rh(2, 2, 4, 4), 3},
		{rh(4, 4, 2, 2), 3},
		{rh(4, 8), 3},
		{rh(3, 1, 4, 4), 2},
		{rh(2, 2, 2, 2, 4), 3},
		{rh(6, 4, 2), 3},
		{rh(4, 6, 2), 3},
		{rh(2, 6, 4), 2},
		{rh(3, 1, 2, 2, 4), 2},
		{rh(8, 2, 2), 2},
		{rh(2, 2, 8), 2},
		{rh(5, 3, 4), 2},
		// jazz waltz
		{rh(3, 1, 3, 1, 4), 3},
		{rh(2, 2, 3, 1, 4), 2},
		{rh(3, 1, 4, 2, 2), 2},
		{rh(6, 3, 1, 2), 2},
	},
}

var phraseEndRhythms = map[int][]weighted[rhythm]{
	4: {
		{rh(16), 3},
		{rh(12, -4), 4},
		{rh(8, 8), 3},
		{rh(4, 4, 8), 4},
		{rh(8, 4, -4), 3},
		{rh(6, 2, 8), 3},
		{rh(4, 12), 3},
		{rh(2, 2, 12), 2},
		{rh(8, 2, 2, -4), 2},
		{rh(4, 4, 4, -4), 3},
		{rh(6, 2, 4, -4), 2},
		{rh(4, 2, 2, 8), 3},
		{rh(10, 6), 2},
		{rh(3, 1, 4, 8), 2},
		// jazz phrase endings / held tones with pickup
		{rh(3, 1, 12), 2},
		{rh(4, 3, 1, 8), 2},
		{rh(2, 2, 3, 1, 8), 2},
		{rh(6, 2, 6, -2), 2},
		{rh(3, 1, 3, 1, 8), 2},
	},
	3: {
		{rh(12), 3},
		{rh(8, -4), 4},
		{rh(4, 8), 4},
		{rh(6, 2, 4), 2},
		{rh(4, 4, -4), 3},
		{rh(2, 2, 8), 2},
		{rh(8, 2, -2), 2},
		{rh(3, 1, 8), 2},
		{rh(4, 2, 2, -4), 2},
		{rh(3, 1, 4, -4), 2},
		{rh(2, 2, 3, 1, 4), 2},
	},
}

var finalRhythms = map[int]rhythm{
	4: rh(16),
	3: rh(12),
}

var sustainedComping = map[int][]weighted[rhythm]{
	4: {{rh(16), 3}, {rh(8, 8), 2}},
	3: {{rh(12), 1}},
}

var pulsedComping = map[int][]weighted[rhythm]{
	4: {
		{rh(8, 8), 3},
		{rh(6, 10), 2},
		{rh(4, 4, 4, 4), 2},
		{rh(6, 6, 4), 2},
		{rh(16), 1},
		// jazz comps: Charleston, push, broken
		{rh(3, 1, -4, 3, 1, -4), 3},
		{rh(-2, 2, 4, -2, 2, 4), 2},
		{rh(4, -4, 4, -4), 2},
		{rh(2, 2, -4, 2, 2, -4), 2},
		{rh(3, 5, 3, 5), 2},
		{rh(-4, 4, -4, 4), 2},
	},
	3: {
		{rh(-4, 4, 4), 3},
		{rh(8, 4), 2},
		{rh(12), 1},
		{rh(3, 1, 4, 4), 2},
		{rh(-2, 2, 4, 4), 2},
		{rh(4, -2, 2, 4), 2},
	},
}

const (
	bassRoot  = 0
	bassFifth = 1
	bassRest  = -1
)

type bassHit struct {
	units int
	tone  int
}

var bassPatterns = map[int][]weighted[[]bassHit]{
	4: {
		{[]bassHit{{8, bassRoot}, {8, bassFifth}}, 3},
		{[]bassHit{{6, bassRoot}, {2, bassRoot}, {8, bassFifth}}, 2},
		{[]bassHit{{4, bassRoot}, {4, bassRest}, {6, bassFifth}, {2, bassFifth}}, 1},
		// walking / jazz bass
		{[]bassHit{{4, bassRoot}, {4, bassFifth}, {4, bassRoot}, {4, bassFifth}}, 3},
		{[]bassHit{{4, bassRoot}, {4, bassRoot}, {4, bassFifth}, {4, bassFifth}}, 2},
		{[]bassHit{{2, bassRoot}, {2, bassRoot}, {4, bassFifth}, {4, bassRoot}, {4, bassFifth}}, 2},
		{[]bassHit{{4, bassRoot}, {2, bassFifth}, {2, bassFifth}, {4, bassRoot}, {4, bassFifth}}, 2},
		{[]bassHit{{6, bassRoot}, {2, bassFifth}, {4, bassFifth}, {4, bassRoot}}, 1},
		{[]bassHit{{4, bassRoot}, {4, bassRest}, {4, bassFifth}, {4, bassRest}}, 1},
	},
	3: {
		{[]bassHit{{8, bassRoot}, {4, bassFifth}}, 3},
		{[]bassHit{{12, bassRoot}}, 1},
		{[]bassHit{{4, bassRoot}, {4, bassFifth}, {4, bassRoot}}, 2},
		{[]bassHit{{4, bassRoot}, {2, bassRoot}, {2, bassFifth}, {4, bassFifth}}, 2},
		{[]bassHit{{6, bassRoot}, {2, bassFifth}, {4, bassRoot}}, 1},
	},
}

type Instrument struct {
	ID        string
	Name      string
	Program   int
	sustained bool
	Drum      bool
}

var (
	squareLead = Instrument{"square", "Square lead", 80, false, false}
	pulseLead  = Instrument{"pulse", "Pulse lead", 80, false, false}
	sawLead    = Instrument{"saw", "Saw lead", 81, false, false}
	chipChord  = Instrument{"chip-chord", "Chip chords", 80, false, false}
	chipArp    = Instrument{"chip-arp", "Chip arpeggio", 81, false, false}
	triBass    = Instrument{"tri-bass", "Triangle bass", 38, false, false}
	pulseBass  = Instrument{"pulse-bass", "Pulse bass", 39, false, false}
	drums      = Instrument{"drums", "Drums", 0, false, true}
)

type ensemble struct {
	lead, chords, bass Instrument
}

var ensembles = []ensemble{
	{squareLead, chipChord, triBass},
	{pulseLead, chipArp, pulseBass},
	{sawLead, chipChord, pulseBass},
	{squareLead, chipArp, triBass},
	{pulseLead, chipChord, pulseBass},
	{sawLead, chipArp, triBass},
}

// General MIDI percussion notes (channel 10).
const (
	drumKick      = 36
	drumSnare     = 38
	drumHatClosed = 42
	drumHatOpen   = 46
)

type drumHit struct {
	onset int // sixteenths into the bar
	pitch int
	units int
	vel   int
}

func dh(onset, pitch, units, vel int) drumHit {
	return drumHit{onset: onset, pitch: pitch, units: units, vel: vel}
}

func hats(step, last, vel int) []drumHit {
	var out []drumHit
	for o := 0; o < last; o += step {
		out = append(out, dh(o, drumHatClosed, step, vel))
	}
	return out
}

func withHats(step, last, vel int, hits ...drumHit) []drumHit {
	return append(hats(step, last, vel), hits...)
}

var drumPatterns = map[int][]weighted[[]drumHit]{
	4: {
		{withHats(2, 16, 70, dh(0, drumKick, 2, 100), dh(4, drumSnare, 2, 96), dh(8, drumKick, 2, 94), dh(12, drumSnare, 2, 98)), 4},
		{withHats(2, 16, 66, dh(0, drumKick, 2, 100), dh(4, drumSnare, 2, 94), dh(6, drumKick, 2, 78), dh(8, drumKick, 2, 92), dh(12, drumSnare, 2, 98)), 3},
		{withHats(2, 16, 64, dh(0, drumKick, 2, 100), dh(4, drumKick, 2, 88), dh(8, drumKick, 2, 96), dh(12, drumKick, 2, 88), dh(4, drumSnare, 2, 86), dh(12, drumSnare, 2, 90)), 2},
		{append(hats(4, 16, 62), dh(0, drumKick, 2, 100), dh(4, drumSnare, 2, 96), dh(8, drumKick, 2, 90), dh(10, drumKick, 2, 74), dh(12, drumSnare, 2, 98), dh(14, drumHatOpen, 2, 80)), 2},
		// jazz: ride-ish hats, snare on 2&4, sparse kick
		{append(hats(4, 16, 58), dh(0, drumKick, 2, 92), dh(4, drumSnare, 2, 88), dh(12, drumSnare, 2, 90), dh(14, drumHatOpen, 2, 70)), 3},
		{append(hats(4, 16, 54), dh(0, drumKick, 2, 90), dh(4, drumSnare, 2, 86), dh(10, drumKick, 2, 72), dh(12, drumSnare, 2, 88)), 3},
		{withHats(2, 16, 52, dh(0, drumKick, 2, 88), dh(4, drumSnare, 2, 84), dh(12, drumSnare, 2, 86), dh(6, drumHatOpen, 2, 60)), 2},
		{append(hats(4, 16, 50), dh(4, drumSnare, 2, 82), dh(12, drumSnare, 2, 84), dh(0, drumKick, 2, 86), dh(8, drumKick, 2, 70)), 2},
	},
	3: {
		{withHats(2, 12, 68, dh(0, drumKick, 2, 100), dh(4, drumSnare, 2, 90), dh(8, drumSnare, 2, 86)), 4},
		{withHats(2, 12, 66, dh(0, drumKick, 2, 100), dh(4, drumKick, 2, 78), dh(8, drumSnare, 2, 94)), 3},
		{append(hats(4, 12, 64), dh(0, drumKick, 2, 100), dh(4, drumSnare, 2, 92), dh(8, drumKick, 2, 88), dh(10, drumHatOpen, 2, 78)), 2},
		// jazz waltz
		{append(hats(4, 12, 56), dh(0, drumKick, 2, 90), dh(4, drumSnare, 2, 78), dh(8, drumHatOpen, 2, 72)), 3},
		{withHats(2, 12, 50, dh(0, drumKick, 2, 88), dh(8, drumSnare, 2, 80)), 2},
	},
}

var drumFills = map[int][]weighted[[]drumHit]{
	4: {
		{[]drumHit{dh(0, drumKick, 2, 96), dh(4, drumSnare, 2, 90), dh(8, drumKick, 2, 88), dh(10, drumSnare, 1, 80), dh(11, drumSnare, 1, 86), dh(12, drumSnare, 2, 100), dh(14, drumHatOpen, 2, 88), dh(0, drumHatClosed, 2, 60), dh(2, drumHatClosed, 2, 60), dh(4, drumHatClosed, 2, 60), dh(6, drumHatClosed, 2, 60)}, 3},
		{[]drumHit{dh(0, drumKick, 2, 96), dh(2, drumSnare, 1, 70), dh(3, drumSnare, 1, 78), dh(4, drumSnare, 2, 92), dh(6, drumKick, 2, 84), dh(8, drumSnare, 1, 80), dh(9, drumSnare, 1, 86), dh(10, drumSnare, 1, 90), dh(11, drumSnare, 1, 94), dh(12, drumKick, 2, 100), dh(14, drumHatOpen, 2, 90)}, 2},
		{withHats(2, 12, 62, dh(0, drumKick, 2, 96), dh(4, drumSnare, 2, 90), dh(8, drumKick, 2, 88), dh(12, drumSnare, 2, 100), dh(14, drumHatOpen, 2, 86)), 2},
		// jazz fills: snare chatter into the downbeat
		{[]drumHit{dh(0, drumKick, 2, 88), dh(4, drumSnare, 2, 80), dh(8, drumSnare, 1, 74), dh(9, drumSnare, 1, 78), dh(10, drumSnare, 1, 84), dh(11, drumSnare, 1, 90), dh(12, drumKick, 2, 100), dh(14, drumHatOpen, 2, 82)}, 2},
		{[]drumHit{dh(0, drumHatClosed, 2, 55), dh(4, drumSnare, 2, 78), dh(6, drumSnare, 1, 70), dh(7, drumSnare, 1, 76), dh(8, drumKick, 2, 88), dh(10, drumSnare, 1, 82), dh(11, drumSnare, 1, 88), dh(12, drumSnare, 2, 96), dh(14, drumHatOpen, 2, 80)}, 2},
	},
	3: {
		{[]drumHit{dh(0, drumKick, 2, 96), dh(2, drumSnare, 1, 72), dh(3, drumSnare, 1, 80), dh(4, drumSnare, 2, 90), dh(6, drumKick, 2, 84), dh(8, drumSnare, 1, 88), dh(9, drumSnare, 1, 94), dh(10, drumHatOpen, 2, 88)}, 3},
		{withHats(2, 10, 64, dh(0, drumKick, 2, 96), dh(4, drumSnare, 2, 90), dh(8, drumKick, 2, 92), dh(10, drumHatOpen, 2, 86)), 2},
		{[]drumHit{dh(0, drumKick, 2, 88), dh(4, drumSnare, 1, 76), dh(5, drumSnare, 1, 82), dh(6, drumSnare, 1, 88), dh(7, drumSnare, 1, 92), dh(8, drumKick, 2, 100), dh(10, drumHatOpen, 2, 80)}, 2},
	},
}

var drumEndings = map[int][]drumHit{
	4: {dh(0, drumKick, 4, 110), dh(0, drumHatOpen, 4, 90), dh(8, drumSnare, 2, 100), dh(10, drumSnare, 1, 88), dh(11, drumSnare, 1, 96), dh(12, drumKick, 4, 112)},
	3: {dh(0, drumKick, 4, 110), dh(0, drumHatOpen, 4, 88), dh(4, drumSnare, 2, 96), dh(6, drumSnare, 1, 86), dh(7, drumSnare, 1, 94), dh(8, drumKick, 4, 112)},
}
