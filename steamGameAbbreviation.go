package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type SteamGameAbbreviation struct {
	AppID     int    `json:"app_ID"`
	BundleID  int    `json:"bundle_ID"`
	CrtrID    []int  `json:"crtr_ID"`
	DescID    []int  `json:"desc_ID"`
	Name      string `json:"name"`
	PackageID int    `json:"package_ID"`
	TagID     []int  `json:"tag_ID"`
	URL       string `json:"URL"`
}

func NewSteamGameAbbreviation(s *goquery.Selection) *SteamGameAbbreviation {
	return &SteamGameAbbreviation{
		AppID:     scrapeSteamAbbreviationAppID(s),
		BundleID:  scrapeSteamAbbreviationBundleID(s),
		CrtrID:    scrapeSteamAbbreviationCrtrID(s),
		DescID:    scrapeSteamAbbreviationDescID(s),
		Name:      scrapeSteamAbbreviationName(s),
		PackageID: scrapeSteamAbbreviationPackageID(s),
		TagID:     scrapeSteamAbbreviationTagID(s),
		URL:       s.AttrOr("href", "NIL")}
}

func chanSteamGameAbbreviation(c chan *SteamGameAbbreviation, d *goquery.Document) {
	d.Find("a.search_result_row[href]").Each(func(i int, s *goquery.Selection) {
		defer wg.Done()
		c <- NewSteamGameAbbreviation(s)
	})
}

func scrapeSteamAbbreviationAppID(s *goquery.Selection) int {
	ID, _ := strconv.Atoi(s.AttrOr("data-ds-appid", "-1"))
	return ID
}

func scrapeSteamAbbreviationBundleID(s *goquery.Selection) int {
	ID, _ := strconv.Atoi(s.AttrOr("data-ds-bundleid", "-1"))
	return ID
}

func scrapeSteamAbbreviationCrtrID(s *goquery.Selection) []int {
	var crtrID []int
	cID := s.AttrOr("data-ds-crtrids", "[]")
	for _, s := range strings.Split(cID[1:len(cID)-1], ",") {
		n, err := strconv.Atoi(s)
		if err != nil {
			continue
		}
		crtrID = append(crtrID, n)
	}
	return crtrID
}

func scrapeSteamAbbreviationDescID(s *goquery.Selection) []int {
	var descID []int
	dID := s.AttrOr("data-ds-descids", "[]")
	for _, s := range strings.Split(dID[1:len(dID)-1], ",") {
		n, err := strconv.Atoi(s)
		if err != nil {
			continue
		}
		descID = append(descID, n)
	}
	return descID
}

func scrapeSteamAbbreviationName(s *goquery.Selection) string {
	return regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(strings.TrimSpace(s.Find(".title").Text()), "")
}

func scrapeSteamAbbreviationPackageID(s *goquery.Selection) int {
	ID, _ := strconv.Atoi(s.AttrOr("data-ds-packageid", "-1"))
	return ID
}

func scrapeSteamAbbreviationTagID(s *goquery.Selection) []int {
	var tagID []int
	tID := s.AttrOr("data-ds-tagids", "[]")
	for _, s := range strings.Split(tID[1:len(tID)-1], ",") {
		n, err := strconv.Atoi(s)
		if err != nil {
			continue
		}
		tagID = append(tagID, n)
	}
	return tagID
}

func writeSteamGameAbbreviation(fullpath string, s *SteamGameAbbreviation) error {
	err := os.MkdirAll(fullpath, os.ModePerm)
	if err != nil {
		return err
	}
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	filename := fmt.Sprintf("search-result-%d.json", s.AppID)
	fullname := filepath.Join(fullpath, filename)
	err = ioutil.WriteFile(fullname, b, os.ModePerm)
	return err
}

func writeSteamGameAbbreviationDefault(s *SteamGameAbbreviation) error {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fullpath := filepath.Join(user.HomeDir, "Desktop", "steambot", s.Name)
	err = writeSteamGameAbbreviation(fullpath, s)
	return err
}
