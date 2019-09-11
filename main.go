package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
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

var w = new(tabwriter.Writer).Init(os.Stdout, 0, 8, 0, '\t', 0)

var (
	flagFarm         = flag.Int("farm", -1, "-farm 1")
	flagPagesFrom    = flag.Int("from", -1, "-from 1")
	flagPagesTo      = flag.Int("to", -1, "-to 2")
	flagPageQuery    = flag.String("options", "", "-options 'a b c' (default '')")
	flagSilent       = flag.Bool("silent", false, "-silent (default false)")
	flagThread       = flag.Int("thread", 1, "-thread (default 1)")
	flagRevisitFound = flag.Int("revisit", -1, "-revisit (default -1)")
	flagVerbose      = flag.Bool("verbose", false, "-verbose (default false)")
	flagWrite        = flag.Int("write", -1, "-write 0 (default -1)")
)

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
	fmt.Println("farm strategy: (1/NIL)")
	return requestInt()
}

func requestPagesFrom() int {
	fmt.Println("search pages from", "\t", "->", "(MUST BE > 0)")
	return requestInt()
}

func requestPagesTo() int {
	fmt.Println("search pages from", "\t", "->", fmt.Sprintf("(MUST BE > %d)", *flagPagesFrom))
	return requestInt()
}

func requestRevisitStrategy() int {
	fmt.Println("revisit page condition", "\t", "->", "(< 2 DONT REVISIT ALL)")
	return requestInt()
}

func requestWriteStrategry() int {
	fmt.Println("write document condition", "\t", "->", "(< 3 DONT WRITE ALL)")
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
	steamSearchQueryMap = NewSteamSearchQueryMap(s)
	if ok := *flagSilent == false; ok {
		fmt.Println("show filters: (YES/NO)")
		if ok := scanner.Scan(); ok {
			switch strings.ToUpper(scanner.Text()) {
			case "Y", "YE", "YES", "YSE", "1", "OK":
				mapLen := len(*steamSearchQueryMap)
				queryMapKeys := make([]string, mapLen)
				i := 0
				for key := range *steamSearchQueryMap {
					queryMapKeys[i] = key
					i = i + 1
				}
				sort.Strings(queryMapKeys)
				for _, key := range queryMapKeys {
					fmt.Println(key)
				}
			}
		}
	}
	fmt.Println("search filters", "\t", "->", "(SEPARATE FILTER USING SPACES)")
	if ok := scanner.Scan(); ok != true {
		return queryString
	}
	querySet := map[string][]string{}
	for _, dataLoc := range strings.Split(strings.ToUpper(scanner.Text()), " ") {
		steamSearchKeyValue, ok := steamSearchQueryMap.Get(dataLoc)
		var key string
		var value string
		if ok {
			key = steamSearchKeyValue.Key
			value = steamSearchKeyValue.Value
		} else {
			key = "term"
			value = dataLoc
		}
		if _, ok := querySet[key]; !ok {
			querySet[key] = []string{}
		}
		s := querySet[key]
		s = append(s, value)
		querySet[key] = s
	}
	queryQueue := []string{}
	for key, value := range querySet {
		queryQueue = append(queryQueue, fmt.Sprintf("%s=%s", key, strings.Join(value, "%2C")))
	}
	queryString = strings.Join(queryQueue, "&")
	return queryString
}

