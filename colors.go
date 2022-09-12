package main

import "github.com/charmbracelet/lipgloss"

var (
	Green       = lipgloss.Color("#01a252")
	Yellow      = lipgloss.Color("#fded02")
	Blue        = lipgloss.Color("#01a0e4")
	Red         = lipgloss.Color("#db2d20")
	BrightWhite = lipgloss.Color("#f7f7f7")

	SelectBackground = lipgloss.Color("#303030")
	BorderColor      = BrightWhite
)

func Colored(color lipgloss.TerminalColor) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(color)
}
