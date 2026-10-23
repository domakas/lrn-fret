package theory

import "fmt"

// Pos is a spot on the fretboard. S is the string number as guitarists count
// it: 1 = high e, 6 = low E. F is the fret, 0 = open.
type Pos struct {
	S, F int
}

func (p Pos) String() string { return fmt.Sprintf("string %d fret %d", p.S, p.F) }

// openMidi is standard tuning, indexed by string number (index 0 unused).
var openMidi = [7]int{0, 64, 59, 55, 50, 45, 40}

var StringNames = [7]string{"", "e", "B", "G", "D", "A", "E"}

// StringLabel is a name for prompts: "low E", "A", ..., "high e".
func StringLabel(s int) string {
	switch s {
	case 1:
		return "high e"
	case 6:
		return "low E"
	}
	return StringNames[s]
}

// Midi returns the absolute pitch at p.
func Midi(p Pos) int { return openMidi[p.S] + p.F }

// OpenMidi returns the absolute pitch of an open string.
func OpenMidi(s int) int { return openMidi[s] }

// PitchAt returns the pitch class at p.
func PitchAt(p Pos) PC { return PC(Midi(p)).Norm() }

// FretsOf returns every fret in [lo, hi] on string s that sounds pc.
func FretsOf(pc PC, s, lo, hi int) []int {
	var out []int
	for f := lo; f <= hi; f++ {
		if PitchAt(Pos{s, f}) == pc {
			out = append(out, f)
		}
	}
	return out
}

// Find returns every position in the fret window [lo, hi] (all strings) that
// sounds pc.
func Find(pc PC, lo, hi int) []Pos {
	var out []Pos
	for s := 1; s <= 6; s++ {
		for _, f := range FretsOf(pc, s, lo, hi) {
			out = append(out, Pos{s, f})
		}
	}
	return out
}
