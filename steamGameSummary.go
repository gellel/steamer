package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

type SteamGameSummary struct {
	Available              bool      `json:"available"`
	AverageDecline         int       `json:"average_decline"`
	AverageGain            int       `json:"average_gain"`
	AverageMaxPlayerCount  int       `json:"average_max_player_count"`
	AverageMinPlayerCount  int       `json:"average_min_player_count"`
	AveragePlayerCount     int       `json:"average_player_count"`
	Categories             []string  `json:"categories"`
	ComingSoon             bool      `json:"coming_soon"`
	Developers             []string  `json:"developers"`
	EarlyAccess            bool      `json:"early_access"`
	Genres                 []string  `json:"genres"`
	Name                   string    `json:"name"`
	MonthsSinceRelease     int       `json:"months_since_release"`
	PeakPlayers            int       `json:"peak_players"`
	PeakPlayersDate        string    `json:"peak_players_date"`
	PlayerPeak24Hour       int       `json:"player_peak_24_hour"`
	PlayerPeakAll          int       `json:"player_peak_all"`
	Publishers             []string  `json:"publishers"`
	ReleaseDate            time.Time `json:"release_date"`
	ReviewsAllCount        int       `json:"reviews_all_count"`
	ReviewsAllSentiment    string    `json:"reviews_all_sentiment"`
	ReviewsRecentCount     int       `json:"reviews_recent_count"`
	ReviewsRecentSentiment string    `json:"reviews_recent_sentiment"`
	SocialMedia            []string  `json:"social_media"`
	Tags                   []string  `json:"tags"`
	Timestamp              time.Time `json:"timestamp"`
	Title                  string    `json:"title"`
	URL                    string    `json:"URL"`
	Website                string    `json:"website"`
}

func NewSteamGameSummary(steamGamePage *SteamGamePage, steamChartPage *SteamChartPage) *SteamGameSummary {
	var averageDecline, averageGain, averageMaxPlayerCount, averageMinPlayerCount, monthsSinceRelease, peakPlayers int
	var peakPlayersDate string

	if len(steamChartPage.Growth) > 0 {
		var n int
		for _, steamChartGameGrowth := range steamChartPage.Growth {
			if steamChartGameGrowth.Gain > 0 {
				averageGain = averageGain + int(steamChartGameGrowth.Gain)
				averageMaxPlayerCount = averageMaxPlayerCount + steamChartGameGrowth.PlayersPeak
			} else {
				averageDecline = averageDecline + int(steamChartGameGrowth.Gain)
				averageMinPlayerCount = averageMinPlayerCount + steamChartGameGrowth.PlayersPeak
			}
			if steamChartGameGrowth.PlayersPeak > peakPlayers {
				peakPlayers = steamChartGameGrowth.PlayersPeak
				peakPlayersDate = steamChartGameGrowth.Month
			}
			n = n + 1
		}
		monthsSinceRelease = int(n / 12)
		averageDecline = averageDecline / n
		averageGain = averageGain / n
		averageMaxPlayerCount = averageMaxPlayerCount / n
		averageMinPlayerCount = averageMinPlayerCount / n
	}
	return &SteamGameSummary{
		Available:              steamGamePage.Available,
		AverageDecline:         averageDecline,
		AverageGain:            averageGain,
		AverageMaxPlayerCount:  averageMaxPlayerCount,
		AverageMinPlayerCount:  averageMinPlayerCount,
		ComingSoon:             steamGamePage.ComingSoon,
		EarlyAccess:            steamGamePage.EarlyAccess,
		Name:                   steamGamePage.Name,
		MonthsSinceRelease:     monthsSinceRelease,
		PeakPlayers:            peakPlayers,
		PeakPlayersDate:        peakPlayersDate,
		PlayerPeak24Hour:       steamChartPage.PlayerPeak24Hour,
		PlayerPeakAll:          steamChartPage.PlayerPeakAll,
		ReviewsAllCount:        steamGamePage.ReviewsAll.Count,
		ReviewsAllSentiment:    steamGamePage.ReviewsAll.Sentiment,
		ReviewsRecentCount:     steamGamePage.ReviewsRecent.Count,
		ReviewsRecentSentiment: steamGamePage.ReviewsRecent.Sentiment,
		Timestamp:              time.Now(),
		Title:                  steamGamePage.Title,
		URL:                    steamGamePage.URL,
		Website:                steamGamePage.Website}
}

func writeSteamGameSummary(fullpath string, s *SteamGameSummary) error {
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
	filename := fmt.Sprintf("summary-%s.json", s.Name)
	fullname := filepath.Join(fullpath, filename)
	err = ioutil.WriteFile(fullname, b, os.ModePerm)
	return err
}

func writeSteamGameSummaryDefault(s *SteamGameSummary) error {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fullpath := filepath.Join(user.HomeDir, "Desktop", "steambot", "games", s.Name)
	err = writeSteamGameSummary(fullpath, s)
	return err
}
