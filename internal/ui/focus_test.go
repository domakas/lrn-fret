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
