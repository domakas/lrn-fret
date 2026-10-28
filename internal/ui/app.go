// Package ui is the Bubble Tea app: menu, drills, theory pages, the circle of
// fifths and stats.
package ui

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/domakas/lrn-fret/internal/drill"
	"github.com/domakas/lrn-fret/internal/progress"
)

type screen int

const (
	scrMenu screen = iota
	scrQuiz
	scrTheory
	scrCircle
	scrStats
)

const mixSize = 10

type tickMsg struct{ seq int }
type advanceMsg struct{ seq int }

type quiz struct {
	mod     drill.Module // nil = quick mix
	mixLeft int
	mixPool []string

	cur      drill.Module
	q        *drill.Question
	seq      int
	input    string
	cursor   Pos
	sel      []Pos
	started  time.Time
	pausedAt time.Time

	answered, ok bool
	ms           int
	msg          string
	recent       []string

	total, correct, streak, sumMs int
}

type Model struct {
	store *progress.Store
	rng   *rand.Rand
	mods  []drill.Module

	scr  screen
	back screen // where overlays (theory, circle) return to

	// blurred is set while the terminal doesn't have focus: the screen is
	// hidden and the clock stops. pendingAdvance holds an auto-advance that
	// fired while away, so the next question doesn't start unseen.
	blurred, pendingAdvance bool
	menuIx                  int
	w, h                    int
	status                  string

	quiz *quiz

	theoryTitle, theoryText string
	scroll                  int
}

// New builds the app. start is "", "quick", "stats", "circle" or a module ID.
func New(store *progress.Store, start string) (Model, error) {
	m := Model{
		store: store,
		rng:   rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 42)),
		mods:  drill.All(),
	}
	switch start {
	case "":
	case "quick":
		m.startQuiz(nil)
	case "stats":
		m.scr = scrStats
	case "circle":
		m.scr, m.back = scrCircle, scrMenu
	default:
		mod := m.modByID(start)
		if mod == nil {
			var ids []string
			for _, md := range m.mods {
				ids = append(ids, md.ID())
			}
			return m, fmt.Errorf("unknown drill %q (try: quick, stats, circle, %s)", start, strings.Join(ids, ", "))
		}
		m.startQuiz(mod)
	}
	return m, nil
}

func (m Model) modByID(id string) drill.Module {
	for _, md := range m.mods {
		if md.ID() == id {
			return md
		}
	}
	return nil
}

func (m Model) Init() tea.Cmd { return m.timerCmd() }

// --- quiz flow ---

func (m *Model) startQuiz(mod drill.Module) {
	qz := &quiz{mod: mod}
	if mod == nil {
		qz.mixLeft = mixSize
		for _, md := range m.mods {
			for _, c := range md.Cards() {
				if m.store.Get(c) != nil {
					qz.mixPool = append(qz.mixPool, c)
				}
			}
		}
	}
	m.quiz = qz
	m.scr = scrQuiz
	m.next()
}

func (m *Model) next() {
	qz := m.quiz
	var card string
	if qz.mod == nil && len(qz.mixPool) > 0 {
		card = m.store.Pick(qz.mixPool, 4000, qz.recent, true, m.rng)
		qz.cur = drill.ModuleFor(m.mods, card)
	} else {
		if qz.mod == nil { // nothing practiced yet: mix starts with notes
			qz.cur = m.mods[0]
		} else {
			qz.cur = qz.mod
		}
		card = m.store.Pick(qz.cur.Cards(), qz.cur.FastMs(), qz.recent, false, m.rng)
	}
	qz.recent = append(qz.recent, card)
	if len(qz.recent) > 3 {
		qz.recent = qz.recent[1:]
	}
	qz.q = qz.cur.Make(card, m.rng)
	qz.seq++
	qz.input, qz.sel, qz.msg = "", nil, ""
	qz.answered, qz.ok = false, false
	qz.cursor = qz.q.Cursor
	if qz.cursor.S == 0 {
		qz.cursor = Pos{S: 3, F: 5}
	}
	qz.started = time.Now()
}

func (m Model) timerCmd() tea.Cmd {
	if m.scr != scrQuiz || m.quiz == nil || m.quiz.answered || m.quiz.q.TimeLimit == 0 {
		return nil
	}
	seq := m.quiz.seq
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return tickMsg{seq} })
}

