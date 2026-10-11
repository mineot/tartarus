package pagehelpers

import (
	"tartarus/views/pages"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
)

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
	Editing         bool
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
	SearchInput     textinput.Model
	SearchKey       key.Binding
	Shortcuts       pages.Shortcuts
	Size            int
	Sized           bool
	Started         bool
	Total           int
}
