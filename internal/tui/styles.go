package tui

import "github.com/charmbracelet/lipgloss"

var (
	styleSelected = lipgloss.NewStyle().Background(lipgloss.Color("236"))
	styleStatus   = lipgloss.NewStyle().Faint(true).Italic(true).Foreground(lipgloss.Color("245"))
	styleLoading  = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("11"))
	styleErrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	styleSep      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	stylePreview  = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
)