func main() {

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

	if *flagRevisitFound == -1 {
		if *flagSilent != true {
			*flagRevisitFound = requestRevisitStrategy()
		} else {
			*flagRevisitFound = 0
		}
	}

	if *flagWrite == -1 {
		if *flagSilent != true {
			*flagWrite = requestWriteStrategry()
		}
	}

	if *flagFarm == -1 {
		if *flagSilent != true {
			*flagFarm = requestFarmStrategy()
		}
	}

	switch *flagFarm {
	case 1:
		args := []string{"-silent", "-from", fmt.Sprintf("%d", (*flagPagesTo/2)+1), "-to", fmt.Sprintf("%d", *flagPagesTo), "-options", *flagPageQuery, "-thread", fmt.Sprintf("%d", *flagThread+1)}
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

	URL := fmt.Sprintf("%s?", steamSearchURL)

	*flagRevisitFound = int(math.Abs(float64(*flagRevisitFound)))

	var revisitStrategy string
	switch *flagRevisitFound {
	case 0:
		revisitStrategy = "NONE"
	case 1:
		revisitStrategy = "PAGES"
	case 2:
		revisitStrategy = "PAGES + GAMES"
	default:
		revisitStrategy = "ALL"
	}
	var writeStrategy string
	switch *flagWrite {
	case -1:
		writeStrategy = "SUMMARY ONLY"
	case 0:
		writeStrategy = "LOGS + SUMMARY"
	case 1:
		writeStrategy = "LOGS + ABBR + SUMMARY"
	case 2:
		writeStrategy = "LOGS + ABBR + GAME + SUMMARY"
	default:
		writeStrategy = "ALL"
	}

	fmt.Fprintln(w, "revisit", "\t", "->", revisitStrategy)

	fmt.Fprintln(w, "write", "\t", "->", writeStrategy)

	fmt.Fprintln(w, "timeStart", *flagThread, "\t", "->", steamerLog.TimeStart)

	w.Flush()

	if ok := len(*flagPageQuery) > 0; ok {
		URL = fmt.Sprintf("%s%s&", URL, *flagPageQuery)
	}

	for i := *flagPagesFrom; i <= *flagPagesTo; i++ {
		wg.Add(1)
		go func(client *http.Client, URL string) {
			defer wg.Done()
			revisit := *flagRevisitFound > 0
			onGetSteamGameAbbreviation(client, URL, revisit,
				func(s *Snapshot) {
					if *flagWrite > 0 {
						writeSnapshotDefault(s)
					}
					if *flagVerbose {
						fmt.Println("URL", "\t", "->", "[PAGE]", URL)
					}
				},
				func(s *SteamGameAbbreviation) {
					if *flagWrite > 1 {
						writeSteamGameAbbreviationDefault(s)
					}
					wg.Add(1)
					go func(client *http.Client, URL string) {
						defer wg.Done()
						revisit := *flagRevisitFound > 1

						onGetSteamGamePage(client, URL, revisit,
							func(s *Snapshot) {
								if *flagWrite > 0 {
									writeSnapshotDefault(s)
								}
								if *flagVerbose {
									fmt.Println("URL", "\t", "->", "[GAME]", URL)
								}
							},
							func(s *SteamGamePage) {
								if *flagWrite > 2 {
									writeSteamGamePageDefault(s)
								}
								wg.Add(1)
								go func(client *http.Client, URL string, steamGamePage *SteamGamePage) {
									defer wg.Done()
									revisit := *flagRevisitFound > 2
									onGetSteamChartPage(client, URL, revisit,
										func(s *Snapshot) {
											if *flagWrite > 0 {
												writeSnapshotDefault(s)
											}
											if *flagVerbose {
												fmt.Println("URL", "\t", "->", "[CHART]", URL)
											}
										},
										func(s *SteamChartPage) {
											if *flagWrite > 3 {
												writeSteamChartPageDefault(s)
											}
											writeSteamGameSummaryDefault(NewSteamGameSummary(steamGamePage, s))
										},
										func(e error) {

										})
								}(client, fmt.Sprintf("https://steamcharts.com/app/%d", s.AppID), s)
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
	fmt.Fprintln(w, "timeEnd", *flagThread, "\t", "->", steamerLog.TimeEnd)
	steamerLog.TimeDuration = steamerLog.TimeEnd.Sub(steamerLog.TimeStart)
	writeSteamerLogDefault(steamerLog)
	fmt.Fprintln(w, "timeDuration", *flagThread, "\t", "->", steamerLog.TimeDuration)
	w.Flush()
	time.Sleep(time.Second)
}
