package components

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"tartarus/themes"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

// FormOpts configures one group of native Huh fields.
// Use Key to identify each value, Title for its label, and Validate for custom
// validation. Fields are owned by this run and must not be shared concurrently.
type FormOpts struct {
	Title       string
	Description string
	Fields      []huh.Field
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
// *huh.Form, whose getters stay valid after submission.
func Form(ctx context.Context, opts FormOpts) (FormResult, error) {
	if ctx == nil {
		return nil, fmt.Errorf("components: form: context is required")
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("components: form: %w", err)
	}

	form, err := newForm(opts)

	if err != nil {
		return nil, err
	}

	// Run Huh's model directly so terminal startup errors are handled before
	// inspecting the final model. Huh still owns field navigation and validation.
	form.SubmitCmd = tea.Quit
	form.CancelCmd = tea.Interrupt

	_, err = tea.NewProgram(formModel{Model: form},
		tea.WithContext(ctx),
		tea.WithOutput(os.Stderr),
	).Run()

	if ctxErr := ctx.Err(); ctxErr != nil {
		return nil, fmt.Errorf("components: running form: %w", ctxErr)
	}

	if form.State == huh.StateAborted || errors.Is(err, tea.ErrInterrupted) {
		return nil, fmt.Errorf("components: running form: %w", huh.ErrUserAborted)
	}

	if errors.Is(err, tea.ErrProgramKilled) {
		return nil, fmt.Errorf("components: running form: program was killed: %w", err)
	}

	if err != nil {
		return nil, fmt.Errorf("components: running form: %w", err)
	}

	if form.State != huh.StateCompleted {
		return nil, fmt.Errorf("components: running form: form ended without submission")
	}

	return form, nil
}

// Huh fields still render strings; adapt their model to Bubble Tea v2 views.
type formModel struct {
	huh.Model
}

func (m formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)
	return m, cmd
}

func (m formModel) View() tea.View {
	view := tea.NewView(m.Model.View())
	view.ReportFocus = true
	return view
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
		WithShowHelp(true).
		WithShowErrors(true), nil
}

func formTheme(isDark bool) *huh.Styles {
	theme := huh.ThemeCharm(isDark)
	primary := lipgloss.Color(themes.PrimaryColor)
	secondary := lipgloss.Color(themes.SecondaryColor)
	tertiary := lipgloss.Color(themes.TertiaryColor)
	asset := lipgloss.Color(themes.AssetColor)

	theme.Group.Title = theme.Group.Title.Foreground(primary).Bold(true)
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
	theme.Help.ShortKey = theme.Help.ShortKey.Foreground(asset)
	theme.Help.FullKey = theme.Help.FullKey.Foreground(asset)

	return theme
}
