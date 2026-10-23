package theory

// Scale is a scale definition plus how its fretboard positions are built.
type Scale struct {
	ID        string
	Name      string
	Semis     []int    // above the root
	Degrees   []string // degree names, parallel to Semis
	PerString int      // notes per string in each position
}

var MinorPent = Scale{
	ID: "minpent", Name: "minor pentatonic",
	Semis:     []int{0, 3, 5, 7, 10},
	Degrees:   []string{"1", "b3", "4", "5", "b7"},
	PerString: 2,
}

var MajorPent = Scale{
	ID: "majpent", Name: "major pentatonic",
	Semis:     []int{0, 2, 4, 7, 9},
	Degrees:   []string{"1", "2", "3", "5", "6"},
	PerString: 2,
}

var MajorScale = Scale{
	ID: "major", Name: "major scale (3 notes per string)",
	Semis:     []int{0, 2, 4, 5, 7, 9, 11},
	Degrees:   []string{"1", "2", "3", "4", "5", "6", "7"},
	PerString: 3,
}

var Scales = []Scale{MinorPent, MajorPent, MajorScale}

// Positions is the number of positions (boxes) the scale has.
func (sc Scale) Positions() int { return len(sc.Semis) }

// BoxNote is one dot of a scale position.
type BoxNote struct {
	Pos    Pos
	Degree int // index into Scale.Degrees
}

// Box builds position `pos` (0-based) of the scale in the given key: it starts
// on the low E string at scale degree `pos` and stacks PerString notes per
// string going up. The box is placed as low as possible on the neck with all
// frets ≥ 0.
func (sc Scale) Box(root PC, pos int) []BoxNote {
	n := len(sc.Semis)
	start := root.Add(sc.Semis[pos])
	f0 := FretsOf(start, 6, 0, 11)[0]
	build := func(f0 int) []BoxNote {
		var out []BoxNote
		deg := pos
		abs := OpenMidi(6) + f0
		for s := 6; s >= 1; s-- {
			for k := 0; k < sc.PerString; k++ {
				out = append(out, BoxNote{Pos{s, abs - OpenMidi(s)}, deg})
				// step to next scale note
				next := (deg + 1) % n
				step := sc.Semis[next] - sc.Semis[deg]
				if step <= 0 {
					step += 12
				}
				abs += step
				deg = next
			}
		}
		return out
	}
	box := build(f0)
	for minFret(box) < 0 {
		f0 += 12
		box = build(f0)
	}
	// A box hanging on fret 0 is fine, but if a box can sit 12 frets lower
	// and still be ≥ 0 it already does, since f0 starts at the lowest fret.
	return box
}

func minFret(b []BoxNote) int {
	m := 99
	for _, n := range b {
		m = min(m, n.Pos.F)
	}
	return m
}

func MaxFret(b []BoxNote) int {
	m := 0
	for _, n := range b {
		m = max(m, n.Pos.F)
	}
	return m
}

// Key is a major key with correct spelling of its scale.
type Key struct{ Tonic Note }

var majorSemis = []int{0, 2, 4, 5, 7, 9, 11}
var diatonicQuality = []Quality{Major, Minor, Minor, Major, Major, Minor, Dim}
var Roman = []string{"I", "ii", "iii", "IV", "V", "vi", "vii°"}

// Chord returns the diatonic triad on scale degree d (0-based).
func (k Key) Chord(d int) Triad {
	return Triad{Root: k.Tonic.Up(d, majorSemis[d]), Quality: diatonicQuality[d]}
}

// Common keys for guitar.
var Keys = []string{"C", "G", "D", "A", "E", "F", "Bb", "Eb"}

// Progressions as 0-based major-key degrees.
var Progressions = [][]int{
	{0, 4, 5, 3}, // I V vi IV
	{5, 3, 0, 4}, // vi IV I V
	{1, 4, 0, 0}, // ii V I I
	{0, 5, 3, 4}, // I vi IV V
	{0, 3, 4, 3}, // I IV V IV
	{5, 1, 4, 0}, // vi ii V I
	{0, 2, 3, 4}, // I iii IV V
}
