package ui

import (
	"fmt"
	"io"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// dbg is the debug log, nil unless turned on with LRN_FRET_DEBUG=1.
var dbg *log.Logger

// slow is the threshold above which an update or render is flagged.
const slow = 30 * time.Millisecond

func SetDebugLog(w io.Writer) {
	dbg = log.New(w, "", log.Ltime|log.Lmicroseconds)
}

func debugf(format string, args ...any) {
	if dbg != nil {
		dbg.Printf(format, args...)
	}
}

var screenNames = map[screen]string{scrMenu: "menu", scrQuiz: "quiz", scrTheory: "theory", scrCircle: "circle", scrStats: "stats"}

// logUpdate records one message and the state it left the app in. Timer
// ticks are only logged when they're slow, or they'd flood the file.
func logUpdate(msg tea.Msg, m Model, took time.Duration) {
	var what string
	switch msg := msg.(type) {
	case tickMsg:
		if took < slow {
			return
		}
		what = fmt.Sprintf("tick seq=%d", msg.seq)
	case advanceMsg:
		what = fmt.Sprintf("auto-advance seq=%d", msg.seq)
	case tea.KeyMsg:
		what = fmt.Sprintf("key %q", msg.String())
	case tea.FocusMsg:
		what = "focus"
	case tea.BlurMsg:
		what = "blur"
	case tea.WindowSizeMsg:
		what = fmt.Sprintf("resize %dx%d", msg.Width, msg.Height)
	default:
		what = fmt.Sprintf("%T", msg)
	}
	state := screenNames[m.scr]
	if m.blurred {
		state += " (paused)"
	}
	if qz := m.quiz; qz != nil && m.scr == scrQuiz {
		state += fmt.Sprintf(" card=%s seq=%d answered=%v input=%q", qz.q.Card, qz.seq, qz.answered, qz.input)
		if m.pendingAdvance {
			state += " pending-advance"
		}
	}
	flag := ""
	if took >= slow {
		flag = "  SLOW"
	}
	debugf("%-22s → %s  [%v]%s", what, state, took.Round(time.Microsecond), flag)
}
