package pagemenus

import (
	"tartarus/views/pages"

	"charm.land/huh/v2"
)

type PageMenuMain struct {
	form      *huh.Form
	selected  string
	shortcuts pages.Shortcuts
	width     int
}

type PageMenuConfigurations struct {
	form      *huh.Form
	selected  string
	shortcuts pages.Shortcuts
	width     int
}
