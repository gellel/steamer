package main

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type SteamPageGameGenre struct {
	Name string `json:"name"`
	URL  string `json:"URL"`
}

func NewSteamPageGameGenre(s *goquery.Selection) SteamPageGameGenre {
	return SteamPageGameGenre{
		Name: strings.TrimSpace(s.Text()),
		URL:  strings.TrimSpace(s.AttrOr("href", "NIL"))}
}
