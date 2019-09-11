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

var mu = &sync.Mutex{}

var steamSearchQueryMap = &SteamSearchQueryMap{}

var scanner = bufio.NewScanner(os.Stdin)

var w = new(tabwriter.Writer).Init(os.Stdout, 0, 8, 0, '\t', 0)

var pID = os.Getpid()

var (
	flagFarm      = flag.Int("farm", -1, "-farm 1")
	flagPagesFrom = flag.Int("from", -1, "-from 1")
	flagPagesTo   = flag.Int("to", -1, "-to 2")
	flagPageQuery = flag.String("options", "", "-options 'tags=19' (default '')")
	flagRevisit   = flag.Int("revisit", -1, "-revisit (default -1)")
	flagSilent    = flag.Bool("silent", false, "-silent (default false)")
	flagVerbose   = flag.Bool("verbose", false, "-verbose (default false)")
	flagWrite     = flag.Int("write", -1, "-write 0 (default -1)")
)

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
	fmt.Println(fmt.Sprintf("[steam][%d]", pID), "farm strategy: (1/NIL)")
	return requestInt()
}

func requestPagesFrom() int {
	fmt.Println(fmt.Sprintf("[steam][%d]", pID), "search pages from", "\t", "->", "(MUST BE > 0)")
	return requestInt()
}

func requestPagesTo() int {
	fmt.Println(fmt.Sprintf("[steam][%d]", pID), "search pages to", "\t", "->", fmt.Sprintf("(MUST BE >= %d)", *flagPagesFrom))
	return requestInt()
}

func requestRevisitStrategy() int {
	fmt.Println(fmt.Sprintf("[steam][%d]", pID), "revisit page condition", "\t", "->", "(< 2 DONT REVISIT ALL)")
	return requestInt()
}

func requestVerbosity() bool {
	fmt.Println(fmt.Sprintf("[steam][%d]", pID), "be verbose", "\t", "->", "(YES/NO)")
	if ok := scanner.Scan(); ok != true {
		return false
	}
	t := strings.ToUpper(strings.TrimSpace(scanner.Text()))
	return (t == "Y" || t == "YE" || t == "YES" || t == "1" || t == "OK")
}

