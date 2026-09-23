package main

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// spanishWeekdays maps Go's Weekday to the uppercase Spanish day names
// used on the UGR comedores site (e.g. "LUNES").
var spanishWeekdays = map[time.Weekday]string{
	time.Monday:    "LUNES",
	time.Tuesday:   "MARTES",
	time.Wednesday: "MIÉRCOLES",
	time.Thursday:  "JUEVES",
	time.Friday:    "VIERNES",
	time.Saturday:  "SÁBADO",
	time.Sunday:    "DOMINGO",
}

// CLI is the root command. Global flags apply to every subcommand.
type CLI struct {
	Comedor string `help:"Only show a specific comedor (matches by substring, case-insensitive)." short:"c"`

	Today    TodayCmd    `cmd:"" default:"withargs" help:"Show today's menu (default)."`
	Tomorrow TomorrowCmd `cmd:"" help:"Show tomorrow's menu."`
	Week     WeekCmd     `cmd:"" help:"Show the full week's menu."`
}

// TodayCmd shows the menu for the current day. It is the default command,
// so running the tool with no arguments shows today's menu.
type TodayCmd struct{}

func (c *TodayCmd) Run(cli *CLI) error {
	return showDay(cli, time.Now().Weekday())
}

// TomorrowCmd shows the menu for the next day.
type TomorrowCmd struct{}

func (c *TomorrowCmd) Run(cli *CLI) error {
	return showDay(cli, time.Now().Add(24*time.Hour).Weekday())
}

// WeekCmd shows the full week's menu for every published day.
type WeekCmd struct{}

func (c *WeekCmd) Run(cli *CLI) error {
	menu := getMenu()

	keys, byKey := groupDaysByDate(menu.Comedores, cli.Comedor)
	for _, key := range keys {
		printMergedDay(byKey[key])
	}

	return nil
}

// showDay fetches the menu and prints the entry matching weekday, filtered
// by the --comedor flag if set. It is shared by TodayCmd and TomorrowCmd.
func showDay(cli *CLI, weekday time.Weekday) error {
	dayName := spanishWeekdays[weekday]
	menu := getMenu()

	var matches []namedDay
	for _, comedor := range menu.Comedores {
		if !matchesComedor(comedor.Name, cli.Comedor) {
			continue
		}
		for _, day := range comedor.Days {
			if day.DayName != dayName {
				continue
			}
			matches = append(matches, namedDay{Name: comedor.Name, Day: day})
		}
	}

	if len(matches) == 0 {
		if isWeekend(weekday) {
			fmt.Printf("No menu found for %s. It's the weekend!\n", dayName)
		} else {
			fmt.Printf("No menu found for %s. It may not be published yet.\n", dayName)
		}
		return nil
	}

	printMergedDay(matches)

	return nil
}

func matchesComedor(name, filter string) bool {
	if filter == "" {
		return true
	}
	return strings.Contains(strings.ToLower(name), strings.ToLower(filter))
}

// namedDay pairs a comedor name with one of its days, so days from
// different comedores can be grouped and deduplicated when identical.
type namedDay struct {
	Name string
	Day  Day
}

// groupDaysByDate collects every comedor's days (filtered by the --comedor
// flag), keyed by "DayName|Date" and in first-seen order, so the week can
// be printed date by date across comedores.
func groupDaysByDate(comedores []Comedor, filter string) ([]string, map[string][]namedDay) {
	var keys []string
	byKey := make(map[string][]namedDay)

	for _, comedor := range comedores {
		if !matchesComedor(comedor.Name, filter) {
			continue
		}
		for _, day := range comedor.Days {
			key := day.DayName + "|" + day.Date
			if _, ok := byKey[key]; !ok {
				keys = append(keys, key)
			}
			byKey[key] = append(byKey[key], namedDay{Name: comedor.Name, Day: day})
		}
	}

	return keys, byKey
}

// printMergedDay prints one day's menu, merging comedores whose menus are
// identical into a single entry (e.g. "Fuentenueva / Cartuja") instead of
// repeating the same dishes for each comedor.
func printMergedDay(entries []namedDay) {
	var groups [][]namedDay

	for _, entry := range entries {
		placed := false
		for i, group := range groups {
			if reflect.DeepEqual(group[0].Day.Menus, entry.Day.Menus) {
				groups[i] = append(group, entry)
				placed = true
				break
			}
		}
		if !placed {
			groups = append(groups, []namedDay{entry})
		}
	}

	for _, group := range groups {
		names := make([]string, len(group))
		for i, entry := range group {
			names[i] = entry.Name
		}
		printComedorDay(strings.Join(names, " / "), group[0].Day)
	}
}

func printComedorDay(comedorName string, day Day) {
	fmt.Printf("\n=== %s \u2014 %s, %s ===\n", comedorName, day.DayName, day.Date)

	for _, m := range day.Menus {
		fmt.Println(m.Name)
		for _, dish := range m.Dishes {
			line := fmt.Sprintf("  %-16s %s", dish.Type, dish.Name)
			if len(dish.Allergens) > 0 {
				line += " (" + strings.Join(dish.Allergens, ", ") + ")"
			}
			fmt.Println(line)
		}
	}
}

func isWeekend(day time.Weekday) bool {
	return day == time.Saturday || day == time.Sunday
}
