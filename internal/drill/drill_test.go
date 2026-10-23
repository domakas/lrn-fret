package drill

import (
	"math/rand/v2"
	"testing"
)

// Every card of every module must build a question whose own answer passes
// its checker (and, for position questions, whose solution passes).
func TestEveryCardAnswerable(t *testing.T) {
	rng := rand.New(rand.NewPCG(7, 7))
	for _, m := range All() {
		for _, card := range m.Cards() {
			for range 5 { // several random roots/keys per card
				q := m.Make(card, rng)
				switch q.Kind {
				case AnswerText:
					if q.Expect == "" {
						t.Fatalf("%s: no Expect", card)
					}
					ok, valid := q.CheckText(q.Expect)
					if !ok || !valid {
						t.Fatalf("%s: expected answer %q rejected (prompt %q)", card, q.Expect, q.Prompt)
					}
				case AnswerPos:
					if len(q.Solution) < q.Need {
						t.Fatalf("%s: solution too short: %v", card, q.Solution)
					}
					if !q.CheckPos(q.Solution[:q.Need]) {
						t.Fatalf("%s: solution %v rejected (prompt %q)", card, q.Solution[:q.Need], q.Prompt)
					}
				}
				if q.OnResult != nil && len(q.Solution) > 0 {
					q.OnResult(q.Solution[:1], true)
				}
			}
		}
	}
}

func TestWrongAnswersRejected(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 1))
	q := Notes{}.Make("note:s5f3", rng) // C
	if ok, valid := q.CheckText("d"); ok || !valid {
		t.Error("d should be a valid wrong answer for C")
	}
	if _, valid := q.CheckText("x"); valid {
		t.Error("x should be invalid input")
	}
	q = Spell{}.Make("spell:Eb:min", rng)
	if ok, _ := q.CheckText("eb g bb"); ok {
		t.Error("Eb G Bb is major, not Ebm")
	}
	if ok, _ := q.CheckText("d# f# a#"); !ok {
		t.Error("enharmonic spelling should be accepted")
	}
	q = Scales{}.Make("scale:minpent:0:1", rng)
	if ok, valid := q.CheckText("3"); !ok || !valid {
		t.Error(`"3" should count as b3 in minor pentatonic`)
	}
}

func TestFindCursorFollowsLastPick(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 1))
	fd := &Find{}
	q := fd.Make("find:0:s5", rng) // C on the A string
	if q.Cursor != (Pos{S: 5, F: 0}) {
		t.Errorf("first cursor = %v; want string 5 fret 0", q.Cursor)
	}
	q.OnResult([]Pos{{S: 5, F: 3}}, true)
	q = fd.Make("find:7:s4", rng) // G on the D string
	if q.Cursor != (Pos{S: 4, F: 3}) {
		t.Errorf("cursor = %v; want string 4 fret 3 (last pick's fret)", q.Cursor)
	}
	q.OnResult(nil, false) // revealed: stays
	if q = fd.Make("find:2:s6", rng); q.Cursor.F != 3 {
		t.Errorf("after reveal cursor fret = %d; want 3", q.Cursor.F)
	}
}