func (m *Model) finish(ok bool, note string) tea.Cmd {
	qz := m.quiz
	q := qz.q
	qz.ms = int(time.Since(qz.started).Milliseconds())
	qz.answered, qz.ok = true, ok
	m.store.Record(q.Card, ok, qz.ms, q.FastMs)
	saveStart := time.Now()
	if err := m.store.Save(); err != nil {
		note = "could not save progress: " + err.Error()
		debugf("save failed: %v", err)
	}
	debugf("answer card=%s ok=%v in %dms (input %q) · save took %v",
		q.Card, ok, qz.ms, qz.input, time.Since(saveStart).Round(time.Microsecond))
	if q.OnResult != nil {
		q.OnResult(qz.sel, ok)
	}
	qz.total++
	if ok {
		qz.correct++
		qz.streak++
		qz.sumMs += qz.ms
	} else {
		qz.streak = 0
	}
	qz.msg = note
	if qz.mod == nil {
		qz.mixLeft--
	}
	if ok {
		seq := qz.seq
		return tea.Tick(900*time.Millisecond, func(time.Time) tea.Msg { return advanceMsg{seq} })
	}
	return nil
}

func (m *Model) advance() tea.Cmd {
	qz := m.quiz
	if qz.mod == nil && qz.mixLeft <= 0 {
		m.status = fmt.Sprintf("Quick mix done: %d/%d correct.", qz.correct, qz.total)
		m.scr = scrMenu
		m.quiz = nil
		return nil
	}
	m.next()
	return m.timerCmd()
}

func (m *Model) leaveQuiz() {
	if qz := m.quiz; qz != nil && qz.total > 0 {
		m.status = fmt.Sprintf("Last session: %d/%d correct%s.", qz.correct, qz.total, avgStr(qz))
	}
	m.quiz = nil
	m.scr = scrMenu
}

func avgStr(qz *quiz) string {
	if qz.correct == 0 {
		return ""
	}
	return fmt.Sprintf(", avg %.1fs", float64(qz.sumMs)/float64(qz.correct)/1000)
}

// The drill clock stops while an overlay is open or the terminal is blurred.
func (m *Model) pauseClock() {
	if m.quiz != nil && m.quiz.pausedAt.IsZero() {
		m.quiz.pausedAt = time.Now()
	}
}

// resumeClock restarts the clock once the drill is visible again.
func (m *Model) resumeClock() tea.Cmd {
	qz := m.quiz
	if m.blurred || m.scr != scrQuiz || qz == nil || qz.pausedAt.IsZero() {
		return nil
	}
	qz.started = qz.started.Add(time.Since(qz.pausedAt))
	qz.pausedAt = time.Time{}
	return m.timerCmd()
}

func (m *Model) openOverlay(s screen) {
	if m.scr == scrQuiz || m.scr == scrMenu {
		m.back = m.scr
	}
	if m.scr == scrQuiz {
		m.pauseClock()
	}
	m.scr = s
}

func (m *Model) closeOverlay() tea.Cmd {
	m.scr = m.back
	return m.resumeClock()
}

func (m *Model) openTheory(mod drill.Module) {
	m.theoryTitle, m.theoryText, m.scroll = mod.Name(), mod.Theory(), 0
	m.openOverlay(scrTheory)
}

// --- update ---

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if dbg == nil {
		return m.update(msg)
	}
	start := time.Now()
	nm, cmd := m.update(msg)
	logUpdate(msg, nm.(Model), time.Since(start))
	return nm, cmd
}

func (m Model) update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		return m, nil
	case tea.BlurMsg:
		m.blurred = true
		if m.scr == scrQuiz {
			m.pauseClock()
		}
		return m, nil
	case tea.FocusMsg:
		m.blurred = false
		if m.pendingAdvance {
			m.pendingAdvance = false
			if m.scr == scrQuiz && m.quiz != nil && m.quiz.answered {
				return m, m.advance()
			}
		}
		return m, m.resumeClock()
	case tickMsg:
		qz := m.quiz
		if m.scr != scrQuiz || qz == nil || qz.seq != msg.seq || qz.answered || !qz.pausedAt.IsZero() {
			return m, nil
		}
		if time.Since(qz.started) >= qz.q.TimeLimit {
			return m, m.finish(false, "time's up")
		}
		return m, m.timerCmd()
	case advanceMsg:
		if m.scr == scrQuiz && m.quiz != nil && m.quiz.seq == msg.seq && m.quiz.answered {
			if m.blurred {
				m.pendingAdvance = true
				return m, nil
			}
			return m, m.advance()
		}
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		switch m.scr {
		case scrMenu:
			return m.updateMenu(msg)
		case scrQuiz:
			return m.updateQuiz(msg)
		case scrTheory:
			return m.updateTheory(msg)
		case scrCircle:
			switch msg.String() {
			case "?":
				m.theoryTitle, m.theoryText, m.scroll = "Circle of fifths", drill.Circle{}.Theory(), 0
				m.scr = scrTheory
			case "esc", "q", "tab", "enter", " ":
				return m, m.closeOverlay()
			}
		case scrStats:
			switch msg.String() {
			case "esc", "q", "enter", "p":
				m.scr = scrMenu
			}
		}
	}
	return m, nil
}

