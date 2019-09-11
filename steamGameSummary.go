package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
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
	TroughPlayers          int       `json:"trough_players"`
	TroughPlayersDate      string    `json:"trough_players_date"`
	URL                    string    `json:"URL"`
	Website                string    `json:"website"`
	YearsSinceRelease      int       `json:"years_since_release"`
}

func NewSteamGameSummary(steamGamePage *SteamGamePage, steamChartPage *SteamChartPage) *SteamGameSummary {
	steamGameSummaryStatistics := NewSteamGameSummaryStatistics(steamChartPage)
	return &SteamGameSummary{
		Available:              steamGamePage.Available,
		AverageDecline:         steamGameSummaryStatistics.AverageDecline,
		AverageGain:            steamGameSummaryStatistics.AverageGain,
		AverageMaxPlayerCount:  steamGameSummaryStatistics.AverageMaxPlayerCount,
		AverageMinPlayerCount:  steamGameSummaryStatistics.AverageMinPlayerCount,
		AveragePlayerCount:     steamGameSummaryStatistics.AveragePlayerCount,
		Categories:             parseSteamGameSummaryCategories(&steamGamePage.Categories),
		ComingSoon:             steamGamePage.ComingSoon,
		Developers:             parseSteamGameSummaryDevelopers(&steamGamePage.Developers),
		EarlyAccess:            steamGamePage.EarlyAccess,
		Genres:                 parseSteamGameSummaryGenres(&steamGamePage.Genres),
		Name:                   steamGamePage.Name,
		MonthsSinceRelease:     steamGameSummaryStatistics.MonthsSinceRelease,
		PeakPlayers:            steamGameSummaryStatistics.PeakPlayers,
		PeakPlayersDate:        steamGameSummaryStatistics.PeakPlayersDate,
		PlayerPeak24Hour:       steamChartPage.PlayerPeak24Hour,
		PlayerPeakAll:          steamChartPage.PlayerPeakAll,
		Publishers:             parseSteamGameSummaryPublishers(&steamGamePage.Publishers),
		ReleaseDate:            steamGamePage.ReleaseDate,
		ReviewsAllCount:        steamGamePage.ReviewsAll.Count,
		ReviewsAllSentiment:    steamGamePage.ReviewsAll.Sentiment,
		ReviewsRecentCount:     steamGamePage.ReviewsRecent.Count,
		ReviewsRecentSentiment: steamGamePage.ReviewsRecent.Sentiment,
		SocialMedia:            parseSteamGameSummarySocialMedia(&steamGamePage.SocialMedia),
		Tags:                   parseSteamGameSummaryTags(&steamGamePage.Tags),
		Timestamp:              time.Now(),
		Title:                  steamGamePage.Title,
		TroughPlayers:          steamGameSummaryStatistics.TroughPlayers,
		TroughPlayersDate:      steamGameSummaryStatistics.TroughPlayersDate,
		URL:                    steamGamePage.URL,
		Website:                steamGamePage.Website,
		YearsSinceRelease:      steamGameSummaryStatistics.YearsSinceRelease}
}

func parseSteamGameSummaryCategories(s *[]SteamPageGameCategory) []string {
	v := *s
	categories := make([]string, len(v))
	for i, p := range v {
		categories[i] = p.Name
	}
	return categories
}

func parseSteamGameSummaryDevelopers(s *[]SteamPageGameDeveloper) []string {
	v := *s
	developers := make([]string, len(v))
	for i, p := range v {
		developers[i] = p.Name
	}
	return developers
}

func parseSteamGameSummaryGenres(s *[]SteamPageGameGenre) []string {
	v := *s
	genres := make([]string, len(v))
	for i, p := range v {
		genres[i] = p.Name
	}
	return genres
}

func parseSteamGameSummaryPublishers(s *[]SteamPageGamePublisher) []string {
	v := *s
	publishers := make([]string, len(v))
	for i, p := range v {
		publishers[i] = p.Name
	}
	return publishers
}

