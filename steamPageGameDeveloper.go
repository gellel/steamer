package main

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type SteamPageGameDeveloper struct {
	Name string `json:"name"`
	URL  string `json:"URL"`
}

func NewSteamPageGameDeveloper(s *goquery.Selection) SteamPageGameDeveloper {
	return SteamPageGameDeveloper{
		Name: strings.TrimSpace(s.Text()),
		URL:  strings.TrimSpace(s.AttrOr("href", "NIL"))}
}
