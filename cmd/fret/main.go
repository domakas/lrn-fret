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
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/domakas/lrn-fret/internal/progress"
	"github.com/domakas/lrn-fret/internal/ui"
)

func debugPath() string {
	return filepath.Join(filepath.Dir(progress.DefaultPath()), "debug.log")
}

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

Progress: ` + progress.DefaultPath() + ` (set LRN_FRET_HOME to move it)
Debug log: LRN_FRET_DEBUG=1 fret … writes ` + debugPath())
		return
	}
	if os.Getenv("LRN_FRET_DEBUG") != "" {
		_ = os.MkdirAll(filepath.Dir(debugPath()), 0o755)
		f, err := os.OpenFile(debugPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			fmt.Fprintln(os.Stderr, "debug log:", err)
			os.Exit(1)
		}
		defer f.Close()
		ui.SetDebugLog(f)
		fmt.Fprintf(f, "\n--- fret %s started %s ---\n", start, time.Now().Format(time.DateTime))
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
