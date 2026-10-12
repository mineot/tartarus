package views

import (
	"tartarus/views/themes"

	"charm.land/lipgloss/v2"
)

func renderLogo() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(themes.PrimaryColor)).
		Foreground(lipgloss.Color(themes.SecondaryColor))

	version := lipgloss.NewStyle().
		Foreground(lipgloss.Color(themes.PrimaryColor))

	return title.Render(" Tartarus ") + version.Render(" - 2.0.0")
}
