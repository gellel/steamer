package main

import (
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type SteamChartGameGrowth struct {
	Month          string  `json:"month"`
	PlayersAverage float64 `json:"players_average"`
	PlayersPeak    int     `json:"players_peak"`
	Gain           int     `json:"gain"`
	GainPercentage float64 `json:"gain_percentage"`
}

func NewSteamChartGameGrowth(s *goquery.Selection) SteamChartGameGrowth {
	return SteamChartGameGrowth{
		Gain:           scrapeSteamChartGain(s),
		GainPercentage: scrapeSteamChartGainPercentage(s),
		Month:          scrapeSteamChartMonth(s),
		PlayersAverage: scrapeSteamChartPlayersAverage(s),
		PlayersPeak:    scrapeSteamChartPlayersPeak(s)}
}

func scrapeSteamChartGain(s *goquery.Selection) int {
	n, err := strconv.Atoi(s.Find("td:nth-child(3)").Text())
	if err != nil {
		return -1
	}
	return n
}
func scrapeSteamChartGainPercentage(s *goquery.Selection) float64 {
	f, err := strconv.ParseFloat(s.Find("td:nth-child(4)").Text(), 64)
	if err != nil {
		return -1
	}
	return f
}

func scrapeSteamChartMonth(s *goquery.Selection) string {
	return strings.TrimSpace(s.Find("td.month-cell").Text())
}

func scrapeSteamChartPlayersAverage(s *goquery.Selection) float64 {
	f, err := strconv.ParseFloat(s.Find("td:nth-child(2)").Text(), 64)
	if err != nil {
		return -1
	}
	return f
}

func scrapeSteamChartPlayersPeak(s *goquery.Selection) int {
	n, err := strconv.Atoi(s.Find("td:nth-child(5)").Text())
	if err != nil {
		return -1
	}
	return n
}
