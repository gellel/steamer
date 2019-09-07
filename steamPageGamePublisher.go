package main

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type SteamPageGamePublisher struct {
	Name string `json:"name"`
	URL  string `json:"URL"`
}

func NewSteamPageGamePublisher(s *goquery.Selection) SteamPageGamePublisher {
	return SteamPageGamePublisher{
		Name: strings.TrimSpace(s.Text()),
		URL:  strings.TrimSpace(s.AttrOr("href", "NIL"))}
}
