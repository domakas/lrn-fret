package ui

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/domakas/lrn-fret/internal/drill"
	"github.com/domakas/lrn-fret/internal/theory"
)

func (m Model) View() string {
	if dbg != nil {
		start := time.Now()
		defer func() {
			if took := time.Since(start); took >= slow {
				debugf("render %s  [%v]  SLOW", screenNames[m.scr], took.Round(time.Microsecond))
			}
		}()
	}
	if m.blurred {
		return m.viewBlurred()
	}
	switch m.scr {
	case scrQuiz:
		return m.viewQuiz()
	case scrTheory:
		return m.viewTheory()
	case scrCircle:
		return m.viewCircle()
	case scrStats:
		return m.viewStats()
	}
	return m.viewMenu()
}

// viewBlurred hides everything while the terminal is in the background.
func (m Model) viewBlurred() string {
	card := sTitle.Render("lrn-fret") + sDim.Render(" · paused")
	if qz := m.quiz; qz != nil && qz.total > 0 {
		card += "\n" + sDim.Render(fmt.Sprintf("✓ %d/%d%s", qz.correct, qz.total, avgStr(qz)))
	}
	card += "\n\n" + sMuted.Render("focus this window to continue")
	box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("8")).
		Padding(1, 3).Render(card)
	if m.w == 0 || m.h == 0 {
		return box
	}
	return lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, box)
}

func keys(pairs ...string) string {
	var out []string
	for i := 0; i+1 < len(pairs); i += 2 {
		out = append(out, sKeyHint.Render(pairs[i])+" "+sDim.Render(pairs[i+1]))
	}
	return strings.Join(out, sDim.Render(" · "))
}

func bar(n, total, width int) string {
	if total == 0 {
		return strings.Repeat("░", width)
	}
	full := n * width / total
	if n > 0 && full == 0 {
		full = 1
	}
	return sGood.Render(strings.Repeat("█", full)) + sDim.Render(strings.Repeat("░", width-full))
}

// --- menu ---

func (m Model) viewMenu() string {
	var b strings.Builder
	b.WriteString(sTitle.Render("lrn-fret") + sDim.Render("  fretboard trainer") + "\n\n")
	item := func(ix int, key, name, right string) {
		ptr := "  "
		st := sPlain
		if ix == m.menuIx {
			ptr, st = sKeyHint.Render("▸ "), sBold
		}
		fmt.Fprintf(&b, "%s%s %s %s\n", ptr, sDim.Render(key), st.Render(fmt.Sprintf("%-18s", name)), right)
	}
	for i, md := range m.mods {
		cards := md.Cards()
		seen, mastered := m.store.Summary(cards)
		right := bar(mastered, len(cards), 16) + sDim.Render(fmt.Sprintf(" %d/%d", mastered, len(cards)))
		if seen > mastered {
			right += sWarn.Render(fmt.Sprintf("  %d learning", seen-mastered))
		}
		item(i, fmt.Sprint(i+1), md.Name(), right)
	}
	b.WriteString("\n")
	item(len(m.mods), "m", "Quick mix", sDim.Render(fmt.Sprintf("%d of your weakest, all drills", mixSize)))
	item(len(m.mods)+1, "c", "Circle of fifths", sDim.Render("view"))
	item(len(m.mods)+2, "p", "Stats", sDim.Render("fretboard heatmap"))

	b.WriteString("\n")
	if m.menuIx < len(m.mods) {
		b.WriteString(sMuted.Render("  "+m.mods[m.menuIx].Desc()) + "\n")
	} else {
		b.WriteString("\n")
	}
	if m.status != "" {
		b.WriteString(sGood.Render("  "+m.status) + "\n")
	}
	b.WriteString("\n" + keys("enter", "start", "?", "theory", "tab", "circle", "p", "stats", "q", "quit"))
	return b.String()
}

// --- quiz ---

