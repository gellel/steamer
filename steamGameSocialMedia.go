package main

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type SteamGameSocialMedia struct {
	Name string `json:"name"`
	URL  string `json:"URL"`
}

func NewSteamGameSocialMedia(s *goquery.Selection) SteamGameSocialMedia {
	rawURL := strings.TrimPrefix(s.AttrOr("href", ""), "https://steamcommunity.com/linkfilter/?url=")
	URL, _ := url.Parse(rawURL)
	return SteamGameSocialMedia{
		Name: URL.Host,
		URL:  rawURL}
}
