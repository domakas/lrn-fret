package theory

import "fmt"

type Quality int

const (
	Major Quality = iota
	Minor
	Dim
)

func (q Quality) Suffix() string { return [...]string{"", "m", "dim"}[q] }
func (q Quality) Name() string   { return [...]string{"major", "minor", "diminished"}[q] }
func (q Quality) Key() string    { return [...]string{"maj", "min", "dim"}[q] }

// third and fifth sizes in semitones
func (q Quality) semis() (int, int) {
	switch q {
	case Minor:
		return 3, 7
	case Dim:
		return 3, 6
	}
	return 4, 7
}

// Triad is a spelled triad.
type Triad struct {
	Root    Note
	Quality Quality
}

func (t Triad) Name() string { return t.Root.String() + t.Quality.Suffix() }

// Tones returns root, third, fifth spelled correctly.
func (t Triad) Tones() [3]Note {
	th, fi := t.Quality.semis()
	return [3]Note{t.Root, t.Root.Up(2, th), t.Root.Up(4, fi)}
}

// Tone returns the chord tone for degree 1, 3 or 5.
func (t Triad) Tone(deg int) Note {
	return t.Tones()[map[int]int{1: 0, 3: 1, 5: 2}[deg]]
}

// StringSet is three adjacent strings, named by the lowest-numbered (highest
// pitched) one: set 1 = strings 3,2,1 (G B e) ... set 4 = strings 6,5,4.
type StringSet int

// Strings returns the strings from lowest pitch to highest.
func (ss StringSet) Strings() [3]int {
	h := int(ss)
	return [3]int{h + 2, h + 1, h}
}

func (ss StringSet) String() string {
	s := ss.Strings()
	return fmt.Sprintf("%s %s %s", StringNames[s[0]], StringNames[s[1]], StringNames[s[2]])
}

// Inversion: 0 root position (1 in bass), 1 first (3 in bass), 2 second (5 in bass).
func InversionName(inv int) string {
	return [...]string{"root position", "1st inversion", "2nd inversion"}[inv]
}

// InversionOrder gives the chord degrees from low string to high string.
func InversionOrder(inv int) [3]int {
	return [...][3]int{{1, 3, 5}, {3, 5, 1}, {5, 1, 3}}[inv]
}

// Voicings returns every close voicing of t with the given inversion on the
// string set, with all frets in [0, maxFret] and a playable span (≤ 4 frets).
// Each voicing lists positions from low string to high string.
func Voicings(t Triad, ss StringSet, inv, maxFret int) [][3]Pos {
	strs := ss.Strings()
	order := InversionOrder(inv)
	var out [][3]Pos
	for _, f0 := range FretsOf(t.Tone(order[0]).PC, strs[0], 0, maxFret) {
		v := [3]Pos{{strs[0], f0}}
		ok := true
		prev := Midi(v[0])
		for i := 1; i < 3; i++ {
			// next chord tone, the first one above the previous note
			pc := t.Tone(order[i]).PC
			m := prev + 1
			for PC(m).Norm() != pc {
				m++
			}
			f := m - OpenMidi(strs[i])
			if f < 0 || f > maxFret {
				ok = false
				break
			}
			v[i] = Pos{strs[i], f}
			prev = m
		}
		if ok && span(v[:]) <= 4 {
			out = append(out, v)
		}
	}
	return out
}

func span(ps []Pos) int {
	lo, hi := 99, -1
	for _, p := range ps {
		lo, hi = min(lo, p.F), max(hi, p.F)
	}
	return hi - lo
}

// MatchesVoicing reports whether sel (any order) is a close voicing of t in
// the given inversion on the string set.
func MatchesVoicing(sel []Pos, t Triad, ss StringSet, inv int) bool {
	if len(sel) != 3 {
		return false
	}
	byString := map[int]Pos{}
	for _, p := range sel {
		byString[p.S] = p
	}
	strs := ss.Strings()
	order := InversionOrder(inv)
	prev := -1
	for i, s := range strs {
		p, ok := byString[s]
		if !ok || PitchAt(p) != t.Tone(order[i]).PC {
			return false
		}
		if Midi(p) <= prev {
			return false
		}
		prev = Midi(p)
	}
	return span(sel) <= 4
}

// Common chord roots with their usual spellings.
var Roots = []string{"C", "G", "D", "A", "E", "F", "Bb", "Eb", "B", "F#", "Ab", "Db", "C#", "Gb"}