func (m Model) menuLen() int { return len(m.mods) + 3 } // + quick mix, circle, stats

func (m Model) updateMenu(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "q", "esc":
		return m, tea.Quit
	case "up", "w":
		m.menuIx = (m.menuIx + m.menuLen() - 1) % m.menuLen()
	case "down", "s":
		m.menuIx = (m.menuIx + 1) % m.menuLen()
	case "?", "t":
		if m.menuIx < len(m.mods) {
			m.openTheory(m.mods[m.menuIx])
		} else {
			m.openTheory(drill.Circle{})
		}
	case "tab", "c":
		m.openOverlay(scrCircle)
	case "p":
		m.scr = scrStats
	case "m":
		return m, m.choose(len(m.mods))
	case "enter", " ":
		return m, m.choose(m.menuIx)
	default:
		if n, err := strconv.Atoi(k.String()); err == nil && n >= 1 && n <= len(m.mods) {
			m.menuIx = n - 1
			return m, m.choose(n - 1)
		}
	}
	return m, nil
}

func (m *Model) choose(ix int) tea.Cmd {
	m.status = ""
	switch {
	case ix < len(m.mods):
		m.startQuiz(m.mods[ix])
	case ix == len(m.mods):
		m.startQuiz(nil)
	case ix == len(m.mods)+1:
		m.openOverlay(scrCircle)
		return nil
	default:
		m.scr = scrStats
		return nil
	}
	return m.timerCmd()
}

func (m Model) updateTheory(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	lines := strings.Count(m.theoryText, "\n") + 1
	page := max(5, m.bodyHeight())
	switch k.String() {
	case "up", "w":
		m.scroll--
	case "down", "s":
		m.scroll++
	case "pgup", "b":
		m.scroll -= page
	case "pgdown", " ", "f":
		m.scroll += page
	case "tab", "c":
		m.scr = scrCircle
	case "esc", "q", "?", "enter":
		return m, m.closeOverlay()
	}
	m.scroll = max(0, min(m.scroll, lines-page))
	return m, nil
}

func (m Model) bodyHeight() int {
	if m.h == 0 {
		return 30
	}
	return m.h - 3
}

func (m Model) updateQuiz(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	qz := m.quiz
	key := k.String()

	switch key {
	case "esc":
		m.leaveQuiz()
		return m, nil
	case "?":
		m.openTheory(qz.cur)
		return m, nil
	case "tab":
		m.openOverlay(scrCircle)
		return m, nil
	}

	if qz.answered {
		cmd := m.advance()
		// Typing ahead: a note typed during the ✓ pause starts the next
		// answer instead of being swallowed by the skip.
		if typeAhead(k) && m.scr == scrQuiz && m.quiz.q.Kind == drill.AnswerText {
			nm, cmd2 := m.updateQuiz(k)
			return nm, tea.Batch(cmd, cmd2)
		}
		return m, cmd
	}
	if key == "/" {
		return m, m.finish(false, "revealed")
	}

	q := qz.q
	if q.Kind == drill.AnswerText {
		switch key {
		case "enter":
			return m, m.submitText()
		case " ":
			// space submits once the answer has all its words, so a one-note
			// answer is just "f" + space
			words := len(strings.Fields(qz.input))
			if words >= max(1, q.Words) {
				return m, m.submitText()
			}
			if words > 0 && !strings.HasSuffix(qz.input, " ") {
				qz.input += " "
			}
		case "backspace":
			qz.input = dropLast(qz.input)
		default:
			if k.Type == tea.KeyRunes {
				qz.input += string(k.Runes)
				if q.NoteInput {
					qz.input = formatNotes(qz.input)
				}
			}
		}
		return m, nil
	}

	// position answers
	switch key {
	case "up", "w":
		qz.cursor.S = max(1, qz.cursor.S-1)
	case "down", "s":
		qz.cursor.S = min(6, qz.cursor.S+1)
	case "left", "a":
		qz.cursor.F = max(0, qz.cursor.F-1)
	case "right", "d":
		qz.cursor.F = min(q.FretMax, qz.cursor.F+1)
	case "x":
		qz.sel = nil
	case "backspace":
		if qz.input != "" {
			qz.input = dropLast(qz.input)
		} else if len(qz.sel) > 0 {
			qz.sel = qz.sel[:len(qz.sel)-1]
		}
	case " ":
		switch {
		case qz.input != "" && (q.String != 0 || len(strings.Fields(qz.input)) >= 2):
			return m, m.pickPos() // typed fret, or "string fret": done
		case qz.input != "":
			qz.input += " "
		case q.Need == 1 || q.FindAll:
			return m, m.pickPos()
		default:
			qz.toggle(qz.cursor)
		}
	case "enter":
		return m, m.pickPos()
	default:
		if len(key) == 1 && key[0] >= '0' && key[0] <= '9' {
			qz.input += key
		}
	}
	return m, nil
}

