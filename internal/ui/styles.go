package ui

import "github.com/charmbracelet/lipgloss"

// ANSI colors so the app follows the terminal's own theme.
var (
	sDim     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	sBold    = lipgloss.NewStyle().Bold(true)
	sTitle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	sDot     = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	sRoot    = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	sTarget  = lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true)
	sPrev    = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Bold(true)
	sSel     = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
	sGood    = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	sBad     = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	sWarn    = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	sCursor  = lipgloss.NewStyle().Reverse(true)
	sWindow  = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
	sInlay   = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	sStrong  = lipgloss.NewStyle().Reverse(true).Bold(true)
	sHilite  = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true).Underline(true)
	sMuted   = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	sPlain   = lipgloss.NewStyle()
	sKeyHint = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
)
