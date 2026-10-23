// Package theory holds the music theory used by the drills: pitch classes,
// spelling, the fretboard, triads and scale shapes. It has no UI code.
package theory

import (
	"fmt"
	"strings"
)

// PC is a pitch class, 0 = C ... 11 = B.
type PC int

func (p PC) Norm() PC { return PC(((int(p) % 12) + 12) % 12) }

func (p PC) Add(semis int) PC { return PC(int(p) + semis).Norm() }

var sharpNames = [12]string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
var flatNames = [12]string{"C", "Db", "D", "Eb", "E", "F", "Gb", "G", "Ab", "A", "Bb", "B"}

// Sharp returns the sharp spelling, e.g. "C#".
func (p PC) Sharp() string { return sharpNames[p.Norm()] }

// Flat returns the flat spelling, e.g. "Db".
func (p PC) Flat() string { return flatNames[p.Norm()] }

// Both returns "C#/Db" for accidentals and "C" for naturals.
func (p PC) Both() string {
	if p.Natural() {
		return p.Sharp()
	}
	return p.Sharp() + "/" + p.Flat()
}

// Natural reports whether the pitch class is a white key.
func (p PC) Natural() bool { return len(sharpNames[p.Norm()]) == 1 }

var letterPC = map[byte]PC{'C': 0, 'D': 2, 'E': 4, 'F': 5, 'G': 7, 'A': 9, 'B': 11}

const letters = "CDEFGAB"

// ParseNote parses names like "c", "F#", "bb", "Ebb", "c##". Case-insensitive.
func ParseNote(s string) (PC, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty note")
	}
	base, ok := letterPC[strings.ToUpper(s[:1])[0]]
	if !ok {
		return 0, fmt.Errorf("bad note %q", s)
	}
	shift := 0
	for _, r := range strings.ToLower(s[1:]) {
		switch r {
		case '#', 's':
			shift++
		case 'b':
			shift--
		default:
			return 0, fmt.Errorf("bad note %q", s)
		}
	}
	return base.Add(shift), nil
}

// Note is a spelled note: a letter plus a pitch class.
type Note struct {
	Letter byte // 'A'..'G'
	PC     PC
}

func (n Note) String() string {
	acc := int(n.PC) - int(letterPC[n.Letter])
	acc = ((acc%12)+12+6)%12 - 6 // range -6..5
	switch {
	case acc > 0:
		return string(n.Letter) + strings.Repeat("#", acc)
	case acc < 0:
		return string(n.Letter) + strings.Repeat("b", -acc)
	}
	return string(n.Letter)
}

// MustNote parses a spelled note name, panicking on error. For constants.
func MustNote(s string) Note {
	pc, err := ParseNote(s)
	if err != nil {
		panic(err)
	}
	return Note{Letter: strings.ToUpper(s[:1])[0], PC: pc}
}

// Up returns the note `steps` letters and `semis` semitones above n,
// spelled correctly (e.g. Up(Eb, 2, 3) = Gb, the minor third).
func (n Note) Up(steps, semis int) Note {
	i := strings.IndexByte(letters, n.Letter)
	return Note{Letter: letters[(i+steps)%7], PC: n.PC.Add(semis)}
}

// Interval is a named distance above a root.
type Interval struct {
	Name  string // degree style: "b3", "5"
	Long  string // "minor 3rd"
	Semis int
	Steps int // letter steps, for spelling
}

var Intervals = []Interval{
	{"1", "root", 0, 0},
	{"b2", "minor 2nd", 1, 1},
	{"2", "major 2nd", 2, 1},
	{"b3", "minor 3rd", 3, 2},
	{"3", "major 3rd", 4, 2},
	{"4", "perfect 4th", 5, 3},
	{"b5", "flat 5th", 6, 4},
	{"5", "perfect 5th", 7, 4},
	{"b6", "minor 6th", 8, 5},
	{"6", "major 6th", 9, 5},
	{"b7", "minor 7th", 10, 6},
	{"7", "major 7th", 11, 6},
	{"8", "octave", 12, 7},
}

func IntervalByName(name string) Interval {
	for _, iv := range Intervals {
		if iv.Name == name {
			return iv
		}
	}
	panic("unknown interval " + name)
}
