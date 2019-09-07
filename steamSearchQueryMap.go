package main

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type SteamSearchQueryMap map[string]*SteamSearchKeyValue

func NewSteamSearchQueryMap(s *goquery.Selection) *SteamSearchQueryMap {
	replacer := strings.NewReplacer("/", "-", "\\", "-", " ", "-", "&", "and", "+", "-")
	regexp := regexp.MustCompile(`-{2,}`)
	steamSearchQueryMap := &SteamSearchQueryMap{}
	s.Each(func(i int, s *goquery.Selection) {
		dataParam, ok := s.Attr("data-param")
		if ok != true {
			return
		}
		dataValue, ok := s.Attr("data-value")
		if ok != true {
			return
		}
		dataLoc, ok := s.Attr("data-loc")
		if ok != true {
			return
		}
		dataLoc = replacer.Replace(dataLoc)
		dataLoc = strings.ToUpper(dataLoc)
		dataLoc = regexp.ReplaceAllString(dataLoc, "")
		steamSearchQueryMap.Add(dataLoc, dataParam, dataValue)
	})
	return steamSearchQueryMap
}

func (steamSearchQueryMap *SteamSearchQueryMap) Add(dataLoc, dataParam, dataValue string) bool {
	ok := steamSearchQueryMap.Has(dataLoc)
	if ok != true {
		(*steamSearchQueryMap)[dataLoc] = NewSteamSearchKeyValue(dataParam, dataValue)
	}
	return (ok == false)
}

func (steamSearchQueryMap *SteamSearchQueryMap) Get(dataLoc string) (*SteamSearchKeyValue, bool) {
	steamSearchKeyValue, ok := (*steamSearchQueryMap)[dataLoc]
	return steamSearchKeyValue, ok
}

func (steamSearchQueryMap *SteamSearchQueryMap) Has(dataLoc string) bool {
	_, ok := steamSearchQueryMap.Get(dataLoc)
	return ok
}
