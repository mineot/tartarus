package pages

import (
	"tartarus/views/themes"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

func NewForm(groups ...*huh.Group) *huh.Form {
	return huh.NewForm(groups...).
		WithTheme(huh.ThemeFunc(themes.Form)).
		WithShowHelp(false)
}

func UpdateForm(form *huh.Form, msg tea.Msg) (*huh.Form, tea.Cmd, bool) {
	if form.State != huh.StateNormal {
		return form, nil, false
	}

	updated, cmd := form.Update(msg)
	if next, ok := updated.(*huh.Form); ok {
		form = next
	}

	return form, cmd, form.State == huh.StateCompleted
}

func FormHelp(form *huh.Form, shortcuts Shortcuts) string {
	var bindings []key.Binding
	if form.State == huh.StateNormal {
		bindings = form.KeyBinds()
	}
	return RenderHelp(bindings, shortcuts)
}
