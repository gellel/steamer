package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	colorDebug   = "\033[0;36m%s\033[0m"
	colorError   = "\033[1;31m%s\033[0m"
	colorInfo    = "\033[1;34m%s\033[0m"
	colorNotice  = "\033[1;36m%s\033[0m"
	colorWarning = "\033[1;33m%s\033[0m"
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

	fmt.Println("[STEAMER START]")

	flagSilent := flag.Bool("silent", false, "-silent (default false)")

	flagPagesFrom := flag.Int("from", -1, "-from 1")

	flagPagesTo := flag.Int("to", -1, "-to 2")

	flagPageQuery := flag.String("options", "", "-options 'a b c' (default '')")

	flagFarm := flag.Int("farm", -1, "-farm 1")

	flagVerbose := flag.Bool("verbose", false, "-verbose (default false)")

	flag.Parse()

	if ok := flag.Parsed(); ok != true {
		return
	}

	if *flagPagesFrom == -1 {
		if *flagSilent == true {
			*flagPagesFrom = 1
		} else {
			*flagPagesFrom = requestPagesFrom()
		}
	}

	if *flagPagesTo == -1 {
		if *flagSilent == true {
			*flagPagesTo = *flagPagesFrom + 1
		} else {
			*flagPagesTo = requestPagesTo()
		}
	}

	if *flagPageQuery == "" {
		if *flagSilent != true {
			*flagPageQuery = requestPageQuery()
		}
	}

	if ok := *flagPagesFrom > *flagPagesTo; ok {
		*flagPagesTo, *flagPagesFrom = *flagPagesFrom, *flagPagesTo
	}

	if *flagFarm == -1 {
		if *flagSilent != true {
			*flagFarm = requestFarmStrategy()
		}
	}

	switch *flagFarm {
	case 1:
		args := []string{"-silent", "-from", fmt.Sprintf("%d", (*flagPagesTo/2)+1), "-to", fmt.Sprintf("%d", *flagPagesTo), "-options", *flagPageQuery}
		cmd := exec.Command(os.Args[0], args...)
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd.Run()
		}()
		*flagPagesTo = (*flagPagesTo / 2)
	default:
	}

	steamerLog := &SteamerLog{
		PagesFrom: *flagPagesFrom,
		PagesTo:   *flagPagesTo,
		PagesOK:   &SteamerLogPageOK{},
		TimeStart: time.Now()}

	fmt.Println("timeStart", "\t", "->", steamerLog.TimeStart)

	URL := fmt.Sprintf("%s?", steamSearchURL)

	fmt.Println("flagPageQuery", "\t", "->", *flagPageQuery)

	if ok := len(*flagPageQuery) > 0; ok {
		URL = fmt.Sprintf("%s%s&", URL, *flagPageQuery)
	}

	for i := *flagPagesFrom; i <= *flagPagesTo; i++ {
		wg.Add(1)
		go func(client *http.Client, URL string) {
			fmt.Println("URL", "\t", "->", URL)
			defer wg.Done()
			onGetSteamGameAbbreviation(client, URL,
				func(s *Snapshot) {
					writeSnapshotDefault(s)

					if *flagVerbose {
					}
				},
				func(s *SteamGameAbbreviation) {

					writeSteamGameAbbreviationDefault(s)

					wg.Add(1)
					go func(client *http.Client, URL string) {
						defer wg.Done()
						onGetSteamGamePage(client, URL,
							func(s *Snapshot) {
								writeSnapshotDefault(s)

								if *flagVerbose {

								}
							},
							func(s *SteamGamePage) {

								writeSteamGamePageDefault(s)

								wg.Add(1)
								go func(client *http.Client, URL string) {
									defer wg.Done()
									onGetSteamChartPage(client, URL,
										func(s *Snapshot) {
											writeSnapshotDefault(s)

											if *flagVerbose {

											}
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
	fmt.Println("[STEAMER END]")
	time.Sleep(time.Second)
}
