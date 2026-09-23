package main

import "strings"

// nameSimilarityThreshold is how similar two dish names must be (0-1) to be
// considered the same dish despite typos or minor wording differences.
const nameSimilarityThreshold = 0.85

// normalizeText lowercases, trims, and collapses whitespace so that
// superficial formatting differences don't affect comparisons.
func normalizeText(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(s)), " ")
}

// levenshtein computes the edit distance between two strings (rune-aware).
func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)

	prev := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(ra); i++ {
		curr := make([]int, len(rb)+1)
		curr[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			curr[j] = min3(curr[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev = curr
	}

	return prev[len(rb)]
}

func min3(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}

// similarity returns a 0-1 ratio of how alike two strings are, where 1 means
// identical and 0 means completely different.
func similarity(a, b string) float64 {
	if a == b {
		return 1
	}

	maxLen := len([]rune(a))
	if l := len([]rune(b)); l > maxLen {
		maxLen = l
	}
	if maxLen == 0 {
		return 1
	}

	dist := levenshtein(a, b)

	return 1 - float64(dist)/float64(maxLen)
}

// fuzzyNamesMatch reports whether two dish/menu names are close enough to be
// treated as the same, tolerating typos and minor wording differences.
func fuzzyNamesMatch(a, b string) bool {
	na, nb := normalizeText(a), normalizeText(b)
	if na == nb {
		return true
	}

	return similarity(na, nb) >= nameSimilarityThreshold
}

// menusMatch reports whether two comedores' menus for a day are
// "essentially the same": same number of menu options and dishes, matching
// dish types, and fuzzy-matching dish names. Allergen differences are
// ignored here since comedores often list them slightly inconsistently.
func menusMatch(a, b []MenuOption) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !fuzzyNamesMatch(a[i].Name, b[i].Name) {
			return false
		}

		dishesA, dishesB := a[i].Dishes, b[i].Dishes
		if len(dishesA) != len(dishesB) {
			return false
		}

		for j := range dishesA {
			if !strings.EqualFold(dishesA[j].Type, dishesB[j].Type) {
				return false
			}
			if !fuzzyNamesMatch(dishesA[j].Name, dishesB[j].Name) {
				return false
			}
		}
	}

	return true
}

// mergeAllergens unions two allergen lists, de-duplicating fuzzily (so
// e.g. "almendra" and "almendras" collapse into one entry, keeping the
// longer/more complete spelling) while preserving first-seen order.
func mergeAllergens(a, b []string) []string {
	merged := make([]string, 0, len(a)+len(b))

	for _, list := range [][]string{a, b} {
		for _, allergen := range list {
			allergen = strings.TrimSpace(allergen)
			if allergen == "" {
				continue
			}

			matchIdx := -1
			for i, existing := range merged {
				if fuzzyNamesMatch(existing, allergen) {
					matchIdx = i
					break
				}
			}

			switch {
			case matchIdx == -1:
				merged = append(merged, allergen)
			case len(allergen) > len(merged[matchIdx]):
				merged[matchIdx] = allergen
			}
		}
	}

	return merged
}

// mergeDay combines a group of near-duplicate days (as judged by
// menusMatch) into one Day, unioning each dish's allergens and preferring
// the longer (likely more complete) spelling of each name.
func mergeDay(group []namedDay) Day {
	merged := deepCopyDay(group[0].Day)

	for _, entry := range group[1:] {
		for i := range merged.Menus {
			if len(entry.Day.Menus[i].Name) > len(merged.Menus[i].Name) {
				merged.Menus[i].Name = entry.Day.Menus[i].Name
			}
			for j := range merged.Menus[i].Dishes {
				otherDish := entry.Day.Menus[i].Dishes[j]
				if len(otherDish.Name) > len(merged.Menus[i].Dishes[j].Name) {
					merged.Menus[i].Dishes[j].Name = otherDish.Name
				}
				merged.Menus[i].Dishes[j].Allergens = mergeAllergens(
					merged.Menus[i].Dishes[j].Allergens,
					otherDish.Allergens,
				)
			}
		}
	}

	return merged
}

// deepCopyDay copies a Day along with its nested Menus/Dishes/Allergens
// slices, so mergeDay can mutate the copy without aliasing the original.
func deepCopyDay(d Day) Day {
	menus := make([]MenuOption, len(d.Menus))
	for i, m := range d.Menus {
		dishes := make([]Dish, len(m.Dishes))
		for j, dish := range m.Dishes {
			allergens := make([]string, len(dish.Allergens))
			copy(allergens, dish.Allergens)
			dishes[j] = Dish{Type: dish.Type, Name: dish.Name, Allergens: allergens}
		}
		menus[i] = MenuOption{Name: m.Name, Dishes: dishes}
	}

	return Day{DayName: d.DayName, Date: d.Date, Menus: menus}
}
