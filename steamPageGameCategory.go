package main

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type SteamPageGameCategory struct {
	Name string `json:"name"`
	URL  string `json:"URL"`
}

func NewSteamPageGameCategory(s *goquery.Selection) SteamPageGameCategory {
	return SteamPageGameCategory{
		Name: strings.TrimSpace(s.Text()),
		URL:  strings.TrimSpace(s.AttrOr("href", "NIL"))}
}
