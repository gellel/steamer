package main

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type SteamPageGameRequirement struct {
	DirectX   string `json:"directx"`
	Graphics  string `json:"graphics"`
	Memory    string `json:"memory"`
	Name      string `json:"name"`
	Network   string `json:"network"`
	OS        string `json:"os"`
	Processor string `json:"processor"`
	SoundCard string `json:"soundcard"`
	Storage   string `json:"storage"`
}

func NewSteamPageGameRequirement(s *goquery.Selection) SteamPageGameRequirement {
	regexp := regexp.MustCompile(`[^a-zA-Z]+`)
	steamPageGameRequirement := SteamPageGameRequirement{}
	s.Find("ul.bb_ul").First().Each(func(i int, s *goquery.Selection) {
		valueMap := map[string]string{}
		s.Find("li").Each(func(j int, s *goquery.Selection) {
			key := s.Find("strong").First().Text()
			key = regexp.ReplaceAllString(key, "")
			key = strings.ToLower(key)
			valueMap[key] = strings.TrimSpace(s.Text())
		})
		b, err := json.Marshal(valueMap)
		if err != nil {
			panic(err)
		}
		if err := json.Unmarshal(b, &steamPageGameRequirement); err != nil {
			panic(err)
		}
	})
	return steamPageGameRequirement
}
