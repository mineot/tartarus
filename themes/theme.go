package themes

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

const (
	primaryColor   string = "#90a1b9"
	secondaryColor string = "#0f172b"
	tertiaryColor  string = "250"
	assetColor     string = "#00d3f2"
	errorColor     string = "#e7000b"
)

func GetPrimaryText(value string, bold bool) string {
	style := lipgloss.NewStyle().Bold(bold).Foreground(lipgloss.Color(primaryColor))
	return style.Render(value)
}

func GetSecondaryText(value string, bold bool) string {
	style := lipgloss.NewStyle().Bold(bold).Foreground(lipgloss.Color(secondaryColor))
	return style.Render(value)
}

func GetTertiaryText(value string, bold bool) string {
	style := lipgloss.NewStyle().Bold(bold).Foreground(lipgloss.Color(tertiaryColor))
	return style.Render(value)
}

func GetAssetText(value string, bold bool) string {
	style := lipgloss.NewStyle().Bold(bold).Foreground(lipgloss.Color(assetColor))
	return style.Render(value)
}

func GetErrorText(value string, bold bool) string {
	style := lipgloss.NewStyle().Bold(bold).Foreground(lipgloss.Color(errorColor))
	return style.Render(value)
}

func RenderLogo() {
	title := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(primaryColor)).
		Foreground(lipgloss.Color(secondaryColor))

	version := lipgloss.NewStyle().
		Foreground(lipgloss.Color(primaryColor))

	s := strings.Builder{}
	s.WriteString("\n")
	s.WriteString(title.Render(" Tartarus "))
	s.WriteString(version.Render(" - 2.0.0"))
	s.WriteString("\n")
	fmt.Println(s.String())
}
