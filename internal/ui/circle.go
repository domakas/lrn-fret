package ui

import (
	"math"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/domakas/lrn-fret/internal/drill"
	"github.com/domakas/lrn-fret/internal/theory"
)

const (
	circW, circH = 48, 15
	circCX       = 24
	circCY       = 7
)

// renderCircle draws the circle of fifths with majors outside and relative
// minors inside, highlighting whatever the current question is about.
func renderCircle(focus *drill.CircleFocus) string {
	grid := make([][]rune, circH)
	styles := make([][]*lipgloss.Style, circH)
	for y := range grid {
		grid[y] = []rune(strings.Repeat(" ", circW))
		styles[y] = make([]*lipgloss.Style, circW)
	}
	put := func(x, y int, s string, st lipgloss.Style) {
		x -= len([]rune(s)) / 2
		for i, r := range []rune(s) {
			if x+i >= 0 && x+i < circW && y >= 0 && y < circH {
				grid[y][x+i] = r
				styles[y][x+i] = &st
			}
		}
	}
	has := func(xs []theory.PC, n string) bool {
		return slices.Contains(xs, theory.MustNote(n).PC)
	}

	for i := range theory.CircleMajors {
		a := float64(i) * math.Pi / 6
		maj, min := theory.CircleMajors[i], theory.CircleMinors[i]+"m"

		st := sBold
		if focus != nil {
			switch {
			case has(focus.Strong, maj):
				st = sStrong
			case has(focus.Major, maj):
				st = sHilite
			}
		}
		ox := circCX + int(math.Round(20*math.Sin(a)))
		oy := circCY - int(math.Round(7*math.Cos(a)))
		put(ox, oy, maj, st)

		st = sMuted
		if focus != nil && has(focus.Minor, theory.CircleMinors[i]) {
			st = sHilite
		}
		ix := circCX + int(math.Round(11*math.Sin(a)))
		iy := circCY - int(math.Round(3.8*math.Cos(a)))
		put(ix, iy, min, st)
	}
	put(circCX, circCY, "↻ +1♯   +1♭ ↺", sDim)

	var sb strings.Builder
	for y := range grid {
		for x := 0; x < circW; {
			st := styles[y][x]
			j := x
			for j < circW && styles[y][j] == st {
				j++
			}
			seg := string(grid[y][x:j])
			if st != nil {
				seg = st.Render(seg)
			}
			sb.WriteString(seg)
			x = j
		}
		sb.WriteString("\n")
	}

	var sigs []string
	for i, k := range theory.CircleMajors {
		sigs = append(sigs, k+" "+theory.Signature(i))
	}
	sb.WriteString(sDim.Render(strings.Join(sigs[:7], "  ")) + "\n")
	sb.WriteString(sDim.Render(strings.Join(sigs[7:], "  ")))
	return sb.String()
}
