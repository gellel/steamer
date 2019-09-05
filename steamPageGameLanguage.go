package main

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type SteamPageGameLanguage struct {
	Audio     bool   `json:"audio"`     // {Audio: true}
	Interface bool   `json:"interface"` // {Interface: true}
	Name      string `json:"name"`      // {Name: "ENGLISH"}
	Subtitles bool   `json:"subtitles"` // {Subtitles: true}
}

func NewSteamPageGameLanguage(s *goquery.Selection) SteamPageGameLanguage {
	var (
		lang      = strings.TrimSpace(s.Find("td:nth-child(1)").Text())
		inter     = strings.TrimSpace(s.Find("td:nth-child(2)").Text())
		audio     = strings.TrimSpace(s.Find("td:nth-child(3)").Text())
		subtitles = strings.TrimSpace(s.Find("td:nth-child(4)").Text())
	)
	return SteamPageGameLanguage{
		Audio:     (len(audio) != 0),
		Interface: (len(inter) != 0),
		Name:      lang,
		Subtitles: (len(subtitles) != 0)}
}