func requestWriteStrategry() int {
	fmt.Println(fmt.Sprintf("[steam][%d]", pID), "write document condition", "\t", "->", "(< 4 DONT WRITE ALL)")
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
	fmt.Println(fmt.Sprintf("[steam][%d]", pID), "use filters", "\t", "->", "(YES/NO)")
	if ok := scanner.Scan(); ok != true {
		return queryString
	}
	t := strings.ToUpper(scanner.Text())
	ok := (t == "Y" || t == "YE" || t == "YES" || t == "1" || t == "OK")
	if ok != true {
		return queryString
	}
	steamSearchQueryMap = NewSteamSearchQueryMap(s)
	fmt.Println(fmt.Sprintf("[steam][%d]", pID), "show filters", "\t", "->", "(YES/NO)")
	if ok := scanner.Scan(); ok != true {
		return queryString
	}
	t = strings.ToUpper(scanner.Text())
	ok = (t == "Y" || t == "YE" || t == "YES" || t == "1" || t == "OK")
	if ok {
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
	fmt.Println(fmt.Sprintf("[steam][%d]", pID), "search filters", "\t", "->", "(SEPARATE FILTER USING SPACES)")
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

	if *flagPagesFrom == -1 && *flagSilent != true {
		*flagPagesFrom = requestPagesFrom()
	}

	if *flagPagesTo == -1 && *flagSilent != true {
		*flagPagesTo = requestPagesTo()
	}

	if *flagPageQuery == "" && *flagSilent != true {
		*flagPageQuery = requestPageQuery()
	}

	if *flagPagesFrom <= 0 {
		*flagPagesFrom = 1
	}

	if *flagPagesTo <= 0 {
		*flagPagesTo = 1
	}

	if ok := *flagPagesFrom > *flagPagesTo; ok {
		*flagPagesTo, *flagPagesFrom = *flagPagesFrom, *flagPagesTo
	}

	if *flagRevisit == -1 && *flagSilent != true {
		*flagRevisit = requestRevisitStrategy()
	}

	if *flagWrite == -1 && *flagSilent != true {
		*flagWrite = requestWriteStrategry()
	}

	if *flagFarm == -1 && *flagSilent != true {
		*flagFarm = requestFarmStrategy()
	}

	if *flagSilent != true && *flagVerbose == false {
		*flagVerbose = requestVerbosity()
	}

	switch *flagFarm {
	case 1:
		args := []string{
			"-silent",
			"-from",
			fmt.Sprintf("%d", (*flagPagesTo/2)+1),
			"-to",
			fmt.Sprintf("%d", *flagPagesTo),
			"-options",
			fmt.Sprintf("%s", *flagPageQuery),
			"-revisit",
			fmt.Sprintf("%d", *flagRevisit),
			"-write",
			fmt.Sprintf("%d", *flagWrite)}
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
	}

	steamerLog := &SteamerLog{
		PagesFrom: *flagPagesFrom,
		PagesTo:   *flagPagesTo,
		PagesOK:   &SteamerLogPageOK{},
		TimeStart: time.Now()}

	steamerSummary := &SteamerSummary{
		Developers: make(map[string][]string),
		Games:      0,
		Genres:     make(map[string]int),
		PagesFrom:  *flagPagesFrom,
		PagesTo:    *flagPagesTo,
		Publishers: make(map[string][]string),
		Sentiments: make(map[string]int)}

	steamGameSummaryCSV := []SteamSummaryCSV{}

	URL := fmt.Sprintf("%s?", steamSearchURL)

	*flagRevisit = int(math.Abs(float64(*flagRevisit)))

	var farmStrategy string
	switch *flagFarm {
	case 1:
		farmStrategy = "EVEN"
	default:
		farmStrategy = "NONE"
	}

	var revisitStrategy string
	switch *flagRevisit {
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

	fmt.Fprintln(w, fmt.Sprintf("[steam][%d]", pID), "farm", "\t", "->", farmStrategy)

	fmt.Fprintln(w, fmt.Sprintf("[steam][%d]", pID), "revisit", "\t", "->", revisitStrategy)

	fmt.Fprintln(w, fmt.Sprintf("[steam][%d]", pID), "write", "\t", "->", writeStrategy)

	fmt.Fprintln(w, fmt.Sprintf("[steam][%d]", pID), "timeStart", "\t", "->", steamerLog.TimeStart)

	w.Flush()

	if ok := len(*flagPageQuery) > 0; ok {
		URL = fmt.Sprintf("%s%s&", URL, *flagPageQuery)
	}

	for i := *flagPagesFrom; i <= *flagPagesTo; i++ {
		wg.Add(1)
		go func(client *http.Client, URL string) {
			defer wg.Done()
			revisit := *flagRevisit > 0
			onGetSteamGameAbbreviation(client, URL, revisit,
				func(s *Snapshot) {
					if *flagWrite > 0 {
						wg.Add(1)
						go func(s *Snapshot) {
							defer wg.Done()
							writeSnapshotDefault(s)
						}(s)
					}
					if *flagVerbose {
						fmt.Println("URL", "\t", "->", "[PAGE]", URL)
					}
				},
				func(s *SteamGameAbbreviation) {
					if *flagWrite >= 1 {
						writeSteamGameAbbreviationDefault(s)
					}
					wg.Add(1)
					go func(client *http.Client, URL string) {
						defer wg.Done()
						revisit := *flagRevisit > 1

						onGetSteamGamePage(client, URL, revisit,
							func(s *Snapshot) {
								if *flagWrite > 0 {
									wg.Add(1)
									go func(s *Snapshot) {
										defer wg.Done()
										writeSnapshotDefault(s)
									}(s)
								}
								if *flagVerbose {
									fmt.Println("URL", "\t", "->", "[GAME]", URL)
								}
							},
							func(s *SteamGamePage) {
								if *flagWrite >= 2 {
									writeSteamGamePageDefault(s)
								}
								wg.Add(1)
								go func(client *http.Client, URL string, steamGamePage *SteamGamePage) {
									defer wg.Done()
									revisit := *flagRevisit > 2
									onGetSteamChartPage(client, URL, revisit,
										func(s *Snapshot) {
											if *flagWrite > 0 {
												wg.Add(1)
												go func(s *Snapshot) {
													defer wg.Done()
													writeSnapshotDefault(s)
												}(s)
											}
											if *flagVerbose {
												fmt.Println("URL", "\t", "->", "[CHART]", URL)
											}
										},
										func(s *SteamChartPage) {
											steamGameSummary := NewSteamGameSummary(steamGamePage, s)
											if *flagWrite >= 3 {
												writeSteamChartPageDefault(s)
											}
											if *flagWrite >= 4 {
												writeSteamGameSummaryDefault(steamGameSummary)
											}
											steamGameSummaryCSV = append(steamGameSummaryCSV, NewSteamSummaryCSV(steamGameSummary))
										},
										func(e error) {

										})
								}(client, fmt.Sprintf("https://steamcharts.com/app/%d", s.AppID), s)

								mu.Lock()
								steamerSummary.Games = steamerSummary.Games + 1

								for _, x := range s.Developers {
									if m, ok := steamerSummary.Developers[x.Name]; !ok {
										steamerSummary.Developers[x.Name] = append(m, s.Title)
									} else {
										steamerSummary.Developers[x.Name] = []string{s.Title}
									}
								}
								for _, x := range s.Genres {
									if n, ok := steamerSummary.Genres[x.Name]; !ok {
										steamerSummary.Genres[x.Name] = n + 1
									} else {
										steamerSummary.Genres[x.Name] = 1
									}
								}
								for _, x := range s.Publishers {
									if m, ok := steamerSummary.Publishers[x.Name]; !ok {
										steamerSummary.Publishers[x.Name] = append(m, s.Title)
									} else {
										steamerSummary.Publishers[x.Name] = []string{s.Title}
									}
								}
								mu.Unlock()
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
	fmt.Fprintln(w, fmt.Sprintf("[steam][%d]", pID), "timeEnd", "\t", "->", steamerLog.TimeEnd)
	steamerLog.TimeDuration = steamerLog.TimeEnd.Sub(steamerLog.TimeStart)
	writeSteamerLogDefault(steamerLog)
	fmt.Fprintln(w, fmt.Sprintf("[steam][%d]", pID), "timeDuration", "\t", "->", steamerLog.TimeDuration)
	w.Flush()
	writeSteamerSummaryDefault(steamerSummary)
	filename := fmt.Sprintf("summary-%d-%d", *flagPagesFrom, *flagPagesTo)
	fmt.Println(writeSteamSummaryCSVDefault(filename, &steamGameSummaryCSV))
	time.Sleep(time.Second)
}
