package components

import (
	"fmt"
	"tartarus/themes"

	"charm.land/lipgloss/v2"
)

type listItem struct {
	name        string
	description string
}

type List struct {
	title    string
	emptyMsg string
	items    []listItem
}

func (l *List) Title(title string) {
	l.title = title
}

func (l *List) EmptyMsg(msg string) {
	l.emptyMsg = msg
}

func (l *List) AppendItem(name, description string) {
	l.items = append(
		l.items,
		listItem{name, description},
	)
}

func (l *List) Print() {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(themes.PrimaryColor))
	nameStyle := lipgloss.NewStyle().Bold(false).Foreground(lipgloss.Color(themes.PrimaryColor))
	descriptionStyle := lipgloss.NewStyle().Bold(false).Foreground(lipgloss.Color(themes.TertiaryColor))
	emptyMsgStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(themes.TertiaryColor))

	fmt.Printf("%s\n\n", titleStyle.Render(l.title))

	if len(l.items) == 0 {
		fmt.Printf("%s\n\n", emptyMsgStyle.Render(l.emptyMsg))
	} else {
		for _, item := range l.items {
			fmt.Printf(" %s\n", nameStyle.Render(item.name))
			fmt.Printf(" %s\n\n", descriptionStyle.Render(item.description))
		}
	}
}
