package drill

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/domakas/lrn-fret/internal/theory"
)

// Changes: follow a chord progression inside a fret window, hitting the 1, 3
// or 5 of each chord as it comes, against the clock.
type Changes struct {
	key    theory.Key
	prog   []int
	idx    int
	zoneLo int
	last   *Pos
	lastLb string
}

const zoneWidth = 4 // window is zoneLo..zoneLo+4 (5 frets)
const loops = 2     // times through a progression before a new one

func (*Changes) ID() string   { return "changes" }
func (*Changes) Name() string { return "Chord changes" }
func (*Changes) Desc() string {
	return "Follow a progression in one zone: hit the 1/3/5 of each chord in time"
}
func (*Changes) FastMs() int { return 5000 }

func (*Changes) Cards() []string { return []string{"chg:1", "chg:3", "chg:5"} }

func (c *Changes) newProgression(rng *rand.Rand) {
	c.key = theory.Key{Tonic: theory.MustNote(pick(rng, theory.Keys))}
	c.prog = pick(rng, theory.Progressions)
	c.idx = 0
	c.zoneLo = 1 + rng.IntN(9)
	c.last = nil
}

func (c *Changes) Make(card string, rng *rand.Rand) *Question {
	if c.prog == nil || c.idx >= len(c.prog)*loops {
		c.newProgression(rng)
	}
	var deg int
	fmt.Sscanf(card, "chg:%d", &deg)
	cur := c.idx % len(c.prog)
	chord := c.key.Chord(c.prog[cur])
	c.idx++
	tone := chord.Tone(deg)
	lo, hi := c.zoneLo, c.zoneLo+zoneWidth

	var bar []string
	for i, d := range c.prog {
		name := c.key.Chord(d).Name()
		if i == cur {
			name = "[" + name + "]"
		}
		bar = append(bar, name)
	}
	sol := theory.Find(tone.PC, lo, hi)
	marks := map[Pos]Mark{}
	cursor := Pos{S: 3, F: lo + 2}
	if c.last != nil {
		marks[*c.last] = Mark{MarkPrev, c.lastLb}
		cursor = *c.last
	}
	degName := map[int]string{1: "root", 3: "3rd", 5: "5th"}[deg]
	q := &Question{
		Card: card,
		Prompt: fmt.Sprintf("Key of %s:  %s   (loop %d/%d)\nNow %s → play its %s, frets %d–%d",
			c.key.Tonic, strings.Join(bar, "  "), (c.idx-1)/len(c.prog)+1, loops, chord.Name(), degName, lo, hi),
		Kind:    AnswerPos,
		Need:    1,
		Cursor:  cursor,
		Marks:   marks,
		FretMax: max(12, hi),
		Window:  [2]int{lo, hi},
		CheckPos: func(sel []Pos) bool {
			return len(sel) == 1 && sel[0].F >= lo && sel[0].F <= hi && theory.PitchAt(sel[0]) == tone.PC
		},
		Answer:         fmt.Sprintf("%s (%s of %s = %s %s %s)", tone, degName, chord.Name(), chord.Tone(1), chord.Tone(3), chord.Tone(5)),
		Solution:       sol,
		SolutionLabels: labelsFor(sol, func(Pos) string { return tone.String() }),
		TimeLimit:      6 * time.Second,
		FastMs:         c.FastMs(),
		Circle:         c.focus(chord),
	}
	q.OnResult = func(sel []Pos, ok bool) {
		p := nearest(sol, c.last)
		if ok {
			p = sel[0]
		}
		c.last, c.lastLb = &p, tone.String()
	}
	return q
}

// focus shows the key and the chord being played on the circle.
func (c *Changes) focus(chord theory.Triad) *CircleFocus {
	f := &CircleFocus{Strong: []theory.PC{c.key.Tonic.PC}}
	if chord.Quality == theory.Minor {
		f.Minor = []theory.PC{chord.Root.PC}
	} else {
		f.Major = []theory.PC{chord.Root.PC}
	}
	return f
}

// nearest picks the solution closest to the previous note (smallest move).
func nearest(ps []Pos, from *Pos) Pos {
	if from == nil {
		return ps[0]
	}
	best, bestD := ps[0], 99
	for _, p := range ps {
		d := abs(p.S-from.S) + abs(p.F-from.F)
		if d < bestD {
			best, bestD = p, d
		}
	}
	return best
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
