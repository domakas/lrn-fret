package theory

import "fmt"

// CircleMajors is the circle of fifths, clockwise from C at the top.
var CircleMajors = []string{"C", "G", "D", "A", "E", "B", "F#", "Db", "Ab", "Eb", "Bb", "F"}

// CircleMinors are the relative minors, parallel to CircleMajors.
var CircleMinors = []string{"A", "E", "B", "F#", "C#", "G#", "D#", "Bb", "F", "C", "G", "D"}

// CircleIndex returns the position of a major key's pitch class on the circle.
func CircleIndex(pc PC) int {
	for i, k := range CircleMajors {
		if MustNote(k).PC == pc {
			return i
		}
	}
	return -1
}

// Signature returns the key signature of the major key at circle index i,
// e.g. "2#", "3b" or "0".
func Signature(i int) string {
	switch {
	case i == 0:
		return "0"
	case i <= 6:
		return fmt.Sprintf("%d#", i)
	}
	return fmt.Sprintf("%db", 12-i)
}

// RelativeMinor returns the relative minor root of a major key (a minor 3rd
// below).
func RelativeMinor(n Note) Note { return n.Up(5, 9) }