func (m *Model) submitText() tea.Cmd {
	qz := m.quiz
	if strings.TrimSpace(qz.input) == "" {
		return nil
	}
	ok, valid := qz.q.CheckText(qz.input)
	if !valid {
		qz.msg = fmt.Sprintf("didn't understand %q", strings.TrimSpace(qz.input))
		debugf("input not understood for card=%s: %q", qz.q.Card, qz.input)
		qz.input = ""
		return nil
	}
	note := ""
	if !ok {
		note = "you typed " + strings.TrimSpace(qz.input)
	}
	return m.finish(ok, note)
}

// pickPos adds the typed position (or the cursor) to the selection and
// submits once enough notes are selected.
func (m *Model) pickPos() tea.Cmd {
	qz := m.quiz
	q := qz.q
	if qz.input != "" {
		p, err := parsePos(qz.input, q)
		qz.input = ""
		if err != nil {
			qz.msg = err.Error()
			return nil
		}
		qz.cursor = p
		if !slices.Contains(qz.sel, p) {
			qz.sel = append(qz.sel, p)
		}
	} else if !slices.Contains(qz.sel, qz.cursor) {
		qz.sel = append(qz.sel, qz.cursor)
	}
	qz.msg = ""
	if q.FindAll && len(qz.sel) > 0 {
		// check each pick as it's made
		if last := qz.sel[len(qz.sel)-1]; !slices.Contains(q.Solution, last) {
			return m.finish(false, "")
		}
	}
	if len(qz.sel) >= q.Need {
		return m.finish(q.CheckPos(qz.sel), "")
	}
	return nil
}

// typeAhead reports whether a key is answer text (a letter, digit or #)
// rather than a control key like enter, space or /.
func typeAhead(k tea.KeyMsg) bool {
	if k.Type != tea.KeyRunes || len(k.Runes) == 0 {
		return false
	}
	r := k.Runes[0]
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '#'
}

func (qz *quiz) toggle(p Pos) {
	if i := slices.Index(qz.sel, p); i >= 0 {
		qz.sel = slices.Delete(qz.sel, i, i+1)
		return
	}
	qz.sel = append(qz.sel, p)
}

// parsePos reads "string fret", or just "fret" when the question fixes the
// string.
func parsePos(in string, q *drill.Question) (Pos, error) {
	f := strings.Fields(in)
	nums := make([]int, len(f))
	for i, s := range f {
		n, err := strconv.Atoi(s)
		if err != nil {
			return Pos{}, fmt.Errorf("type: string fret, e.g. 4 7")
		}
		nums[i] = n
	}
	var p Pos
	switch {
	case len(nums) == 2:
		p = Pos{S: nums[0], F: nums[1]}
	case len(nums) == 1 && q.String != 0:
		p = Pos{S: q.String, F: nums[0]}
	default:
		return Pos{}, fmt.Errorf("type: string fret, e.g. 4 7 (string 1 = high e)")
	}
	if p.S < 1 || p.S > 6 || p.F < 0 || p.F > q.FretMax {
		return Pos{}, fmt.Errorf("no such spot: string %d fret %d", p.S, p.F)
	}
	return p, nil
}

// formatNotes shows typed note names the standard way: the letter that starts
// a word is uppercase, anything right after it lowercase, so "eb" and "EB"
// both read "Eb" and "bb" reads "Bb". A 3 after the letter becomes # (no
// shift needed: "f3" → "F#"). Words not starting with A–G are left alone.
func formatNotes(s string) string {
	words := strings.Split(s, " ")
	for i, w := range words {
		if w == "" || !strings.ContainsRune("abcdefgABCDEFG", rune(w[0])) {
			continue
		}
		words[i] = strings.ToUpper(w[:1]) + strings.ReplaceAll(strings.ToLower(w[1:]), "3", "#")
	}
	return strings.Join(words, " ")
}

func dropLast(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return s
	}
	return string(r[:len(r)-1])
}
