package drill

import (
	"fmt"
	"math/rand/v2"
	"slices"

	"github.com/domakas/lrn-fret/internal/theory"
)

// Intervals: a root is marked, find an interval within reach.
type Intervals struct{}

// curriculum order: the chord-tone intervals first.
var intervalOrder = []string{"8", "5", "3", "b3", "4", "b7", "7", "2", "6", "b6", "b2", "b5"}

const reach = 4 // frets either side of the root

func (Intervals) ID() string   { return "intervals" }
func (Intervals) Name() string { return "Intervals" }
func (Intervals) Desc() string { return "Root is marked R: find the 3, b3, 5, b7… within 4 frets" }
func (Intervals) FastMs() int  { return 4000 }

func (Intervals) Cards() []string {
	var out []string
	for _, iv := range intervalOrder {
		for _, s := range stringOrder {
			out = append(out, fmt.Sprintf("int:%s:s%d", iv, s))
		}
	}
	return out
}

func (iv Intervals) Make(card string, rng *rand.Rand) *Question {
	var name string
	var s int
	fmt.Sscanf(card, "int:%s", &name)
	// Sscanf %s grabs "3:s4"; split manually
	for i := range name {
		if name[i] == ':' {
			fmt.Sscanf(name[i:], ":s%d", &s)
			name = name[:i]
			break
		}
	}
	interval := theory.IntervalByName(name)
	root := Pos{S: s, F: 2 + rng.IntN(9)} // frets 2..10
	rootNote := spellPC(theory.PitchAt(root))
	target := rootNote.Up(interval.Steps, interval.Semis)

	lo, hi := max(0, root.F-reach), root.F+reach
	var sol []Pos
	for _, p := range theory.Find(target.PC, lo, hi) {
		if p != root {
			sol = append(sol, p)
		}
	}
	label := interval.Name
	if label == "8" {
		label = "R"
	}
	q := &Question{
		Card: card,
		Prompt: fmt.Sprintf("Root %s on the %s string. Find a %s (%s) — frets %d–%d",
			rootNote, theory.StringLabel(s), interval.Name, interval.Long, lo, hi),
		Kind:    AnswerPos,
		Need:    1,
		Cursor:  root,
		Marks:   map[Pos]Mark{root: {MarkRoot, "R"}},
		FretMax: max(12, hi),
		Window:  [2]int{lo, hi},
		CheckPos: func(sel []Pos) bool {
			if len(sel) != 1 {
				return false
			}
			p := sel[0]
			return p != root && p.F >= lo && p.F <= hi && theory.PitchAt(p) == target.PC
		},
		Answer:         fmt.Sprintf("%s (%s of %s)", target, interval.Name, rootNote),
		Solution:       sol,
		SolutionLabels: labelsFor(sol, func(Pos) string { return label }),
		FastMs:         iv.FastMs(),
		Circle:         &CircleFocus{Strong: []theory.PC{rootNote.PC}},
	}
	if interval.Name == "8" {
		// octaves: find every one in reach, it's the map of where a note lives
		q.FindAll, q.Need = true, len(sol)
		q.Prompt = fmt.Sprintf("Root %s on the %s string. Find every octave (%d) — frets %d–%d",
			rootNote, theory.StringLabel(s), len(sol), lo, hi)
		q.CheckPos = func(sel []Pos) bool {
			found := map[Pos]bool{}
			for _, p := range sel {
				if !slices.Contains(sol, p) {
					return false
				}
				found[p] = true
			}
			return len(found) == len(sol)
		}
		q.FastMs = iv.FastMs() + 2000*(len(sol)-1)
	}
	return q
}
