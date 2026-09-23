package main

// Menu represents the entire mensa menu
type Menu struct {
	Comedores []Comedor `json:"comedores"`
}

// Comedor represents a dining hall (e.g., Fuentenueva, Cartuja, PTS)
type Comedor struct {
	Name  string `json:"name"`
	Days  []Day  `json:"days"`
}

// Day represents a single day's menu
type Day struct {
	DayName string       `json:"day_name"`      // e.g., "LUNES"
	Date    string       `json:"date"`          // e.g., "21 DE SEPTIEMBRE DE 2026"
	Menus   []MenuOption `json:"menus"`
}

// MenuOption represents a single menu option for a day (e.g., "Menú 1", "Menú 2")
type MenuOption struct {
	Name   string  `json:"name"`
	Dishes []Dish  `json:"dishes"`
}

// Dish represents a single dish in a menu
type Dish struct {
	Type       string   `json:"type"`        // e.g., "Primero", "Segundo", "Acompañamiento", "Postre"
	Name       string   `json:"name"`
	Allergens  []string `json:"allergens"`
}
