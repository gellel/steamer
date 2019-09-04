package main

import (
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func newSteamSearchResult(s *goquery.Selection) *steamSearchResult {
	var cntrlID, descID, tagID []int
	appID, _ := strconv.Atoi(s.AttrOr("data-ds-appid", "-1"))
	bundleID, _ := strconv.Atoi(s.AttrOr("data-ds-bundleid", "-1"))
	packageID, _ := strconv.Atoi(s.AttrOr("data-ds-packageid", "-1"))
	URL := s.AttrOr("href", "NIL")
	cID := s.AttrOr("data-ds-crtrids", "[]")
	for _, s := range strings.Split(cID[1:len(cID)-1], ",") {
		n, err := strconv.Atoi(s)
		if err != nil {
			continue
		}
		cntrlID = append(cntrlID, n)
	}
	dID := s.AttrOr("data-ds-descids", "[]")
	for _, s := range strings.Split(dID[1:len(dID)-1], ",") {
		n, err := strconv.Atoi(s)
		if err != nil {
			continue
		}
		descID = append(descID, n)
	}
	tID := s.AttrOr("data-ds-tagids", "[]")
	for _, s := range strings.Split(tID[1:len(tID)-1], ",") {
		n, err := strconv.Atoi(s)
		if err != nil {
			continue
		}
		tagID = append(tagID, n)
	}
	return &steamSearchResult{
		AppID:     appID,
		BundleID:  bundleID,
		CrtrID:    cntrlID,
		DescID:    descID,
		PackageID: packageID,
		TagID:     tagID,
		URL:       URL}
}

type steamSearchResult struct {
	AppID     int    `json:"app_ID"`
	BundleID  int    `json:"bundle_ID"`
	CrtrID    []int  `json:"crtr_ID"`
	DescID    []int  `json:"desc_ID"`
	PackageID int    `json:"package_ID"`
	TagID     []int  `json:"tag_ID"`
	URL       string `json:"URL"`
}
