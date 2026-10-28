package drill

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"

	"github.com/domakas/lrn-fret/internal/theory"
)

var qualities = []theory.Quality{theory.Major, theory.Minor}

func parseQuality(s string) theory.Quality {
	for _, q := range []theory.Quality{theory.Major, theory.Minor, theory.Dim} {
		if q.Key() == s {
			return q
		}
	}
	panic("bad quality " + s)
}

// Spell: type the notes of a triad.
type Spell struct{}

func (Spell) ID() string   { return "spell" }
func (Spell) Name() string { return "Triad spelling" }
func (Spell) Desc() string { return "Type the 1 3 5 of a chord: Ebm → eb gb bb" }
func (Spell) FastMs() int  { return 6000 }

func (Spell) Cards() []string {
	var out []string
	for _, r := range theory.Roots {
		for _, q := range qualities {
			out = append(out, fmt.Sprintf("spell:%s:%s", r, q.Key()))
		}
	}
	return out
}

func (sp Spell) Make(card string, _ *rand.Rand) *Question {
	parts := strings.Split(card, ":")
	t := theory.Triad{Root: theory.MustNote(parts[1]), Quality: parseQuality(parts[2])}
	tones := t.Tones()
	var want []string
	for _, n := range tones {
		want = append(want, n.String())
	}
	return &Question{
		Card:      card,
		Prompt:    fmt.Sprintf("Spell %s (%s triad): type 1 3 5, e.g. \"c e g\"", t.Name(), t.Quality.Name()),
		Kind:      AnswerText,
		NoBoard:   true,
		Words:     3,
		NoteInput: true,
		CheckText: func(in string) (bool, bool) {
			fields := strings.FieldsFunc(in, func(r rune) bool { return r == ' ' || r == ',' })
			if len(fields) != 3 {
				return false, false
			}
			for i, f := range fields {
				pc, err := theory.ParseNote(f)
				if err != nil {
					return false, false
				}
				if pc != tones[i].PC {
					return false, true
				}
			}
			return true, true
		},
		Answer: strings.Join(want, " "),
		Expect: strings.Join(want, " "),
		FastMs: sp.FastMs(),
		Circle: chordFocus(t),
	}
}

// Triads: find triad shapes on 3-string sets, and name their chord tones.
type Triads struct{}

func (Triads) ID() string   { return "triads" }
func (Triads) Name() string { return "Triad shapes" }
func (Triads) Desc() string {
	return "Triads on string sets in every inversion, and which note is the 1/3/5"
}
func (Triads) FastMs() int { return 6000 }

func (Triads) Cards() []string {
	var out []string
	for ss := 1; ss <= 4; ss++ {
		for _, q := range qualities {
			for inv := 0; inv < 3; inv++ {
				out = append(out,
					fmt.Sprintf("tri:%s:%d:%d", q.Key(), ss, inv),
					fmt.Sprintf("trideg:%s:%d:%d", q.Key(), ss, inv))
			}
		}
	}
	return out
}

func (tr Triads) Make(card string, rng *rand.Rand) *Question {
	parts := strings.Split(card, ":")
	q := parseQuality(parts[1])
	var ss theory.StringSet
	var inv int
	fmt.Sscanf(parts[2], "%d", &ss)
	fmt.Sscanf(parts[3], "%d", &inv)

	var t theory.Triad
	var voicings [][3]Pos
	for len(voicings) == 0 {
		t = theory.Triad{Root: theory.MustNote(pick(rng, theory.Roots)), Quality: q}
		voicings = theory.Voicings(t, ss, inv, 15)
	}
	degOf := func(p Pos) string {
		pc := theory.PitchAt(p)
		for _, d := range []int{1, 3, 5} {
			if t.Tone(d).PC == pc {
				return fmt.Sprint(d)
			}
		}
		return "?"
	}
	strs := ss.Strings()
	fretMax := 12
	for _, v := range voicings {
		for _, p := range v {
			fretMax = max(fretMax, p.F)
		}
	}

	if parts[0] == "tri" {
		var sol []Pos
		for _, v := range voicings {
			sol = append(sol, v[:]...)
		}
		order := theory.InversionOrder(inv)
		return &Question{
			Card: card,
			Prompt: fmt.Sprintf("Play %s, %s (%d in the bass), on strings %s.\nSelect 3 notes: space to mark, enter to submit.",
				t.Name(), theory.InversionName(inv), order[0], ss),
			Kind:    AnswerPos,
			Need:    3,
			Cursor:  Pos{S: strs[1], F: 5},
			Marks:   map[Pos]Mark{},
			FretMax: fretMax,
			CheckPos: func(sel []Pos) bool {
				return theory.MatchesVoicing(sel, t, ss, inv)
			},
			Answer:         fmt.Sprintf("%s = %s %s %s", t.Name(), t.Tone(1), t.Tone(3), t.Tone(5)),
			Solution:       sol,
			SolutionLabels: labelsFor(sol, degOf),
			FastMs:         10000,
			Circle:         chordFocus(t),
		}
	}

	v := pick(rng, voicings)
	target := v[rng.IntN(3)]
	marks := map[Pos]Mark{}
	for _, p := range v {
		marks[p] = Mark{MarkDot, "●"}
	}
	marks[target] = Mark{MarkTarget, "?"}
	want := degOf(target)
	return &Question{
		Card:    card,
		Prompt:  fmt.Sprintf("This is %s. Which chord tone is the ? — 1, 3 or 5?", t.Name()),
		Kind:    AnswerText,
		Marks:   marks,
		FretMax: fretMax,
		CheckText: func(in string) (bool, bool) {
			in = normalize(in)
			if in == "r" || in == "root" {
				in = "1"
			}
			if !slices.Contains([]string{"1", "3", "5"}, in) {
				return false, false
			}
			return in == want, true
		},
		Answer:         fmt.Sprintf("%s (%s)", want, noteName(t, want)),
		Expect:         want,
		Solution:       v[:],
		SolutionLabels: labelsFor(v[:], degOf),
		FastMs:         4000,
		Circle:         chordFocus(t),
	}
}

// chordFocus highlights a chord on the circle: majors on the outer ring,
// minors on the inner one.
func chordFocus(t theory.Triad) *CircleFocus {
	if t.Quality == theory.Major {
		return &CircleFocus{Strong: []theory.PC{t.Root.PC}}
	}
	return &CircleFocus{Minor: []theory.PC{t.Root.PC}}
}

func noteName(t theory.Triad, deg string) string {
	var d int
	fmt.Sscanf(deg, "%d", &d)
	return t.Tone(d).String()
}
