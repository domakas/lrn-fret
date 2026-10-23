package drill

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"

	"github.com/domakas/lrn-fret/internal/theory"
)

// Scales: a scale position is shown with roots marked; name the degree of ?.
type Scales struct{}

func (Scales) ID() string   { return "scales" }
func (Scales) Name() string { return "Scale degrees" }
func (Scales) Desc() string { return "Pentatonic & major positions: which degree is the ?" }
func (Scales) FastMs() int  { return 4000 }

func (Scales) Cards() []string {
	var out []string
	for _, sc := range theory.Scales {
		for p := 0; p < sc.Positions(); p++ {
			for d := 1; d < len(sc.Degrees); d++ { // roots are shown, so skip 1
				out = append(out, fmt.Sprintf("scale:%s:%d:%d", sc.ID, p, d))
			}
		}
	}
	return out
}

func scaleByID(id string) theory.Scale {
	for _, sc := range theory.Scales {
		if sc.ID == id {
			return sc
		}
	}
	panic("bad scale " + id)
}

func (m Scales) Make(card string, rng *rand.Rand) *Question {
	parts := strings.Split(card, ":")
	sc := scaleByID(parts[1])
	var pos, deg int
	fmt.Sscanf(parts[2], "%d", &pos)
	fmt.Sscanf(parts[3], "%d", &deg)

	key := theory.MustNote(pick(rng, theory.Roots))
	box := sc.Box(key.PC, pos)
	var candidates []Pos
	marks := map[Pos]Mark{}
	var all []Pos
	for _, n := range box {
		all = append(all, n.Pos)
		if n.Degree == 0 {
			marks[n.Pos] = Mark{MarkRoot, "R"}
		} else {
			marks[n.Pos] = Mark{MarkDot, "●"}
		}
		if n.Degree == deg {
			candidates = append(candidates, n.Pos)
		}
	}
	target := pick(rng, candidates)
	marks[target] = Mark{MarkTarget, "?"}
	want := sc.Degrees[deg]
	degOf := func(p Pos) string {
		for _, n := range box {
			if n.Pos == p {
				return sc.Degrees[n.Degree]
			}
		}
		return ""
	}
	return &Question{
		Card: card,
		Prompt: fmt.Sprintf("%s %s, position %d. Which degree is the ? (degrees: %s)",
			key, sc.Name, pos+1, strings.Join(sc.Degrees, " ")),
		Kind:    AnswerText,
		Marks:   marks,
		FretMax: max(12, theory.MaxFret(box)),
		CheckText: func(in string) (bool, bool) {
			in = normalize(in)
			if in == "r" || in == "root" {
				in = "1"
			}
			if slices.Contains(sc.Degrees, in) {
				return in == want, true
			}
			// "3" is fine for "b3" when the scale has no natural 3
			if slices.Contains(sc.Degrees, "b"+in) {
				return "b"+in == want, true
			}
			return false, false
		},
		Answer:         fmt.Sprintf("%s (%s)", want, key.Up(degreeSteps(want), sc.Semis[deg])),
		Expect:         want,
		Solution:       all,
		SolutionLabels: labelsFor(all, degOf),
		FastMs:         m.FastMs(),
		Circle:         chordFocus(theory.Triad{Root: key, Quality: scaleQuality(sc)}),
	}
}

func scaleQuality(sc theory.Scale) theory.Quality {
	if sc.ID == theory.MinorPent.ID {
		return theory.Minor
	}
	return theory.Major
}

// degreeSteps turns "b3" into 2 letter steps.
func degreeSteps(deg string) int {
	var d int
	fmt.Sscanf(strings.TrimPrefix(deg, "b"), "%d", &d)
	return d - 1
}
