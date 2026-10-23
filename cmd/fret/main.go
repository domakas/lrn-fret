// Command fret is a terminal fretboard trainer.
//
//	fret            menu
//	fret quick      10 questions from your weakest cards
//	fret <drill>    jump into a drill: notes find intervals spell triads scales fifths changes
//	fret stats      heatmap and progress
//	fret circle     circle of fifths
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/domakas/lrn-fret/internal/progress"
	"github.com/domakas/lrn-fret/internal/ui"
)

func main() {
	start := ""
	if len(os.Args) > 1 {
		start = os.Args[1]
	}
	if start == "-h" || start == "--help" || start == "help" {
		fmt.Println(`fret — fretboard trainer

  fret            menu
  fret quick      10 questions from your weakest cards
  fret <drill>    notes | find | intervals | spell | triads | scales | fifths | changes
  fret stats      heatmap and progress
  fret circle     circle of fifths

Progress: ` + progress.DefaultPath() + ` (set LRN_FRET_HOME to move it)`)
		return
	}
	store, err := progress.Load(progress.DefaultPath())
	if err != nil {
		fmt.Fprintln(os.Stderr, "loading progress:", err)
		os.Exit(1)
	}
	m, err := ui.New(store, start)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if _, err := tea.NewProgram(m, tea.WithAltScreen(), tea.WithReportFocus()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
