package main

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type SteamPageGameAggregateReview struct {
	Count      int    `json:"count"`
	Percentage int    `json:"percentage"`
	Sentiment  string `json:"sentiment"`
}

func NewSteamPageGameAggregateReview(s *goquery.Selection) SteamPageGameAggregateReview {
	var (
		count      int
		sentiment  string
		percentage int
	)
	s.First().Each(func(i int, s *goquery.Selection) {
		s.Find("span.game_review_summary").First().Each(func(i int, s *goquery.Selection) {
			sentiment = strings.TrimSpace(s.Text())
		})
		s.Find("span.responsive_hidden").First().Each(func(i int, s *goquery.Selection) {
			substring := regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(s.Text(), "")
			n, err := strconv.Atoi(substring)
			if err != nil {
				return
			}
			count = n
		})
		s.Find("span.nonresponsive_hidden").First().Each(func(i int, s *goquery.Selection) {
			substring := regexp.MustCompile(`\s(\d+%)`).FindString(s.Text())
			substring = regexp.MustCompile(`[^0-9]`).ReplaceAllString(substring, "")
			n, err := strconv.Atoi(substring)
			if err != nil {
				return
			}
			percentage = n
		})
	})
	return SteamPageGameAggregateReview{
		Count:      count,
		Sentiment:  sentiment,
		Percentage: percentage}
}