func parseSteamGameSummarySocialMedia(s *[]SteamGameSocialMedia) []string {
	v := *s
	social := make([]string, len(v))
	for i, p := range v {
		social[i] = p.URL
	}
	return social
}

func parseSteamGameSummaryTags(s *[]SteamPageGameTag) []string {
	v := *s
	tags := make([]string, len(v))
	for i, p := range v {
		tags[i] = p.Name
	}
	return tags
}

func writeSteamGameSummaryCSV(fullpath string, s *SteamGameSummary) error {
	err := os.MkdirAll(fullpath, os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.Create(filepath.Join(fullpath, "summary.csv"))
	defer file.Close()
	if err != nil {
		return err
	}
	writer := csv.NewWriter(file)
	defer writer.Flush()
	rows := [][]string{
		[]string{"Available", fmt.Sprintf("%t", s.Available)},
		[]string{"Average Decline", fmt.Sprintf("%d", s.AverageDecline)},
		[]string{"Average Gain", fmt.Sprintf("%d", s.AverageGain)},
		[]string{"Average Max Player Count", fmt.Sprintf("%d", s.AverageMaxPlayerCount)},
		[]string{"Average Min Player Count", fmt.Sprintf("%d", s.AverageMinPlayerCount)},
		[]string{"Average Player Count", fmt.Sprintf("%d", s.AveragePlayerCount)},
		append([]string{"Categories"}, s.Categories...),
		[]string{"ComingSoon", fmt.Sprintf("%t", s.ComingSoon)},
		append([]string{"Developers"}, s.Developers...),
		[]string{"Early Access", fmt.Sprintf("%t", s.EarlyAccess)},
		append([]string{"Genres"}, s.Genres...),
		[]string{"Name", s.Name},
		[]string{"Months Since Release", fmt.Sprintf("%d", s.MonthsSinceRelease)},
		[]string{"Peak Players", fmt.Sprintf("%d", s.PeakPlayers)},
		[]string{"Peak Players Date", s.PeakPlayersDate},
		[]string{"Player Peak 24 Hour", fmt.Sprintf("%d", s.PlayerPeak24Hour)},
		[]string{"Player Peak All", fmt.Sprintf("%d", s.PlayerPeakAll)},
		append([]string{"Publishers"}, s.Publishers...),
		[]string{"Release Date", s.ReleaseDate.String()},
		[]string{"Reviews All Count", fmt.Sprintf("%d", s.ReviewsAllCount)},
		[]string{"Reviews All Sentiment", s.ReviewsAllSentiment},
		[]string{"Reviews Recent Count", fmt.Sprintf("%d", s.ReviewsRecentCount)},
		[]string{"Reviews Recent Sentiment", s.ReviewsRecentSentiment},
		append([]string{"Social Media"}, s.SocialMedia...),
		append([]string{"Tags"}, s.Tags...),
		[]string{"Timestamp", s.Timestamp.String()},
		[]string{"Trough Players", fmt.Sprintf("%d", s.TroughPlayers)},
		[]string{"Trough Players Date", s.TroughPlayersDate},
		[]string{"URL", s.URL},
		[]string{"Website", s.Website},
		[]string{"Years Since Release", fmt.Sprintf("%d", s.YearsSinceRelease)}}
	return writer.WriteAll(rows)
}

func writeSteamGameSummaryCSVDefault(s *SteamGameSummary) error {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fullpath := filepath.Join(user.HomeDir, "Desktop", "steambot", "games", s.Name)
	err = writeSteamGameSummaryCSV(fullpath, s)
	return err
}

func writeSteamGameSummary(fullpath string, s *SteamGameSummary) error {
	err := os.MkdirAll(fullpath, os.ModePerm)
	if err != nil {
		return err
	}
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	filename := fmt.Sprintf("summary-%s.json", strings.ToLower(s.Name))
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
