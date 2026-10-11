package pageutils

import (
	"tartarus/views/pages"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

// RenderHelp combines a component's bindings and page shortcuts without changing them.
func RenderHelp(bindings []key.Binding, shortcuts pages.Shortcuts) string {
	combined := make([]key.Binding, 0, len(bindings)+len(shortcuts))
	combined = append(combined, bindings...)
	combined = append(combined, shortcuts.Bindings()...)

	h := help.New()
	return h.ShortHelpView(combined)
}
