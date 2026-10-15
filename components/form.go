package components

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"tartarus/themes"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

// FormShortcut is a custom key binding rendered in the form footer.
type FormShortcut struct {
	// Key is matched against the pressed key, for example "ctrl+s" or "f1".
	Key string
	// Label is shown next to the key in the footer, for example "save".
	Label string
	// Run is called when Key is pressed. focus is the key of the focused field,
	// or "" when no field is focused. Returning quit stops the form; returning
	// a non-nil error stops the form and is returned by Form. A nil Run displays
	// the shortcut without acting on it.
	Run func(focus string) (quit bool, err error)
}

// FormOpts configures one group of native Huh fields.
// Use Key to identify each value, Title for its label, and Validate for custom
// validation. Fields are owned by this run and must not be shared concurrently.
type FormOpts struct {
	Title       string
	Description string
	Fields      []huh.Field
	// Shortcuts adds custom key bindings to the footer. Keys the form or its
	// fields already use are rejected. A printable key that is not otherwise
	// bound shadows typing in text fields.
	Shortcuts []FormShortcut
}

// FormResult exposes Huh's value getters after successful submission.
// Get returns nil for missing keys. Typed getters return the zero value for
// missing keys or mismatched types; use Get for slices and other custom values.
type FormResult interface {
	Get(string) any
	GetString(string) string
	GetBool(string) bool
	GetInt(string) int
}

// Form runs an inline form with Huh's editing, validation and contextual help.
// It returns no result on failure or cancellation. Errors preserve
// huh.ErrUserAborted and context cancellation through errors.Is.
// Fields bound with Value may update caller-owned variables before submission.
//
// Cancellation is Ctrl+C only, Huh's default quit binding. Escape and q cannot
// be added without colliding with field-level bindings (Escape clears filters
// and closes pickers; q would be swallowed before text inputs see it).
//
// The form renders to stderr, keeping stdout free for the command's output, and
// runs inline without alternate-screen mode. The returned FormResult is the live
// *huh.Form, whose getters stay valid after submission. The footer is a single
// line rendered through Huh's help model: the focused field's keybinds first,
// then the configured Shortcuts, then Ctrl+C quit last.
func Form(ctx context.Context, opts FormOpts) (FormResult, error) {
	if ctx == nil {
		return nil, fmt.Errorf("components: form: context is required")
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("components: form: %w", err)
	}

	m, err := newFormModel(opts)

	if err != nil {
		return nil, err
	}

	// Run Huh's model directly so terminal startup errors are handled before
	// inspecting the final model. Huh still owns field navigation and validation.
	m.form.SubmitCmd = tea.Quit
	m.form.CancelCmd = tea.Interrupt

	final, err := tea.NewProgram(m,
		tea.WithContext(ctx),
		tea.WithOutput(os.Stderr),
	).Run()

	if ctxErr := ctx.Err(); ctxErr != nil {
		return nil, fmt.Errorf("components: running form: %w", ctxErr)
	}

	if fm, ok := final.(formModel); ok && fm.runErr != nil {
		return nil, fmt.Errorf("components: running form: %w", fm.runErr)
	}

	aborted := m.form.State == huh.StateAborted || errors.Is(err, tea.ErrInterrupted)

	if fm, ok := final.(formModel); ok && fm.aborted {
		aborted = true
	}

	if aborted {
		return nil, fmt.Errorf("components: running form: %w", huh.ErrUserAborted)
	}

	if errors.Is(err, tea.ErrProgramKilled) {
		return nil, fmt.Errorf("components: running form: program was killed: %w", err)
	}

	if err != nil {
		return nil, fmt.Errorf("components: running form: %w", err)
	}

	if m.form.State != huh.StateCompleted {
		return nil, fmt.Errorf("components: running form: form ended without submission")
	}

	return m.form, nil
}

// Huh fields still render strings; adapt their model to Bubble Tea v2 views.
type formModel struct {
	huh.Model
	form      *huh.Form
	shortcuts map[string]FormShortcut
	order     []FormShortcut
	runErr    error
	aborted   bool
}

func (m formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if press, ok := msg.(tea.KeyPressMsg); ok {
		if shortcut, exists := m.shortcuts[press.String()]; exists {
			if shortcut.Run == nil {
				return m, nil
			}

			focus := ""

			if field := m.form.GetFocusedField(); field != nil {
				focus = field.GetKey()
			}

			quit, err := shortcut.Run(focus)

			if err != nil {
				m.runErr = err
				return m, tea.Quit
			}

			if quit {
				m.aborted = true
				return m, tea.Quit
			}

			return m, nil
		}
	}

	var cmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)
	return m, cmd
}

func (m formModel) View() tea.View {
	content := m.Model.View()

	if footer := m.footerView(); footer != "" {
		content = strings.Join([]string{content, "", footer}, "\n")
	}

	view := tea.NewView(content)
	view.ReportFocus = true
	return view
}

func (m formModel) footerView() string {
	bindings := append([]key.Binding{}, m.form.KeyBinds()...)

	for _, shortcut := range m.order {
		bindings = append(bindings, key.NewBinding(
			key.WithKeys(shortcut.Key),
			key.WithHelp(shortcut.Key, shortcut.Label),
		))
	}

	bindings = append(bindings, key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	))

	return m.form.Help().ShortHelpView(bindings)
}

