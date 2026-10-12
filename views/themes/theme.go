package themes

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/list"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

const (
	PrimaryColor   string = "#90a1b9"
	SecondaryColor string = "#0f172b"
	AssetColor     string = "#00d3f2"
)

// Help returns readable shortcut styles using the Tartarus slate palette.
func Help(isDark bool) help.Styles {
	lightDark := lipgloss.LightDark(isDark)
	keys := lipgloss.NewStyle().Foreground(lightDark(lipgloss.Color("#334155"), lipgloss.Color("#cbd5e1")))
	descriptions := lipgloss.NewStyle().Foreground(lightDark(lipgloss.Color("#475569"), lipgloss.Color("#90a1b9")))
	separators := lipgloss.NewStyle().Foreground(lightDark(lipgloss.Color("#64748b"), lipgloss.Color("#90a1b9")))

	return help.Styles{
		ShortKey:       keys,
		ShortDesc:      descriptions,
		ShortSeparator: separators,
		FullKey:        keys,
		FullDesc:       descriptions,
		FullSeparator:  separators,
		Ellipsis:       separators,
	}
}

// Form returns shared form styles using the Tartarus slate palette.
func Form(isDark bool) *huh.Styles {
	t := huh.ThemeCharm(isDark)

	for _, field := range []*huh.FieldStyles{&t.Focused, &t.Blurred} {
		field.Title = field.Title.Foreground(lipgloss.Color(PrimaryColor))
		field.SelectSelector = field.SelectSelector.Foreground(lipgloss.Color(PrimaryColor))
		field.SelectedOption = field.SelectedOption.Foreground(lipgloss.Color(AssetColor))
		field.TextInput.Prompt = field.TextInput.Cursor.Foreground(lipgloss.Color(PrimaryColor))
		field.TextInput.Cursor = field.TextInput.Cursor.Foreground(lipgloss.Color(AssetColor))
	}

	return t
}

// ListItems returns list item styles using the shared form palette.
func ListItems(isDark bool) list.DefaultItemStyles {
	form := Form(isDark)
	styles := list.NewDefaultItemStyles(isDark)

	selector := form.Focused.SelectSelector.GetForeground()
	selected := form.Focused.SelectedOption.GetForeground()

	styles.SelectedTitle = styles.SelectedTitle.
		Foreground(selected).
		BorderForeground(selector)
	styles.SelectedDesc = styles.SelectedDesc.
		Foreground(selector).
		BorderForeground(selector)

	return styles
}
