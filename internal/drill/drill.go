// Package drill turns cards (facts to learn) into questions. Modules are
// pure: they build a Question with its checker, and the UI asks it.
package drill

import (
	"math/rand/v2"
	"strings"
	"time"

	"github.com/domakas/lrn-fret/internal/theory"
)

type Pos = theory.Pos

type MarkKind int

const (
	MarkDot    MarkKind = iota // a plain note dot
	MarkRoot                   // the root
	MarkTarget                 // the "?" being asked about
	MarkPrev                   // context, e.g. the note you played on the last chord
)

type Mark struct {
	Kind  MarkKind
	Label string
}

type AnswerKind int

const (
	AnswerText AnswerKind = iota // type a note, degree, spelling
	AnswerPos                    // pick positions on the fretboard
)

type Question struct {
	Card   string
	Prompt string // may span several lines
	Kind   AnswerKind

	Marks   map[Pos]Mark
	NoBoard bool
	FretMax int    // highest fret to draw (≥ 12)
	Window  [2]int // fret window to highlight; [0,0] = none

	// AnswerPos only
	Need int // positions to select
	// FindAll: find every position in Solution (Need = len(Solution)). Each
	// pick is checked as you go; a wrong pick ends the question.
	FindAll bool
	String  int // if set, the answer is on this string and a bare number is a fret
	Cursor  Pos // where the cursor starts

	// CheckText returns ok, and valid=false if the input didn't parse (so the
	// UI can ask again without counting a miss).
	CheckText func(string) (ok, valid bool)
	Words     int // AnswerText: words in the answer (0 = 1); space submits after the last one
	// NoteInput: the answer is note names, so the UI shows typing as notes
	// ("eb" → "Eb"). Not set for degrees like "b3".
	NoteInput bool
	CheckPos  func([]Pos) bool

	Answer         string         // shown after answering, e.g. "A" or "Eb Gb Bb"
	Expect         string         // AnswerText: an input that is accepted (used by tests)
	Solution       []Pos          // correct positions to reveal
	SolutionLabels map[Pos]string // optional labels for Solution

	TimeLimit time.Duration
	FastMs    int // a correct answer this fast counts as "known"

	// OnResult lets stateful modules see the outcome.
	OnResult func(sel []Pos, ok bool)

	// Circle is what to highlight on the circle-of-fifths view, if anything.
	Circle *CircleFocus
}

// CircleFocus highlights keys on the circle of fifths: Strong is the key or
// root in play, Major and Minor are related chords.
type CircleFocus struct {
	Strong       []theory.PC
	Major, Minor []theory.PC
}

type Module interface {
	ID() string
	Name() string
	Desc() string
	Cards() []string // curriculum order
	FastMs() int
	Make(card string, rng *rand.Rand) *Question
	Theory() string // the help page for this module
}

func All() []Module {
	return []Module{
		&Notes{}, &Find{}, &Intervals{}, &Spell{}, &Triads{}, &Scales{}, &Circle{}, &Changes{},
	}
}

// ModuleFor finds the module a card belongs to.
func ModuleFor(mods []Module, card string) Module {
	for _, m := range mods {
		for _, c := range m.Cards() {
			if c == card {
				return m
			}
		}
	}
	return nil
}

func ByID(id string) Module {
	for _, m := range All() {
		if m.ID() == id {
			return m
		}
	}
	return nil
}

// --- shared helpers ---

// preferred spelling of each pitch class for roots, e.g. Bb not A#.
var preferred = func() map[theory.PC]theory.Note {
	m := map[theory.PC]theory.Note{}
	for _, r := range theory.Roots {
		n := theory.MustNote(r)
		if _, ok := m[n.PC]; !ok {
			m[n.PC] = n
		}
	}
	return m
}()

func spellPC(pc theory.PC) theory.Note { return preferred[pc.Norm()] }

func pick[T any](rng *rand.Rand, xs []T) T { return xs[rng.IntN(len(xs))] }

func normalize(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

func checkNote(want theory.PC) func(string) (bool, bool) {
	return func(in string) (bool, bool) {
		pc, err := theory.ParseNote(in)
		if err != nil {
			return false, false
		}
		return pc == want, true
	}
}

func labelsFor(ps []Pos, label func(Pos) string) map[Pos]string {
	m := map[Pos]string{}
	for _, p := range ps {
		m[p] = label(p)
	}
	return m
}

func noteLabel(p Pos) string { return spellPC(theory.PitchAt(p)).String() }
