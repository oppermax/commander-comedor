package main

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

const menuURL = "https://scu.ugr.es/"

// getMenu fetches and parses the current menu from the UGR comedores site.
func getMenu() Menu {
	client := http.Client{}

	req, err := http.NewRequest("GET", menuURL, nil)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to create GET menu request")
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Panic().Err(err).Msg("GET menu request failed")
	}
	defer resp.Body.Close()

	menu, err := parseMenu(resp.Body)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to parse menu response body")
	}

	return menu
}
