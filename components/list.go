package components

import (
	"fmt"
	"tartarus/themes"
)

type listItem struct {
	id          uint64
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

func (l *List) AppendItem(id uint64, name, description string) {
	l.items = append(
		l.items,
		listItem{id, name, description},
	)
}

func (l *List) Print() {
	fmt.Printf("%s\n\n", themes.GetPrimaryText(l.title, true))

	if len(l.items) == 0 {
		fmt.Printf("%s\n\n", themes.GetTertiaryText(l.emptyMsg, true))
	} else {
		for _, item := range l.items {
			fmt.Println(item.getFormated())
		}
	}
}

func (l *listItem) getId() string {
	result := fmt.Sprintf("%d", l.id)
	result = fmt.Sprintf("ID: %s", result)
	return themes.GetAssetText(result, true)
}

func (l *listItem) getName() string {
	return themes.GetPrimaryText(fmt.Sprintf("Name: %s", l.name), false)
}

func (l *listItem) getDescription() string {
	return themes.GetTertiaryText(fmt.Sprintf("Description: %s", l.description), false)
}

func (l *listItem) getFormated() string {
	return fmt.Sprintf("%s %s %s", l.getId(), l.getName(), l.getDescription())
}
