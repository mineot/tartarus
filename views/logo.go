package views

import (
	"charm.land/lipgloss/v2"
)

func renderLogo() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color("#90a1b9")).
		Foreground(lipgloss.Color("#0f172b"))

	version := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#90a1b9"))

	return title.Render(" Tartarus ") + version.Render(" - 2.0.0")
}
