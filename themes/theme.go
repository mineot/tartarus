package themes

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

const (
	PrimaryColor   string = "#90a1b9"
	SecondaryColor string = "#0f172b"
	TertiaryColor  string = "250"
	AssetColor     string = "#00d3f2"
)

func GetPrimaryText(value string, bold bool) string {
	style := lipgloss.NewStyle().Bold(bold).Foreground(lipgloss.Color(PrimaryColor))
	return style.Render(value)
}

func GetSecondaryText(value string, bold bool) string {
	style := lipgloss.NewStyle().Bold(bold).Foreground(lipgloss.Color(SecondaryColor))
	return style.Render(value)
}

func GetTertiaryText(value string, bold bool) string {
	style := lipgloss.NewStyle().Bold(bold).Foreground(lipgloss.Color(TertiaryColor))
	return style.Render(value)
}

func GetAssetText(value string, bold bool) string {
	style := lipgloss.NewStyle().Bold(bold).Foreground(lipgloss.Color(AssetColor))
	return style.Render(value)
}

func RenderLogo() {
	title := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color(PrimaryColor)).
		Foreground(lipgloss.Color(SecondaryColor))

	version := lipgloss.NewStyle().
		Foreground(lipgloss.Color(PrimaryColor))

	s := strings.Builder{}
	s.WriteString("\n")
	s.WriteString(title.Render(" Tartarus "))
	s.WriteString(version.Render(" - 2.0.0"))
	s.WriteString("\n")
	fmt.Println(s.String())
}
