package main

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type SteamPageGameTag struct {
	Name string `json:"name"`
	URL  string `json:"URL"`
}

func NewSteamPageGameTag(s *goquery.Selection) SteamPageGameTag {
	return SteamPageGameTag{
		Name: strings.TrimSpace(s.Text()),
		URL:  strings.TrimSpace(s.AttrOr("href", "NIL"))}
}
