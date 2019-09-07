package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const steamSearchURL string = "https://store.steampowered.com/search/"

var client = &http.Client{Timeout: time.Second * 10}

var wg = &sync.WaitGroup{}

var steamSearchQueryMap = &SteamSearchQueryMap{}

var scanner = bufio.NewScanner(os.Stdin)

func requestQueryFrom() int {
	var n int
	fmt.Println("search from:")
	if ok := scanner.Scan(); ok != true {
		return n
	}
	n, err := strconv.Atoi(scanner.Text())
	if ok := err == nil; ok != true {
		return n
	}
	return n
}

func requestQueryTo() int {
	var n int
	fmt.Println("search to:")
	if ok := scanner.Scan(); ok != true {
		return n
	}
	n, err := strconv.Atoi(scanner.Text())
	if ok := err == nil; ok != true {
		return n
	}
	return n
}

func requestQueryOptions() string {
	var queryString string
	req, err := http.NewRequest(http.MethodGet, steamSearchURL, nil)
	if err != nil {
		return queryString
	}
	res, err := client.Do(req)
	if err != nil {
		return queryString
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return queryString
	}
	s := doc.Find("div.tab_filter_control[data-param]")
	if ok := (s.Length() > 0); ok != true {
		return queryString
	}
	fmt.Println("enter filters:")
	steamSearchQueryMap = NewSteamSearchQueryMap(s)
	if ok := scanner.Scan(); ok != true {
		return queryString
	}
	querySet := map[string][]string{}
	for _, dataLoc := range strings.Split(strings.ToUpper(scanner.Text()), " ") {
		steamSearchKeyValue, ok := steamSearchQueryMap.Get(dataLoc)
		switch ok {
		case true:
			if _, ok := querySet[steamSearchKeyValue.Key]; !ok {
				querySet[steamSearchKeyValue.Key] = []string{}
			}
			s := querySet[steamSearchKeyValue.Key]
			s = append(s, steamSearchKeyValue.Value)
			querySet[steamSearchKeyValue.Key] = s
		default:
		}
	}
	queryQueue := []string{}
	for key, value := range querySet {
		queryQueue = append(queryQueue, fmt.Sprintf("%s=%s", key, strings.Join(value, "%2C")))
	}
	queryString = strings.Join(queryQueue, "&")
	return queryString
}

func main() {
	queryString := requestQueryOptions()
	i := requestQueryFrom()
	n := requestQueryTo()
	if i == 0 {
		i = 1
	}
	if n == 0 {
		i = 1
	}
	if i > n {
		i, n = n, i
	}
	steamerLog := &SteamerLog{
		PagesFrom: i,
		PagesTo:   n,
		PagesOK:   &SteamerLogPageOK{},
		TimeStart: time.Now()}
	fmt.Println("timeStart", "\t", "->", steamerLog.TimeStart)
	URL := fmt.Sprintf("%s?", steamSearchURL)
	if ok := len(queryString) > 0; ok {
		URL = fmt.Sprintf("%s%s&", URL, queryString)
	}
	for i := 1; i <= n; i++ {
		URL = fmt.Sprintf("%spage=%d", URL, i)
		fmt.Println(URL)
		wg.Add(1)
		go func(client *http.Client, URL string) {
			defer wg.Done()
			onGetSteamGameAbbreviation(client, URL,
				func(s *Snapshot) {
					writeSnapshotDefault(s)
				},
				func(s *SteamGameAbbreviation) {

					writeSteamGameAbbreviationDefault(s)

					wg.Add(1)
					go func(client *http.Client, URL string) {
						defer wg.Done()
						onGetSteamGamePage(client, URL,
							func(s *Snapshot) {
								writeSnapshotDefault(s)
							},
							func(s *SteamGamePage) {

								writeSteamGamePageDefault(s)

								wg.Add(1)
								go func(client *http.Client, URL string) {
									defer wg.Done()
									onGetSteamChartPage(client, URL,
										func(s *Snapshot) {
											writeSnapshotDefault(s)
										},
										func(s *SteamChartPage) {

											writeSteamChartPageDefault(s)
										},
										func(e error) {

										})
								}(client, fmt.Sprintf("https://steamcharts.com/app/%d", s.AppID))
							},
							func(e error) {
							})
					}(client, s.URL)
				},
				func(e error) {
				})
		}(client, URL)
	}
	wg.Wait()
	steamerLog.TimeEnd = time.Now()
	fmt.Println("timeEnd", "\t", "->", steamerLog.TimeEnd)
	steamerLog.TimeDuration = steamerLog.TimeEnd.Sub(steamerLog.TimeStart)
	writeSteamerLogDefault(steamerLog)
	fmt.Println("timeDuration", "\t", "->", steamerLog.TimeDuration)
}
