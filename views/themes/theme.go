package themes

import (
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

// Form returns shared form styles using the Tartarus slate palette.
func Form(isDark bool) *huh.Styles {
	t := huh.ThemeCharm(isDark)

	for _, field := range []*huh.FieldStyles{&t.Focused, &t.Blurred} {
		field.Title = field.Title.Foreground(lipgloss.Color("#90a1b9"))
		field.SelectSelector = field.SelectSelector.Foreground(lipgloss.Color("#90a1b9"))
		field.SelectedOption = field.SelectedOption.Foreground(lipgloss.Color("#00d3f2"))
	}

	return t
}
