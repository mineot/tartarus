package pagehelpers

import (
	"fmt"
	"strings"
	"tartarus/views/pages"
	pageutils "tartarus/views/pages/utils"
	"tartarus/views/themes"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type FetchPage func(int) tea.Cmd

type ListLoadedMsg struct {
	Err     error
	Page    int
	Request uint64
	Search  string
	Size    int
	Total   int
}

type ListModel struct {
	DisplayedSearch string
	DisplayedSize   int
	Err             error
	List            list.Model
	Loading         bool
	Next            key.Binding
	Page            int
	PendingPage     int
	Previous        key.Binding
	Request         uint64
	Retry           key.Binding
	Search          string
	SearchEditing   bool
	SearchInput     textinput.Model
	SearchKey       key.Binding
	Shortcuts       pages.Shortcuts
	Size            int
	Sized           bool
	Started         bool
	Total           int
}

func (lm *ListModel) SearchUpdate(msg tea.Msg, m pages.Page, fetchPage FetchPage) (pages.Page, tea.Cmd) {
	if pressed, ok := msg.(tea.KeyPressMsg); ok {
		if lm.SearchEditing {
			switch pressed.String() {
			case "enter":
				lm.Search = strings.TrimSpace(lm.SearchInput.Value())
				lm.SearchInput.Blur()
				lm.SearchEditing = false

				if lm.Sized {
					return m, fetchPage(0)
				}

				return m, nil

			case "esc":
				lm.SearchInput.SetValue(lm.Search)
				lm.SearchInput.Blur()
				lm.SearchEditing = false
				return m, nil

			default:
				var cmd tea.Cmd
				lm.SearchInput, cmd = lm.SearchInput.Update(msg)
				return m, cmd
			}
		}

		if key.Matches(pressed, lm.SearchKey) {
			lm.SearchInput.SetValue(lm.Search)
			lm.SearchEditing = true
			return m, lm.SearchInput.Focus()
		}

		if pressed.String() == "esc" && (lm.Search != "" || lm.DisplayedSearch != "") {
			lm.Search = ""
			lm.SearchInput.SetValue("")

			if lm.Sized {
				return m, fetchPage(0)
			}

			return m, nil
		}
	}

	return nil, nil
}

func (lm *ListModel) ShortcutsUpdate(msg tea.Msg, m pages.Page) (pages.Page, tea.Cmd) {
	if cmd, handled := lm.Shortcuts.Handle(msg, lm.SearchEditing); handled {
		return m, cmd
	}

	return nil, nil
}

func (lm *ListModel) WindowSizeUpdate(msg tea.WindowSizeMsg, m pages.Page, fetchPage FetchPage) (pages.Page, tea.Cmd) {
	height := max(2, msg.Height-8)
	size := min(5, height/2)
	oldSize := lm.Size

	lm.SearchInput.SetWidth(max(1, msg.Width-8))
	lm.List.SetSize(max(1, msg.Width), height)
	lm.Size, lm.Sized = size, true

	if lm.Started && oldSize != size {
		page := lm.Page * lm.DisplayedSize / size

		if lm.Search != lm.DisplayedSearch {
			page = 0
		}

		return m, fetchPage(page)
	}

	return m, nil
}

func (lm *ListModel) KeyPressUpdate(msg tea.KeyPressMsg, m pages.Page, fetchPage FetchPage) (pages.Page, tea.Cmd) {
	if lm.Loading || !lm.Sized {
		return m, nil
	}

	switch {
	case key.Matches(msg, lm.Retry):
		if lm.Err != nil {
			return m, fetchPage(lm.PendingPage)
		}

	case key.Matches(msg, lm.Previous):
		if lm.Err == nil && lm.Page > 0 {
			return m, fetchPage(lm.Page - 1)
		}

	case key.Matches(msg, lm.Next):
		if lm.Err == nil && (lm.Page+1)*lm.Size < lm.Total {
			return m, fetchPage(lm.Page + 1)
		}

	case key.Matches(msg, lm.List.KeyMap.CursorUp):
		lm.List.CursorUp()
		return m, nil

	case key.Matches(msg, lm.List.KeyMap.CursorDown):
		lm.List.CursorDown()
		return m, nil
	}

	return nil, nil
}

func (lm *ListModel) PageBoundsUpdate(msg ListLoadedMsg, m pages.Page, fetchPage FetchPage) (pages.Page, tea.Cmd) {
	if msg.Page > 0 && msg.Page*msg.Size >= msg.Total {
		return m, fetchPage(max(0, (msg.Total-1)/msg.Size))
	}

	return nil, nil
}

func (lm *ListModel) PaginationView(itemLabel string) string {
	return fmt.Sprintf(
		"Page %d of %d · %d %s",
		lm.Page+1,
		1+(lm.Total-1)/max(1, lm.DisplayedSize),
		lm.Total,
		itemLabel,
	)
}

func (lm *ListModel) HelpBindings() []key.Binding {
	return []key.Binding{
		lm.List.KeyMap.CursorUp,
		lm.List.KeyMap.CursorDown,
		lm.Previous,
		lm.Next,
		lm.SearchKey,
	}
}

func (lm *ListModel) SearchView() string {
	return "Search: " + lm.DisplayedSearch
}

func (lm *ListModel) SearchHelpView(help string, bindings []key.Binding) string {
	if lm.Search == "" && lm.DisplayedSearch == "" {
		return help
	}

	shortcuts := append(pages.Shortcuts(nil), lm.Shortcuts...)

	clearSearch := key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "clear search"),
	)

	replaced := false

	for i := range shortcuts {
		if key.Matches(tea.KeyPressMsg{Code: tea.KeyEscape}, shortcuts[i].Binding) {
			shortcuts[i].Binding = clearSearch
			replaced = true
		}
	}

	if !replaced {
		bindings = append(append([]key.Binding(nil), bindings...), clearSearch)
	}

	return pageutils.RenderHelp(bindings, shortcuts)
}

func (lm *ListModel) SearchStatusView() string {
	if lm.Loading && lm.Search != "" {
		return fmt.Sprintf("Searching for %q...", lm.Search)
	}

	return ""
}

func (lm *ListModel) SearchEditingView(search string, help string) (string, string) {
	if lm.SearchEditing {
		return lm.SearchInput.View(),
			pageutils.RenderHelp([]key.Binding{
				key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "search")),
				key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
			}, nil)
	}

	return search, help
}

func NewList(searchPlaceholder string, shortcuts pages.Shortcuts) ListModel {
	delegate := list.NewDefaultDelegate()
	delegate.Styles = themes.ListItems(true)
	delegate.SetSpacing(0)

	items := list.New(nil, delegate, 0, 0)
	items.SetShowTitle(false)
	items.SetShowStatusBar(false)
	items.SetShowPagination(false)
	items.SetShowHelp(false)
	items.SetFilteringEnabled(false)
	items.DisableQuitKeybindings()

	input := textinput.New()
	input.Prompt = "Search: "
	input.Placeholder = searchPlaceholder

	return ListModel{
		SearchInput: input,
		SearchKey:   key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		List:        items,
		Shortcuts:   shortcuts,
		Previous:    key.NewBinding(key.WithKeys("left", "pgup"), key.WithHelp("←/pgup", "previous page")),
		Next:        key.NewBinding(key.WithKeys("right", "pgdown"), key.WithHelp("→/pgdn", "next page")),
		Retry:       key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "retry")),
	}
}
