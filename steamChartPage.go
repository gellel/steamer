package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type SteamChartPage struct {
	Delta            time.Time              `json:"delta"`
	Growth           []SteamChartGameGrowth `json:"growth"`
	Name             string                 `json:"name"`
	PlayerPeek24Hour int                    `json:"player_peak_24_hour"`
	PlayerPeekAll    int                    `json:"player_peek_all"`
	PlayerPeekDelta  int                    `json:"player_peek_delta"`
	Timestamp        time.Time              `json:"timestamp"`
	URL              string                 `json:"URL"`
}

func NewSteamChartPage(s *goquery.Selection) *SteamChartPage {
	return &SteamChartPage{
		Delta:            scrapeSteamChartGameDelta(s),
		Growth:           scrapeSteamChartGameGrowth(s),
		Name:             scrapeSteamChartGameName(s),
		PlayerPeekAll:    scrapeSteamChartGamePlayerPeekAll(s),
		PlayerPeek24Hour: scrapeSteamChartGamePlayerPeek24Hour(s),
		PlayerPeekDelta:  scrapeSteamChartGamePlayerPeekDelta(s)}
}

func onGetSteamChartPage(c *http.Client, URL string, revisit bool, snap func(s *Snapshot), success func(s *SteamChartPage), err func(e error)) {
	if revisit == false {
		if u, err := url.Parse(URL); err == nil {
			if ok, _ := hasVisitedURLDefault(u); ok {
				return
			}
		}
	}
	snapshot := NewSnapshot(c, http.MethodGet, URL, nil)
	snap(snapshot)
	if ok := (snapshot.StatusCode == http.StatusOK); ok != true {
		err(errors.New(snapshot.Status))
		return
	}
	if ok := (snapshot.document != nil); ok != true {
		err(snapshot.ErrDoc)
		return
	}
	CSSSelector := "html"
	goQuerySelection := snapshot.document.Find(CSSSelector)
	goQuerySelectionLength := goQuerySelection.Length()
	if ok := (goQuerySelectionLength > 0); ok != true {
		err(errors.New("goquery.Selection empty"))
		return
	}
	steamChartPage := NewSteamChartPage(goQuerySelection)
	if ok := len(steamChartPage.Name) > 0; ok != true {
		err(errors.New("SteamChart.AppID negative"))
		return
	}
	success(steamChartPage)
}

func scrapeSteamChartGameDelta(s *goquery.Selection) time.Time {
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(s.Find("div.app-stat abbr.timeago").Text()))
	if err != nil {
		return time.Time{}
	}
	return t
}

func scrapeSteamChartGameGrowth(s *goquery.Selection) []SteamChartGameGrowth {
	var steamChartGameGrowth []SteamChartGameGrowth
	s.Find("table.common-table tbody").First().Find("tr").Each(func(i int, s *goquery.Selection) {
		steamChartGameGrowth = append(steamChartGameGrowth, NewSteamChartGameGrowth(s))
	})
	return steamChartGameGrowth
}

func scrapeSteamChartGameName(s *goquery.Selection) string {
	return regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(strings.TrimSpace(s.Find("#app-title").Text()), "")
}

func scrapeSteamChartGamePlayerPeek24Hour(s *goquery.Selection) int {
	n, err := strconv.Atoi(strings.TrimSpace(s.Find("#app-heading div:nth-child(3) span.num").Text()))
	if err != nil {
		return -1
	}
	return n
}

func scrapeSteamChartGamePlayerPeekAll(s *goquery.Selection) int {
	n, err := strconv.Atoi(strings.TrimSpace(s.Find("#app-heading div:nth-child(4) span.num").Text()))
	if err != nil {
		return -1
	}
	return n
}

func scrapeSteamChartGamePlayerPeekDelta(s *goquery.Selection) int {
	n, err := strconv.Atoi(strings.TrimSpace(s.Find("#app-heading div:nth-child(2) span.num").Text()))
	if err != nil {
		return -1
	}
	return n
}

func writeSteamChartPage(fullpath string, s *SteamChartPage) error {
	err := os.MkdirAll(fullpath, os.ModePerm)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	filename := fmt.Sprintf("chart-result-%s.json", s.Name)
	fullname := filepath.Join(fullpath, filename)
	err = ioutil.WriteFile(fullname, b, os.ModePerm)
	return err
}

func writeSteamChartPageDefault(s *SteamChartPage) error {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fullpath := filepath.Join(user.HomeDir, "Desktop", "steambot", "games", s.Name)
	err = writeSteamChartPage(fullpath, s)
	return err
}
