package main

import (
	"io"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// splitAllergens splits raw allergen text on runs of 2+ spaces,
// which is how the site separates multiple allergens within one cell.
var allergenSplitter = regexp.MustCompile(`\s{2,}`)

func splitAllergens(raw string) []string {
	parts := allergenSplitter.Split(strings.TrimSpace(raw), -1)

	allergens := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			allergens = append(allergens, p)
		}
	}

	return allergens
}

// parseMenu parses the Comedores Universitarios de Granada HTML page
// into a Menu struct.
func parseMenu(r io.Reader) (Menu, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return Menu{}, err
	}

	menu := Menu{}

	doc.Find("h1:contains('Menú semanal')").Each(func(_ int, h1 *goquery.Selection) {
		comedorName := strings.TrimSpace(
			strings.TrimPrefix(h1.Text(), "Menú semanal |"),
		)

		table := h1.NextUntil("h1").Find("table.inline").First()
		if table.Length() == 0 {
			return
		}

		comedor := Comedor{Name: comedorName}

		var currentDay *Day
		var currentMenu *MenuOption

		table.Find("tr").Each(func(_ int, row *goquery.Selection) {
			cells := row.Find("th, td")
			if cells.Length() == 0 {
				return
			}

			first := cells.Eq(0)

			switch goquery.NodeName(first) {
			case "th":
				// New day header, e.g. "LUNES,  21  DE  SEPTIEMBRE  DE  2026"
				if currentDay != nil {
					if currentMenu != nil {
						currentDay.Menus = append(currentDay.Menus, *currentMenu)
						currentMenu = nil
					}
					comedor.Days = append(comedor.Days, *currentDay)
				}

				text := strings.TrimSpace(first.Text())
				dayName, date, _ := strings.Cut(text, ",")
				currentDay = &Day{
					DayName: strings.TrimSpace(dayName),
					Date:    strings.TrimSpace(date),
				}

			case "td":
				firstText := strings.TrimSpace(first.Text())
				if firstText == "" {
					// Footer row, e.g. "Consultar ingredientes semanales"
					return
				}

				if attr, ok := first.Attr("colspan"); ok && attr == "2" {
					// New menu option header, e.g. "Menú 1"
					if currentMenu != nil && currentDay != nil {
						currentDay.Menus = append(currentDay.Menus, *currentMenu)
					}
					currentMenu = &MenuOption{Name: firstText}
					return
				}

				if currentMenu == nil {
					return
				}

				dish := Dish{
					Type: firstText,
					Name: strings.TrimSpace(cells.Eq(1).Find("strong").Text()),
				}
				if cells.Length() > 2 {
					dish.Allergens = splitAllergens(cells.Eq(2).Text())
				}

				currentMenu.Dishes = append(currentMenu.Dishes, dish)
			}
		})

		if currentMenu != nil && currentDay != nil {
			currentDay.Menus = append(currentDay.Menus, *currentMenu)
		}
		if currentDay != nil {
			comedor.Days = append(comedor.Days, *currentDay)
		}

		menu.Comedores = append(menu.Comedores, comedor)
	})

	return menu, nil
}
