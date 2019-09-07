package main

import (
	"bufio"
	"flag"
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

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func requestInt() int {
	if ok := scanner.Scan(); ok != true {
		return 0
	}
	n, err := strconv.Atoi(scanner.Text())
	if ok := err == nil; ok != true {
		return 0
	}
	if n < 0 {
		return 0
	}
	return n
}

func requestFarmStrategy() int {
	fmt.Println("farm strategy:")
	return requestInt()
}

func requestPagesFrom() int {
	fmt.Println("search from:")
	return requestInt()
}

func requestPagesTo() int {
	fmt.Println("search to:")
	return requestInt()
}

func requestPageQuery() string {
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

	silent := flag.Bool("silent", false, "-silent (default false)")

	pagesFrom := flag.Int("from", -1, "-from 1")

	pagesTo := flag.Int("to", -1, "-to 2")

	pageQuery := flag.String("options", "", "-options 'a b c' (default '')")

	//farm := flag.Bool("farm", false, "")

	//farmStrategy := flag.Int("farm-strategy", -1, "")

	flag.Parse()

	fmt.Println(*silent)

	if ok := flag.Parsed(); ok != true {
		return
	}
	if *pagesFrom == -1 {
		if *silent == true {
			*pagesFrom = 1
		} else {
			*pagesFrom = requestPagesFrom()
		}
	}
	if *pagesTo == -1 {
		if *silent == true {
			*pagesTo = *pagesFrom + 1
		} else {
			*pagesTo = requestPagesTo()
		}
	}
	if *pageQuery == "" {
		if *silent != true {
			*pageQuery = requestPageQuery()
		}
	}

	if ok := *pagesFrom > *pagesTo; ok {
		*pagesTo, *pagesFrom = *pagesFrom, *pagesTo
	}

	steamerLog := &SteamerLog{
		PagesFrom: *pagesFrom,
		PagesTo:   *pagesTo,
		PagesOK:   &SteamerLogPageOK{},
		TimeStart: time.Now()}

	fmt.Println("timeStart", "\t", "->", steamerLog.TimeStart)
	URL := fmt.Sprintf("%s?", steamSearchURL)

	fmt.Println("pageQuery", "\t", "->", *pageQuery)

	if ok := len(*pageQuery) > 0; ok {
		URL = fmt.Sprintf("%s%s&", URL, *pageQuery)
	}
	for i := *pagesFrom; i <= *pagesTo; i++ {
		wg.Add(1)
		go func(client *http.Client, URL string) {
			fmt.Println("URL", "\t", "->", URL)
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
		}(client, fmt.Sprintf("%spage=%d", URL, i))
	}
	wg.Wait()
	steamerLog.TimeEnd = time.Now()
	fmt.Println("timeEnd", "\t", "->", steamerLog.TimeEnd)
	steamerLog.TimeDuration = steamerLog.TimeEnd.Sub(steamerLog.TimeStart)
	writeSteamerLogDefault(steamerLog)
	fmt.Println("timeDuration", "\t", "->", steamerLog.TimeDuration)
}
