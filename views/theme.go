package views

import (
	"strings"

	"charm.land/lipgloss/v2"
)

const (
	PrimaryColor   string = "#90a1b9"
	SecondaryColor string = "#0f172b"
	AssetColor     string = "#00d3f2"
)

func RenderLogo(shortDescription string) string {
	title := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(PrimaryColor)).
		Foreground(lipgloss.Color(SecondaryColor))

	version := lipgloss.NewStyle().
		Foreground(lipgloss.Color(PrimaryColor))

	description := lipgloss.NewStyle().
		Foreground(lipgloss.Color(PrimaryColor))

	s := strings.Builder{}
	s.WriteString("\n")
	s.WriteString(title.Render(" Tartarus "))
	s.WriteString(version.Render(" - 2.0.0"))
	s.WriteString("\n\n")
	s.WriteString(description.Render(shortDescription))
	return s.String()
}
