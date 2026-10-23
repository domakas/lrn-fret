package drill

import (
	"fmt"
	"math/rand/v2"

	"github.com/domakas/lrn-fret/internal/theory"
)

// curriculum string order: low E and A first, then up the neck.
var stringOrder = []int{6, 5, 4, 3, 2, 1}

// Notes: a spot is marked, name the note.
type Notes struct{}

func (Notes) ID() string   { return "notes" }
func (Notes) Name() string { return "Note names" }
func (Notes) Desc() string { return "A spot is marked: type its note (c, f#, bb)" }
func (Notes) FastMs() int  { return 3000 }

func (Notes) Cards() []string {
	var out []string
	for _, naturals := range []bool{true, false} {
		for _, s := range stringOrder {
			for f := 0; f <= 12; f++ {
				if theory.PitchAt(Pos{S: s, F: f}).Natural() == naturals {
					out = append(out, fmt.Sprintf("note:s%df%d", s, f))
				}
			}
		}
	}
	return out
}

func (n Notes) Make(card string, _ *rand.Rand) *Question {
	var p Pos
	fmt.Sscanf(card, "note:s%df%d", &p.S, &p.F)
	pc := theory.PitchAt(p)
	return &Question{
		Card:           card,
		Prompt:         fmt.Sprintf("What note is this? (%s string, fret %d)", theory.StringLabel(p.S), p.F),
		Kind:           AnswerText,
		Marks:          map[Pos]Mark{p: {MarkTarget, "?"}},
		FretMax:        12,
		CheckText:      checkNote(pc),
		Answer:         pc.Both(),
		Expect:         pc.Sharp(),
		Solution:       []Pos{p},
		SolutionLabels: map[Pos]string{p: spellPC(pc).String()},
		FastMs:         n.FastMs(),
	}
}

// Find: given a note and a string, find it. The cursor starts at the fret of
// your last pick, like your hand staying where it was on the neck.
type Find struct {
	lastFret int
}

func (Find) ID() string   { return "find" }
func (Find) Name() string { return "Find the note" }
func (Find) Desc() string { return "Find a note on a string: move the cursor or type the fret" }
func (Find) FastMs() int  { return 4000 }

func (Find) Cards() []string {
	var out []string
	for _, naturals := range []bool{true, false} {
		for _, s := range stringOrder {
			for pc := theory.PC(0); pc < 12; pc++ {
				if pc.Natural() == naturals {
					out = append(out, fmt.Sprintf("find:%d:s%d", pc, s))
				}
			}
		}
	}
	return out
}

func (fd *Find) Make(card string, rng *rand.Rand) *Question {
	var pcInt, s int
	fmt.Sscanf(card, "find:%d:s%d", &pcInt, &s)
	pc := theory.PC(pcInt)
	name := pc.Sharp()
	if rng.IntN(2) == 0 {
		name = pc.Flat()
	}
	var sol []Pos
	for _, f := range theory.FretsOf(pc, s, 0, 12) {
		sol = append(sol, Pos{S: s, F: f})
	}
	return &Question{
		Card:    card,
		Prompt:  fmt.Sprintf("Find %s on the %s string", name, theory.StringLabel(s)),
		Kind:    AnswerPos,
		Need:    1,
		String:  s,
		Cursor:  Pos{S: s, F: fd.lastFret},
		Marks:   map[Pos]Mark{},
		FretMax: 12,
		CheckPos: func(sel []Pos) bool {
			return len(sel) == 1 && sel[0].S == s && sel[0].F <= 12 && theory.PitchAt(sel[0]) == pc
		},
		Answer:         pc.Both(),
		Solution:       sol,
		SolutionLabels: labelsFor(sol, func(Pos) string { return name }),
		FastMs:         fd.FastMs(),
		OnResult: func(sel []Pos, _ bool) {
			if len(sel) > 0 { // nothing picked on reveal: hand stays put
				fd.lastFret = sel[len(sel)-1].F
			}
		},
	}
}
