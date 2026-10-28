package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/domakas/lrn-fret/internal/progress"
)

func newTestModel(t *testing.T, start string) Model {
	t.Helper()
	m, err := New(&progress.Store{Cards: map[string]*progress.Card{}}, start)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func send(m Model, msg tea.Msg) Model {
	nm, _ := m.Update(msg)
	return nm.(Model)
}

func TestBlurHidesAndStopsClock(t *testing.T) {
	m := newTestModel(t, "changes")
	m = send(m, tea.BlurMsg{})
	if !strings.Contains(m.View(), "paused") || strings.Contains(m.View(), "Key of") {
		t.Fatalf("blurred view should hide the drill:\n%s", m.View())
	}
	// pretend we were away for a minute
	m.quiz.started = m.quiz.started.Add(-time.Minute)
	m.quiz.pausedAt = m.quiz.pausedAt.Add(-time.Minute)
	// a timer tick while away must not time the question out
	m = send(m, tickMsg{m.quiz.seq})
	if m.quiz.answered {
		t.Fatal("question timed out while blurred")
	}
	m = send(m, tea.FocusMsg{})
	if el := time.Since(m.quiz.started); el > 2*time.Second {
		t.Errorf("time away counted: elapsed %v", el)
	}
	if strings.Contains(m.View(), "paused") {
		t.Error("view still paused after focus")
	}
}

func TestAdvanceWaitsForFocus(t *testing.T) {
	m := newTestModel(t, "notes")
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")}) // first card is low E open
	m = send(m, tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")})
	if !m.quiz.answered || !m.quiz.ok {
		t.Fatalf("expected a correct answer, got answered=%v ok=%v", m.quiz.answered, m.quiz.ok)
	}
	seq := m.quiz.seq
	m = send(m, tea.BlurMsg{})
	m = send(m, advanceMsg{seq})
	if m.quiz.seq != seq {
		t.Fatal("advanced to the next question while blurred")
	}
	m = send(m, tea.FocusMsg{})
	if m.quiz.seq == seq {
		t.Fatal("did not advance after focus came back")
	}
}

func TestWASDMovesCursor(t *testing.T) {
	m := newTestModel(t, "intervals")
	start := m.quiz.cursor
	key := func(r string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(r)} }
	m = send(m, key("d"))
	m = send(m, key("w"))
	want := Pos{S: max(1, start.S-1), F: start.F + 1}
	if m.quiz.cursor != want {
		t.Errorf("after d,w cursor = %v; want %v", m.quiz.cursor, want)
	}
	m = send(m, key("a"))
	m = send(m, key("s"))
	if m.quiz.cursor.F != start.F {
		t.Errorf("a should move back a fret: %v", m.quiz.cursor)
	}
}

func TestMenuWASDAndStats(t *testing.T) {
	m := newTestModel(t, "")
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	if m.menuIx != 1 {
		t.Errorf("s should move down the menu, menuIx = %d", m.menuIx)
	}
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	if m.scr != scrStats {
		t.Error("p should open stats")
	}
}

func TestOctavesFindAll(t *testing.T) {
	m := newTestModel(t, "intervals") // first card is the octave on the low E
	q := m.quiz.q
	if !q.FindAll || q.Need != len(q.Solution) || q.Need < 2 {
		t.Fatalf("octave question should be find-all, got FindAll=%v Need=%d sol=%v", q.FindAll, q.Need, q.Solution)
	}
	space := tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")}
	for i, p := range q.Solution {
		m.quiz.cursor = p
		m = send(m, space)
		if i < len(q.Solution)-1 {
			m = send(m, space) // picking the same spot twice is ignored
			if m.quiz.answered {
				t.Fatalf("answered after %d of %d", i+1, len(q.Solution))
			}
			if !strings.Contains(m.View(), fmt.Sprintf("found %d/%d", i+1, q.Need)) {
				t.Errorf("view should show found %d/%d", i+1, q.Need)
			}
		}
	}
	if !m.quiz.answered || !m.quiz.ok {
		t.Fatal("finding every octave should be correct")
	}

	m = newTestModel(t, "intervals")
	m.quiz.cursor = m.quiz.q.Cursor // the root itself is not an answer
	m.quiz.cursor.F++
	m = send(m, space)
	if !m.quiz.answered || m.quiz.ok {
		t.Fatal("a wrong pick should end the question as a miss")
	}
}

func TestFormatNotes(t *testing.T) {
	cases := map[string]string{
		"c":        "C",
		"eb gb bb": "Eb Gb Bb",
		"EB GB BB": "Eb Gb Bb",
		"bB c# F#": "Bb C# F#",
		"am":       "Am",
		"eb ":      "Eb ",
		"x 3b":     "x 3b",
		"f3":       "F#",
		"c3 e3 g3": "C# E# G#",
		"F33":      "F##",
	}
	for in, want := range cases {
		if got := formatNotes(in); got != want {
			t.Errorf("formatNotes(%q) = %q; want %q", in, got, want)
		}
	}
}

func TestSpellInputShownAsNotes(t *testing.T) {
	m := newTestModel(t, "spell")
	for _, k := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune("E")},
		{Type: tea.KeyRunes, Runes: []rune("B")},
		{Type: tea.KeySpace, Runes: []rune(" ")},
		{Type: tea.KeyRunes, Runes: []rune("g")},
		{Type: tea.KeyRunes, Runes: []rune("B")},
	} {
		m = send(m, k)
	}
	if m.quiz.input != "Eb Gb" {
		t.Errorf("input = %q; want %q", m.quiz.input, "Eb Gb")
	}
}

func TestDegreeInputNotFormatted(t *testing.T) {
	m := newTestModel(t, "scales")
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	if m.quiz.input != "b" {
		t.Errorf("degree input = %q; want %q (b3 must not become B3)", m.quiz.input, "b")
	}
}

func TestThreeIsSharpInNoteAnswers(t *testing.T) {
	m := newTestModel(t, "notes")
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("3")})
	if m.quiz.input != "F#" {
		t.Errorf("input = %q; want F#", m.quiz.input)
	}
	if ok, valid := m.quiz.q.CheckText(m.quiz.input); !valid {
		t.Errorf("F# should parse (ok=%v)", ok)
	}
}

func TestTypeAheadDuringPauseIsKept(t *testing.T) {
	m := newTestModel(t, "spell") // first card is C major
	typeStr := func(m Model, s string) Model {
		for _, r := range s {
			if r == ' ' {
				m = send(m, tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")})
			} else {
				m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			}
		}
		return m
	}
	m = typeStr(m, "c e g ")
	if !m.quiz.answered || !m.quiz.ok {
		t.Fatal("C E G should be answered correctly")
	}
	seq := m.quiz.seq
	m = typeStr(m, "c") // during the ✓ pause
	if m.quiz.seq == seq {
		t.Fatal("typing during the pause should move to the next question")
	}
	if m.quiz.input != "C" {
		t.Errorf("first typed note was lost: input = %q; want %q", m.quiz.input, "C")
	}

	// control keys still only continue
	m = newTestModel(t, "spell")
	m = typeStr(m, "c e g ")
	m = send(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.quiz.answered || m.quiz.input != "" {
		t.Errorf("enter should just continue: answered=%v input=%q", m.quiz.answered, m.quiz.input)
	}
}
