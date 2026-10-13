package components

import (
	"errors"
	"strings"
	"tartarus/repositories"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestCommandsListPagination(t *testing.T) {
	var offsets []int
	m := newCommandsList(func(limit, offset int, search string) ([]repositories.Command, int, error) {
		if limit != 5 || search != "" {
			t.Fatalf("unexpected query: limit=%d search=%q", limit, search)
		}
		offsets = append(offsets, offset)
		commands := make([]repositories.Command, min(limit, 7-offset))
		return commands, 7, nil
	})
	if m.Init() != nil || len(offsets) != 0 {
		t.Fatal("queried before receiving dimensions")
	}
	_, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m.Update(cmd())
	if len(m.list.Items()) != 5 || m.list.Paginator.TotalPages != 1 {
		t.Fatal("expected five items fitting in one visual page")
	}
	_, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	if cmd != nil || m.page != 0 {
		t.Fatal("moved before the first page")
	}
	_, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if !m.loading {
		t.Fatal("expected loading state")
	}
	_, blocked := m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if blocked != nil || m.page != 1 {
		t.Fatal("accepted page navigation while loading")
	}
	m.Update(cmd())
	if len(m.list.Items()) != 2 || len(offsets) != 2 || offsets[1] != 5 {
		t.Fatalf("unexpected last page: items=%d offsets=%v", len(m.list.Items()), offsets)
	}
	_, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if cmd != nil || m.page != 1 {
		t.Fatal("moved beyond the last page")
	}
	_, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
	m.Update(cmd())
	if m.page != 0 || offsets[2] != 0 {
		t.Fatal("did not return to the first page")
	}
}

func TestCommandsListRetryAndEmpty(t *testing.T) {
	calls := 0
	m := newCommandsList(func(int, int, string) ([]repositories.Command, int, error) {
		calls++
		if calls == 1 {
			return nil, 0, errors.New("query failed")
		}
		return nil, 0, nil
	})
	_, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m.Update(cmd())
	if m.err == nil || m.loading {
		t.Fatal("expected recoverable query error")
	}
	_, cmd = m.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
	m.Update(cmd())
	if m.err != nil || m.total != 0 || calls != 2 {
		t.Fatal("retry did not produce an empty successful result")
	}
	if !strings.Contains(m.View().Content, "No commands found") {
		t.Fatal("missing empty state")
	}
}

func TestCommandsListResizeIgnoresStaleResults(t *testing.T) {
	m := newCommandsList(func(limit, offset int, search string) ([]repositories.Command, int, error) {
		return make([]repositories.Command, limit), 20, nil
	})
	_, old := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	_, latest := m.Update(tea.WindowSizeMsg{Width: 80, Height: 10})
	m.Update(latest())
	m.Update(old())
	if m.pageSize != 2 || len(m.list.Items()) != 2 || m.list.Paginator.TotalPages != 1 {
		t.Fatal("resize did not retain the latest page sized to fit")
	}
	m.Update(commandsLoadedMsg{owner: &commandsListModel{}, requestID: m.requestID, total: 100})
	if m.total != 20 {
		t.Fatal("accepted another component's response")
	}
	_, cmd := m.Update(tea.WindowSizeMsg{Width: 10, Height: 3})
	if cmd != nil || m.pageSize != 0 {
		t.Fatal("queried when the terminal was too small")
	}
	_, cmd = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m.Update(cmd())
	if m.pageSize != 5 || len(m.list.Items()) != 5 {
		t.Fatal("did not reload after the terminal grew")
	}
}

func TestCommandsListRecoversAfterLastPageRemoved(t *testing.T) {
	total := 7
	m := newCommandsList(func(limit, offset int, search string) ([]repositories.Command, int, error) {
		return make([]repositories.Command, max(0, min(limit, total-offset))), total, nil
	})
	_, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m.Update(cmd())
	_, cmd = m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	total = 3
	_, reload := m.Update(cmd())
	if reload == nil || m.page != 0 {
		t.Fatal("did not recover from a removed last page")
	}
	m.Update(reload())
	if len(m.list.Items()) != 3 || m.total != 3 {
		t.Fatal("did not load the remaining commands")
	}
}
