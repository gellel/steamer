package main

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type SteamSearchQueryCategoryMap map[string]*SteamSearchQueryMap

func NewSteamSearchQueryCategoryMap(s *goquery.Selection) *SteamSearchQueryCategoryMap {
	steamSearchQueryCategoryMap := &SteamSearchQueryCategoryMap{}
	s.Each(func(i int, s *goquery.Selection) {
		dataParam, ok := s.Attr("data-param")
		if ok != true {
			return
		}
		if ok := strings.ToUpper(dataParam) == "HIDE"; ok {
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
		steamSearchQueryCategoryMap.Add(dataParam, dataLoc, dataValue)
	})
	return steamSearchQueryCategoryMap
}

func (steamSearchQueryCategoryMap *SteamSearchQueryCategoryMap) Add(dataParam, dataLoc, dataValue string) bool {
	steamSearchQueryCategoryMap.Set(dataParam)
	steamSearchQueryMap := steamSearchQueryCategoryMap.Fetch(dataParam)
	steamSearchQueryMap.Add(dataLoc, dataValue)
	return steamSearchQueryCategoryMap.Has(dataParam) && steamSearchQueryMap.Has(dataLoc)
}

func (steamSearchQueryCategoryMap *SteamSearchQueryCategoryMap) Fetch(dataParam string) *SteamSearchQueryMap {
	steamSearchQueryMap, _ := steamSearchQueryCategoryMap.Get(dataParam)
	return steamSearchQueryMap
}

func (steamSearchQueryCategoryMap *SteamSearchQueryCategoryMap) Get(dataParam string) (*SteamSearchQueryMap, bool) {
	steamSearchQueryMap, ok := (*steamSearchQueryCategoryMap)[dataParam]
	return steamSearchQueryMap, ok
}

func (steamSearchQueryCategoryMap *SteamSearchQueryCategoryMap) Has(dataParam string) bool {
	_, ok := steamSearchQueryCategoryMap.Get(dataParam)
	return ok
}

func (steamSearchQueryCategoryMap *SteamSearchQueryCategoryMap) Set(dataParam string) bool {
	ok := steamSearchQueryCategoryMap.Has(dataParam)
	if ok != true {
		(*steamSearchQueryCategoryMap)[dataParam] = &SteamSearchQueryMap{}
	}
	return (ok == false)
}
