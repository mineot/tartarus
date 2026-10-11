package pagemenus

// import (
// 	"errors"
// 	"strings"
// 	"tartarus/repositories"
// 	"testing"

// 	tea "charm.land/bubbletea/v2"
// )

// func menuForTest(t *testing.T, count int) *MenuCmds {
// 	t.Helper()
// 	m := NewMenuCmds(&repositories.Repos{})
// 	m.loadCommands = func(limit, offset int, search string) ([]repositories.Command, int, error) {
// 		var commands []repositories.Command
// 		for i := offset; i < min(offset+limit, count); i++ {
// 			commands = append(commands, repositories.Command{ID: uint64(i + 1), Name: "build", Description: "Build the project"})
// 		}
// 		return commands, count, nil
// 	}
// 	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
// 	return m
// }

// func TestMenuCmdsPagination(t *testing.T) {
// 	m := menuForTest(t, 23)
// 	m.Update(m.Init()())
// 	if len(m.list.Items()) != 5 || !strings.Contains(m.View(), "Page 1 of 5") {
// 		t.Fatalf("unexpected first page: %s", m.View())
// 	}
// 	item := m.list.Items()[0].(commandListItem)
// 	if item.Title() != "build" || item.Description() != "Build the project" {
// 		t.Fatalf("unexpected item: %+v", item)
// 	}
// 	for _, want := range []int{1, 2, 3, 4} {
// 		_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
// 		if cmd == nil {
// 			t.Fatal("expected page query")
// 		}
// 		_, duplicate := m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
// 		if duplicate != nil {
// 			t.Fatal("queried while loading")
// 		}
// 		m.Update(cmd())
// 		if m.page != want {
// 			t.Fatalf("want page %d, got %d", want, m.page)
// 		}
// 	}
// 	if len(m.list.Items()) != 3 {
// 		t.Fatal("expected partial last page")
// 	}
// 	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
// 	if cmd != nil {
// 		t.Fatal("advanced beyond last page")
// 	}
// 	_, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
// 	m.Update(cmd())
// 	if m.page != 3 {
// 		t.Fatal("did not return to previous page")
// 	}
// }

// func TestMenuCmdsFailedPageCanRetry(t *testing.T) {
// 	m := menuForTest(t, 23)
// 	m.Update(m.Init()())
// 	load := m.loadCommands
// 	m.loadCommands = func(int, int, string) ([]repositories.Command, int, error) {
// 		return nil, 0, errors.New("database unavailable")
// 	}
// 	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
// 	m.Update(cmd())
// 	if m.page != 0 || len(m.list.Items()) != 5 || m.err == nil {
// 		t.Fatal("failed query changed displayed page")
// 	}
// 	m.loadCommands = load
// 	_, cmd = m.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
// 	if cmd == nil {
// 		t.Fatal("expected retry")
// 	}
// 	m.Update(cmd())
// 	if m.page != 1 || m.err != nil {
// 		t.Fatal("retry did not load requested page")
// 	}
// }

// func TestMenuCmdsResizeIgnoresOldResponses(t *testing.T) {
// 	m := menuForTest(t, 23)
// 	old := m.Init()
// 	_, resized := m.Update(tea.WindowSizeMsg{Width: 40, Height: 12})
// 	m.Update(old())
// 	if !m.loading || len(m.list.Items()) != 0 {
// 		t.Fatal("accepted outdated response")
// 	}
// 	m.Update(resized())
// 	if m.size != 2 || len(m.list.Items()) != 2 || m.list.Paginator.TotalPages != 1 {
// 		t.Fatal("resized page does not fit")
// 	}
// 	other := menuForTest(t, 23)
// 	m.Update(other.Init()())
// 	if len(m.list.Items()) != 2 {
// 		t.Fatal("accepted response from another menu")
// 	}
// }

