package pageutils

import (
	"strings"
	"tartarus/views/pages"
	"tartarus/views/themes"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// RenderHelp combines a component's bindings and page shortcuts without changing them.
func RenderHelp(bindings []key.Binding, shortcuts pages.Shortcuts, width int) string {
	combined := make([]key.Binding, 0, len(bindings)+len(shortcuts))
	combined = append(combined, bindings...)
	combined = append(combined, shortcuts.Bindings()...)

	h := help.New()
	h.Styles = themes.Help(true)
	if width <= 0 {
		return h.ShortHelpView(combined)
	}

	var lines []string
	line := ""
	separator := h.Styles.ShortSeparator.Inline(true).Render(h.ShortSeparator)
	for _, binding := range combined {
		if !binding.Enabled() {
			continue
		}
		item := h.ShortHelpView([]key.Binding{binding})
		if line != "" && lipgloss.Width(line+separator+item) > width {
			lines = append(lines, line)
			line = ""
		}
		if line != "" {
			line += separator
		}
		if lipgloss.Width(item) > width {
			item = ansi.Wrap(item, width, "")
		}
		line += item
	}
	if line != "" {
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
