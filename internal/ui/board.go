package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"

	"github.com/domakas/lrn-fret/internal/theory"
)

type Pos = theory.Pos

type cell struct {
	label string
	style lipgloss.Style
}

type board struct {
	fretMax int
	window  [2]int // highlighted fret window, [0,0] = none
	cells   map[Pos]cell
	cursor  *Pos
}

var inlays = map[int]bool{3: true, 5: true, 7: true, 9: true, 12: true, 15: true, 17: true}

// render draws the neck with high e on top:
//
//	    0   1   2   3
//	e  ─?─║───│───│───│
func (b board) render() string {
	var sb strings.Builder
	hasWindow := b.window != [2]int{}

	// fret numbers
	sb.WriteString("   ")
	for f := 0; f <= b.fretMax; f++ {
		num := center(fmt.Sprint(f), 3, " ")
		st := sDim
		if inlays[f] {
			st = sInlay
		}
		if hasWindow && f >= b.window[0] && f <= b.window[1] {
			st = sWindow
		}
		sb.WriteString(st.Render(num))
		sb.WriteString(" ")
	}
	sb.WriteString("\n")

	for s := 1; s <= 6; s++ {
		sb.WriteString(sMuted.Render(fmt.Sprintf("%2s ", theory.StringNames[s])))
		for f := 0; f <= b.fretMax; f++ {
			p := Pos{S: s, F: f}
			c, marked := b.cells[p]
			var txt string
			if marked {
				txt = c.style.Render(center(c.label, 3, "─"))
			} else {
				txt = sDim.Render("───")
			}
			if b.cursor != nil && *b.cursor == p {
				lbl := "   "
				st := sCursor
				if marked {
					lbl = center(c.label, 3, " ")
					st = c.style.Reverse(true)
				}
				txt = st.Render(lbl)
			}
			sb.WriteString(txt)
			if f == 0 {
				sb.WriteString(sMuted.Render("║"))
			} else {
				sb.WriteString(sDim.Render("│"))
			}
		}
		sb.WriteString("\n")
	}

	// inlay dots under the low E, like the side of a real neck
	sb.WriteString("   ")
	for f := 0; f <= b.fretMax; f++ {
		dot := "   "
		switch {
		case f%12 == 0 && f > 0:
			dot = " : "
		case inlays[f]:
			dot = " • "
		}
		sb.WriteString(sInlay.Render(dot) + " ")
	}
	return strings.TrimRight(sb.String(), "\n ")
}

// center pads a label to width w with pad.
func center(s string, w int, pad string) string {
	n := utf8.RuneCountInString(s)
	if n >= w {
		return s
	}
	left := (w - n) / 2
	return strings.Repeat(pad, left) + s + strings.Repeat(pad, w-n-left)
}