func (m Model) viewQuiz() string {
	qz := m.quiz
	q := qz.q
	var b strings.Builder

	title := qz.cur.Name()
	if qz.mod == nil {
		title = fmt.Sprintf("Quick mix %d/%d · %s", mixSize-qz.mixLeft+boolInt(!qz.answered), mixSize, qz.cur.Name())
	}
	stats := fmt.Sprintf("  ✓ %d/%d", qz.correct, qz.total)
	if qz.streak >= 3 {
		stats += fmt.Sprintf(" · streak %d", qz.streak)
	}
	stats += avgStr(qz)
	b.WriteString(sTitle.Render(title) + sDim.Render(stats) + "\n\n")

	for _, line := range strings.Split(q.Prompt, "\n") {
		b.WriteString(sBold.Render(line) + "\n")
	}
	if q.TimeLimit > 0 && !qz.answered {
		left := q.TimeLimit - time.Since(qz.started)
		if !qz.pausedAt.IsZero() {
			left = q.TimeLimit - qz.pausedAt.Sub(qz.started)
		}
		left = max(0, left)
		w := 24
		n := int(float64(w) * float64(left) / float64(q.TimeLimit))
		st := sGood
		if left < q.TimeLimit/3 {
			st = sBad
		}
		b.WriteString(st.Render(strings.Repeat("█", n)) + sDim.Render(strings.Repeat("░", w-n)) +
			sDim.Render(fmt.Sprintf(" %.1fs", left.Seconds())) + "\n")
	}
	b.WriteString("\n")

	if !q.NoBoard {
		b.WriteString(m.quizBoard().render() + "\n\n")
	}

	switch {
	case qz.answered && qz.ok:
		b.WriteString(sGood.Render("✓ "+q.Answer) + sDim.Render(fmt.Sprintf("  %.1fs", float64(qz.ms)/1000)))
	case qz.answered:
		b.WriteString(sBad.Render("✗ ") + sBold.Render(q.Answer))
		if qz.msg != "" {
			b.WriteString(sDim.Render("  (" + qz.msg + ")"))
		}
		b.WriteString("\n" + sDim.Render("the answer is shown in green · any key to continue"))
	case q.Kind == drill.AnswerText:
		b.WriteString(sKeyHint.Render("> ") + qz.input + sDim.Render("█"))
	default:
		hint := "wasd move · space/enter pick"
		if q.FindAll {
			hint = fmt.Sprintf("found %d/%d · wasd move · space/enter pick", len(qz.sel), q.Need)
		} else if q.Need > 1 {
			hint = fmt.Sprintf("wasd move · space mark %d/%d · enter submit", len(qz.sel), q.Need)
		}
		if q.String != 0 {
			hint += " · or type a fret"
		} else {
			hint += " · or type: string fret"
		}
		b.WriteString(sKeyHint.Render("> ") + qz.input + sDim.Render("█  "+hint))
	}
	b.WriteString("\n")
	if !qz.answered && qz.msg != "" {
		b.WriteString(sWarn.Render(qz.msg))
	}
	b.WriteString("\n" + keys("?", "theory", "tab", "circle", "/", "reveal", "esc", "menu"))
	return b.String()
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (m Model) quizBoard() board {
	qz := m.quiz
	q := qz.q
	bd := board{fretMax: q.FretMax, window: q.Window, cells: map[Pos]cell{}}
	if bd.fretMax < 12 {
		bd.fretMax = 12
	}
	for p, mk := range q.Marks {
		st := map[drill.MarkKind]cell{
			drill.MarkDot:    {mk.Label, sDot},
			drill.MarkRoot:   {mk.Label, sRoot},
			drill.MarkTarget: {mk.Label, sTarget},
			drill.MarkPrev:   {mk.Label, sPrev},
		}[mk.Kind]
		bd.cells[p] = st
	}
	if !qz.answered {
		for _, p := range qz.sel {
			if q.FindAll { // already checked, so it's a find
				bd.cells[p] = cell{q.SolutionLabels[p], sGood}
			} else {
				bd.cells[p] = cell{"●", sSel}
			}
		}
		if q.Kind == drill.AnswerPos {
			c := qz.cursor
			bd.cursor = &c
		}
		return bd
	}
	// reveal
	for _, p := range q.Solution {
		lbl := q.SolutionLabels[p]
		if lbl == "" {
			lbl = "●"
		}
		bd.cells[p] = cell{lbl, sGood}
	}
	for _, p := range qz.sel {
		if qz.ok {
			lbl := q.SolutionLabels[p]
			if lbl == "" {
				lbl = "●"
			}
			bd.cells[p] = cell{lbl, sGood}
		} else if !slices.Contains(q.Solution, p) {
			bd.cells[p] = cell{"✗", sBad}
		}
	}
	return bd
}

// --- theory & circle ---

func (m Model) viewTheory() string {
	lines := strings.Split(m.theoryText, "\n")
	page := max(5, m.bodyHeight())
	end := min(len(lines), m.scroll+page)
	var b strings.Builder
	b.WriteString(sTitle.Render("Theory · "+m.theoryTitle) + "\n")
	for _, l := range lines[m.scroll:end] {
		if l != "" && l == strings.ToUpper(l) && strings.ContainsAny(l, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
			b.WriteString(sBold.Render(l) + "\n")
		} else {
			b.WriteString(l + "\n")
		}
	}
	more := ""
	if end < len(lines) {
		more = fmt.Sprintf(" (%d%%)", end*100/len(lines))
	}
	b.WriteString(keys("w/s", "scroll"+more, "tab", "circle", "esc", "back"))
	return b.String()
}

func (m Model) viewCircle() string {
	var focus *drill.CircleFocus
	if m.back == scrQuiz && m.quiz != nil {
		focus = m.quiz.q.Circle
	}
	var b strings.Builder
	b.WriteString(sTitle.Render("Circle of fifths") + sDim.Render("  majors outside, relative minors inside") + "\n\n")
	b.WriteString(renderCircle(focus) + "\n\n")
	if focus != nil {
		b.WriteString(sStrong.Render(" key/root ") + " " + sHilite.Render("current chord") + "\n")
	}
	b.WriteString(keys("?", "theory", "esc", "back"))
	return b.String()
}

// --- stats ---

func (m Model) viewStats() string {
	var b strings.Builder
	b.WriteString(sTitle.Render("Stats") + sDim.Render("  note names heatmap") + "\n\n")
	bd := board{fretMax: 12, cells: map[Pos]cell{}}
	for s := 1; s <= 6; s++ {
		for f := 0; f <= 12; f++ {
			p := Pos{S: s, F: f}
			c := m.store.Get(fmt.Sprintf("note:s%df%d", s, f))
			name := theory.PitchAt(p).Sharp()
			switch {
			case c == nil:
				bd.cells[p] = cell{"·", sDim}
			case c.Mastered:
				bd.cells[p] = cell{name, sGood}
			case c.Accuracy() >= 0.5:
				bd.cells[p] = cell{name, sWarn}
			default:
				bd.cells[p] = cell{name, sBad}
			}
		}
	}
	b.WriteString(bd.render() + "\n")
	b.WriteString(sGood.Render("known") + "  " + sWarn.Render("learning") + "  " + sBad.Render("weak") + "  " + sDim.Render("· not seen yet") + "\n\n")

	for _, md := range m.mods {
		cards := md.Cards()
		seen, mastered := m.store.Summary(cards)
		att, cor := 0, 0
		for _, c := range cards {
			if st := m.store.Get(c); st != nil {
				att += st.Attempts
				cor += st.Correct
			}
		}
		acc := "   –"
		if att > 0 {
			acc = fmt.Sprintf("%3d%%", cor*100/att)
		}
		fmt.Fprintf(&b, "%-18s %s %s %s\n", md.Name(), bar(mastered, len(cards), 16),
			sDim.Render(fmt.Sprintf("%3d/%-3d known, %3d seen", mastered, len(cards), seen)),
			sMuted.Render(acc+" accuracy"))
	}
	b.WriteString("\n" + keys("esc", "back"))
	return b.String()
}
