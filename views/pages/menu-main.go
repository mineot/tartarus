package views

import "charm.land/huh/v2"

func NewMenuMain() *huh.Form {
	selection := huh.NewSelect[string]().
		Key("action").
		Title("Choose your burger").
		Options(
			huh.NewOption("Charmburger Classic", "classic"),
			huh.NewOption("Chickwich", "chickwich"),
			huh.NewOption("Fishburger", "fishburger"),
			huh.NewOption("Charmpossible™ Burger", "charmpossible"),
		)

	return huh.NewForm(huh.NewGroup(selection))
}
