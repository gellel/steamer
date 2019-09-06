package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
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

func chanWriteSteamChartPageDefault(wg *sync.WaitGroup, s *SteamChartPage, URL string) {
	defer wg.Done()
	err := writeSteamChartPageDefault(s)
	if err != nil {
		log.Println(fmt.Sprintf("[STEAMER] CHART %s FAILED. ERR(s): CANNOT WRITE %s", URL, err))
	}
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
