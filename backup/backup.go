package backup

import (
	"time"
)

type commandItem struct {
	ID        uint64    `json:"id"`
	CommandID uint64    `json:"command_id"`
	Script    string    `json:"script"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type command struct {
	ID        uint64        `json:"id"`
	Name      string        `json:"name"`
	Items     []commandItem `json:"items"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type manual struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type jsonFile struct {
	Commands []command `json:"commands"`
	Manuals  []manual  `json:"manuals"`
}