// func TestMenuCmdsEmptyAndShrinkingDatabase(t *testing.T) {
// 	m := menuForTest(t, 0)
// 	m.Update(m.Init()())
// 	if !strings.Contains(m.View(), "No commands found.") {
// 		t.Fatal(m.View())
// 	}
// 	m = menuForTest(t, 23)
// 	m.Update(m.Init()())
// 	m.Update(m.fetch(2)())
// 	m.loadCommands = func(int, int, string) ([]repositories.Command, int, error) { return nil, 0, nil }
// 	_, cmd := m.Update(m.fetch(2)())
// 	if cmd == nil {
// 		t.Fatal("expected correction to first page")
// 	}
// 	m.Update(cmd())
// 	if m.page != 0 || m.total != 0 {
// 		t.Fatal("did not correct empty database page")
// 	}
// }

// func TestMenuCmdsWaitsForDimensions(t *testing.T) {
// 	m := NewMenuCmds(&repositories.Repos{})
// 	called := false
// 	m.loadCommands = func(int, int, string) ([]repositories.Command, int, error) {
// 		called = true
// 		return nil, 0, nil
// 	}
// 	if m.Init() != nil {
// 		t.Fatal("queried before dimensions arrived")
// 	}
// 	_, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
// 	if cmd == nil {
// 		t.Fatal("expected initial query after sizing")
// 	}
// 	m.Update(cmd())
// 	if !called {
// 		t.Fatal("initial query was not executed")
// 	}
// }

// func TestMenuCmdsSearchAndPagination(t *testing.T) {
// 	m := menuForTest(t, 23)
// 	m.Update(m.Init()())
// 	m.Update(m.fetch(2)())
// 	var searches []string
// 	var offsets []int
// 	m.loadCommands = func(limit, offset int, search string) ([]repositories.Command, int, error) {
// 		searches = append(searches, search)
// 		offsets = append(offsets, offset)
// 		return []repositories.Command{{Name: "query", Description: "Found in database"}}, 8, nil
// 	}
// 	m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
// 	m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
// 	if !m.editing || m.input.Value() != "q" || len(searches) != 0 {
// 		t.Fatal("typing triggered a query or quit")
// 	}
// 	m.input.SetValue("query")
// 	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
// 	m.Update(cmd())
// 	if m.page != 0 || m.search != "query" || m.displayedSearch != "query" || offsets[0] != 0 {
// 		t.Fatal("search did not reset to first page")
// 	}
// 	_, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
// 	m.Update(cmd())
// 	if searches[1] != "query" || offsets[1] != 5 || m.page != 1 {
// 		t.Fatal("pagination lost the search")
// 	}
// 	m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
// 	m.input.SetValue("discard")
// 	_, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
// 	if cmd != nil || m.editing || m.search != "query" || len(searches) != 2 {
// 		t.Fatal("cancel changed the applied search")
// 	}
// 	_, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
// 	m.Update(cmd())
// 	if searches[2] != "" || offsets[2] != 0 || m.search != "" {
// 		t.Fatal("escape did not clear the search")
// 	}
// 	_, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
// 	if msg, ok := cmd().(NavigateMsg); !ok || msg.To != PageMenuConfig {
// 		t.Fatal("escape without a search did not navigate back")
// 	}
// }

// func TestMenuCmdsSearchDiscardsOutdatedResultAndRetries(t *testing.T) {
// 	m := menuForTest(t, 23)
// 	m.Update(m.Init()())
// 	old := m.fetch(1)
// 	m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
// 	m.input.SetValue("new search")
// 	m.loadCommands = func(int, int, string) ([]repositories.Command, int, error) {
// 		return nil, 0, errors.New("search failed")
// 	}
// 	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
// 	m.Update(old())
// 	if !m.loading || m.page != 0 {
// 		t.Fatal("old page replaced pending search")
// 	}
// 	m.Update(cmd())
// 	if m.err == nil || len(m.list.Items()) != 5 || m.displayedSearch != "" {
// 		t.Fatal("failed search changed displayed results")
// 	}
// 	m.loadCommands = func(limit, offset int, search string) ([]repositories.Command, int, error) {
// 		if search != "new search" || offset != 0 {
// 			t.Fatalf("unexpected retry: %q at %d", search, offset)
// 		}
// 		return nil, 0, nil
// 	}
// 	_, cmd = m.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
// 	m.Update(cmd())
// 	if m.err != nil || m.displayedSearch != "new search" || !strings.Contains(m.View(), "No commands found.") {
// 		t.Fatal("retry failed to apply empty search results")
// 	}
// }