func newFormModel(opts FormOpts) (formModel, error) {
	form, err := newForm(opts)

	if err != nil {
		return formModel{}, err
	}

	shortcuts, err := newShortcuts(opts.Shortcuts, opts.Fields)

	if err != nil {
		return formModel{}, err
	}

	return formModel{
		Model:     form,
		form:      form,
		shortcuts: shortcuts,
		order:     opts.Shortcuts,
	}, nil
}

func newShortcuts(shortcuts []FormShortcut, fields []huh.Field) (map[string]FormShortcut, error) {
	reserved := formReservedKeys(fields)
	byKey := make(map[string]FormShortcut, len(shortcuts))

	for i, shortcut := range shortcuts {
		if shortcut.Key == "" {
			return nil, fmt.Errorf("components: form: shortcut %d: key is required", i+1)
		}

		if shortcut.Label == "" {
			return nil, fmt.Errorf("components: form: shortcut %d: label is required", i+1)
		}

		if _, taken := reserved[shortcut.Key]; taken {
			return nil, fmt.Errorf("components: form: shortcut %d: key %q is reserved", i+1, shortcut.Key)
		}

		if _, duplicate := byKey[shortcut.Key]; duplicate {
			return nil, fmt.Errorf("components: form: shortcut %d: duplicate key %q", i+1, shortcut.Key)
		}

		byKey[shortcut.Key] = shortcut
	}

	return byKey, nil
}

// formReservedKeys returns the keys the form handles itself: Ctrl+C quitting
// plus every key bound by the provided fields, which Huh would otherwise use.
func formReservedKeys(fields []huh.Field) map[string]struct{} {
	reserved := map[string]struct{}{
		"ctrl+c": {},
	}

	for _, field := range fields {
		value := reflect.ValueOf(field)

		if !value.IsValid() || (value.Kind() == reflect.Pointer && value.IsNil()) {
			continue
		}

		for _, binding := range field.KeyBinds() {
			if !binding.Enabled() {
				continue
			}

			for _, k := range binding.Keys() {
				reserved[k] = struct{}{}
			}
		}
	}

	return reserved
}

func newForm(opts FormOpts) (*huh.Form, error) {
	if len(opts.Fields) == 0 {
		return nil, fmt.Errorf("components: form: at least one field is required")
	}

	keys := make(map[string]struct{}, len(opts.Fields))

	for i, field := range opts.Fields {
		value := reflect.ValueOf(field)

		if !value.IsValid() || (value.Kind() == reflect.Pointer && value.IsNil()) {
			return nil, fmt.Errorf("components: form: field %d is nil", i+1)
		}

		key := field.GetKey()

		if strings.TrimSpace(key) == "" {
			if _, note := field.(*huh.Note); note {
				continue
			}

			return nil, fmt.Errorf("components: form: field %d requires a key", i+1)
		}

		if _, exists := keys[key]; exists {
			return nil, fmt.Errorf("components: form: duplicate key %q", key)
		}

		keys[key] = struct{}{}
	}

	group := huh.NewGroup(opts.Fields...).
		Title(opts.Title).
		Description(opts.Description)

	return huh.NewForm(group).
		WithAccessible(false).
		WithTheme(huh.ThemeFunc(formTheme)).
		WithShowHelp(false).
		WithShowErrors(true), nil
}

func formTheme(isDark bool) *huh.Styles {
	theme := huh.ThemeCharm(isDark)
	primary := lipgloss.Color(themes.PrimaryColor)
	secondary := lipgloss.Color(themes.SecondaryColor)
	tertiary := lipgloss.Color(themes.TertiaryColor)
	asset := lipgloss.Color(themes.AssetColor)

	theme.Group.Title = theme.Group.Title.Foreground(primary).Padding(0, 0, 1, 0).Bold(true)
	theme.Group.Description = theme.Group.Description.Foreground(tertiary)

	for _, field := range []*huh.FieldStyles{&theme.Focused, &theme.Blurred} {
		field.Title = field.Title.Foreground(primary)
		field.Description = field.Description.Foreground(tertiary)
		field.NoteTitle = field.NoteTitle.Foreground(primary)
		field.SelectSelector = field.SelectSelector.Foreground(asset)
		field.MultiSelectSelector = field.MultiSelectSelector.Foreground(asset)
		field.SelectedOption = field.SelectedOption.Foreground(asset)
		field.SelectedPrefix = field.SelectedPrefix.Foreground(asset)
		field.NextIndicator = field.NextIndicator.Foreground(asset)
		field.PrevIndicator = field.PrevIndicator.Foreground(asset)
		field.Directory = field.Directory.Foreground(asset)
		field.TextInput.Cursor = field.TextInput.Cursor.Foreground(asset)
		field.TextInput.Prompt = field.TextInput.Prompt.Foreground(asset)
		field.FocusedButton = field.FocusedButton.Background(asset).Foreground(secondary)
		field.Next = field.FocusedButton
	}

	theme.Focused.Base = theme.Focused.Base.BorderForeground(asset)
	theme.Focused.Card = theme.Focused.Base
	theme.Focused.Title = theme.Focused.Title.Foreground(asset)
	theme.Help.ShortKey = theme.Help.ShortKey.Foreground(tertiary)
	theme.Help.FullKey = theme.Help.FullKey.Foreground(tertiary)

	return theme
}
