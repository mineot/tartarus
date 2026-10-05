package backup

import (
	"time"
)

// commandItem is one entry of a command, mirroring a row of command_items.
//
// Script is the text to run and Description is free text that may be left empty.
// Both are written even when empty, since no field in this format uses omitempty.
//
// CommandID repeats the id of the owning command, which the nesting in jsonFile
// already implies. It was in the format before items were nested, it is redundant
// now, and it is kept and exported rather than dropped: it is the value the row
// actually carries, and dropping a column from a backup format is a decision worth
// making deliberately rather than by omission.
type commandItem struct {
	ID          uint64    `json:"id"`
	CommandID   uint64    `json:"command_id"`
	Script      string    `json:"script"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// command is a named command with its items nested inside, mirroring a row of
// commands plus every row of its command_items.
//
// Items is the nested list rather than a separate top-level array, so a command
// and its items stay together when the file is read. Export initializes it to an
// empty slice, so a command with no items marshals as [] rather than null.
//
// ID is the id the command had in the exported database. Nothing in the format
// requires an import to reuse it.
type command struct {
	ID          uint64        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Items       []commandItem `json:"items"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// manual is a free-form document, mirroring a row of manuals.
//
// Body holds the document itself, so unlike Name and Description it is not
// optional: repositories.InsertManual rejects an empty Body with
// ErrManualBodyRequired.
type manual struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// jsonFile is the top-level shape of a backup file: every command with its items,
// then every manual.
//
// Both slices are initialized before marshalling, so an empty database produces
// an empty array on each key instead of null. That is a property of the file, not
// an accident of the encoding: the repositories return a nil slice for an empty
// table, and nil would marshal as null.
type jsonFile struct {
	Commands []command `json:"commands"`
	Manuals  []manual  `json:"manuals"`
}
