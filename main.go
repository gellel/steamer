package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var wg sync.WaitGroup

var mu sync.Mutex

var isZero = false

var flagSteamPageSearchFrom = flag.Int("search-from", -1, "search-from=1")

var flagSteamPageSearchTo = flag.Int("search-to", -1, "search-to=1")

var flagSteamPageTerminateOnZero = flag.Bool("terminate-zero", false, "terminate-zero=true")

var stdinScanner = bufio.NewScanner(os.Stdin)

var steamerLog *SteamerLog

func mainRequestStdinSearchFrom() int {
	if i := *flagSteamPageSearchFrom; i != -1 {
		return 1
	}
	log.Println("[STEAMER] SCRAPE FROM")
	if ok := stdinScanner.Scan(); ok != true {
		return 1
	}
	i, err := strconv.Atoi(stdinScanner.Text())
	if err != nil {
		return 1
	}
	return i
}

func mainRequestStdinSearchTo() int {
	if i := *flagSteamPageSearchTo; i != -1 {
		return 1
	}
	log.Println("[STEAMER] SCRAPE TO")
	if ok := stdinScanner.Scan(); ok != true {
		return 1
	}
	i, err := strconv.Atoi(stdinScanner.Text())
	if err != nil {
		return 1
	}
	return i
}

func mainRequestStdinTerminateOnZero() bool {
	log.Println("[STEAMER] TERMINATE ON NIL RECORDS")
	if ok := stdinScanner.Scan(); ok != true {
		return *flagSteamPageTerminateOnZero
	}
	b, err := strconv.ParseBool(stdinScanner.Text())
	if err != nil {
		return *flagSteamPageTerminateOnZero
	}
	return b
}

func mainRequestSteamGameChart(c chan *Snapshot, client *http.Client, URL string) {
	defer wg.Done()
	go chanSnapshot(c, client, http.MethodGet, URL)
	steamGameChartSnapshot := <-c
	if steamGameChartSnapshot.StatusCode != http.StatusOK {
		log.Println(fmt.Sprintf("[STEAMER] CHART %s FAILED. ERR(s): HTTP.STATUS %s", URL, steamGameChartSnapshot.Status))
		wg.Done()
		return
	}
	if steamGameChartSnapshot.document == nil {
		wg.Done()
		return
	}
	log.Println(fmt.Sprintf("[STEAMER] CHART %s IS OK!", steamGameChartSnapshot.URL))
	steamChartPage := NewSteamChartPage(steamGameChartSnapshot.document.Find("html"))
	wg.Add(1)
	go chanWriteSteamChartPageDefault(steamChartPage, URL)
}

func mainRequestSteamGamePage(c chan *Snapshot, client *http.Client, URL string, i int) {
	defer wg.Done()
	wg.Add(1)
	go chanSnapshot(c, client, http.MethodGet, URL)
	steamGameSnapshot := <-c
	if steamGameSnapshot.StatusCode != http.StatusOK {
		log.Println(fmt.Sprintf("[STEAMER] PAGE %s FAILED. ERR(s): HTTP.STATUS %s", URL, steamGameSnapshot.Status))
		return
	}
	if steamGameSnapshot.document == nil {
		return
	}
	log.Println(fmt.Sprintf("[STEAMER] PAGE %s IS OK!", steamGameSnapshot.URL))
	steamGamePage := NewSteamGamePage(steamGameSnapshot.document.Find("html"))
	wg.Add(1)
	go chanWriteSteamGamePageDefault(steamGamePage, URL)
	steamGameChartChan := make(chan *Snapshot)
	wg.Add(1)
	go mainRequestSteamGameChart(steamGameChartChan, client, fmt.Sprintf("https://steamcharts.com/app/%d/", steamGamePage.AppID))
}

func getSteamGameAbbreviation(c chan *SteamGameAbbreviation, client *http.Client, URL string, i int) {
	snapshot := getSnapshot(client, http.MethodGet, URL)
	if snapshot.StatusCode != http.StatusOK || snapshot.document == nil {
		log.Println("[STEAMER] [GAME.ABBR] ERROR: HTML NOT OK")
		if steamerLog != nil {
			mu.Lock()
			steamerLog.PagesOK.Add(i, false)
			mu.Unlock()
		}
		return
	}
	s := snapshot.document.Find("a.search_result_row[href]")
	if ok := s.Length() > 0; ok != true {
		log.Println("[STEAMER] [GAME.ABBR] WARN: UNABLE TO FIND ANY SEARCH RESULTS")
		isZero = true
		return
	}
	log.Println("[STEAMER] [GAME.ABBR] OK!")
	if err := writeSnapshotDefault(snapshot); err != nil {
		log.Println("[STEAMER] [GAME.ABBR] ERROR: SNAPSHOT WRITE NOT OK")
	}
	s.Each(func(i int, s *goquery.Selection) {
		c <- NewSteamGameAbbreviation(s)
	})
}

func handleSteamGameAbbreviation(s *SteamGameAbbreviation) {
	defer wg.Done()
	if err := writeSteamGameAbbreviationDefault(s); err != nil {
		log.Println("[STEAMER] [GAME.ABBR] ERROR: ABBR WRITE NOT OK")
	}
	if ok := s.AppID > -1; ok != true {
		log.Println("[STEAMER] [GAME.ABBR] ERROR: APPID NOT OK")
		return
	}
}

func reqSteamGameAbbreviation(client *http.Client, URL string, i int) {
	defer wg.Done()
	c := make(chan *SteamGameAbbreviation)
	go getSteamGameAbbreviation(c, client, URL, i)
	steamGameAbbreviation := <-c
	go handleSteamGameAbbreviation(steamGameAbbreviation)
	fmt.Println(steamGameAbbreviation)

}

func mainRequestSteamGameAbbreviation(c chan *Snapshot, client *http.Client, URL string, i int) {
	defer wg.Done()
	wg.Add(1)
	go chanSnapshot(c, client, http.MethodGet, URL)
	steamPageSnapshot := <-c
	if steamPageSnapshot.StatusCode != http.StatusOK {
		log.Println(fmt.Sprintf("[STEAMER] ABBR %s FAILED. ERR(s): HTTP.STATUS %s | ", URL, steamPageSnapshot.Status))
		return
	}
	if steamPageSnapshot.document == nil {
		return
	}
	log.Println(fmt.Sprintf("[STEAMER] ABBR %s IS OK!", steamPageSnapshot.URL))
	if steamerLog != nil {

	}
	s := steamPageSnapshot.document.Find("a.search_result_row[href]")
	if s.Length() == 0 {
		isZero = true
		return
	}
	steamGameSnapshotChan := make(chan *Snapshot, s.Length())
	s.Each(func(i int, s *goquery.Selection) {
		steamGameAbbreviation := NewSteamGameAbbreviation(s)
		if steamPageSnapshot.URL == "NIL" {
			return
		}
		wg.Add(1)
		go chanWriteSteamGameAbbreviationDefault(steamGameAbbreviation, URL)
		if steamGameAbbreviation.AppID == -1 {
			log.Println(fmt.Sprintf("[STEAMER] ABBR %s FAILED. ERR(s): APP-ID IS NEGATIVE", URL))
			return
		}
		wg.Add(1)
		go mainRequestSteamGamePage(steamGameSnapshotChan, client, steamGameAbbreviation.URL, i)
	})
}

func main() {
	flag.Parse()
	client := &http.Client{Timeout: time.Second * 10}
	i := mainRequestStdinSearchFrom()
	n := mainRequestStdinSearchTo()
	breakOnNoPageRecords := mainRequestStdinTerminateOnZero()
	if i == n {
		n = i + 1
	}
	req, err := http.NewRequest(http.MethodGet, "https://store.steampowered.com/search/", nil)
	if err != nil {
		log.Println(fmt.Sprintf("[STEAMER] REQ FAILED. ERR(s): %v", err))
	}
	res, err := client.Do(req)
	if err != nil {
		log.Println(fmt.Sprintf("[STEAMER] RES FAILED. ERR(s): %v", err))
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		log.Println(fmt.Sprintf("[STEAMER] DOC FAILED. ERR(s): %v", err))
	}
	if doc != nil {
		fmt.Println(NewSteamSearchQueryCategoryMap(doc.Find("#additional_search_options div.tab_filter_control")))
	}
	//steamPageSnapshotChan := make(chan *Snapshot, n)
	steamerLog = &SteamerLog{
		PagesFrom:     i,
		PagesTo:       n - 1,
		PagesOK:       &SteamerLogPageOK{},
		TerminateZero: breakOnNoPageRecords,
		TimeStart:     time.Now()}
	log.Println(fmt.Sprintf("[STEAMER] COLLECTING FROM %d TO %d", i, n))
	for i := i; i < n; i++ {
		if breakOnNoPageRecords == true {
			if isZero == true {
				break
			}
		}
		wg.Add(1)
		URL := fmt.Sprintf("store.steampowered.com/search/?page=%d", i)
		//go mainRequestSteamGameAbbreviation(steamPageSnapshotChan, client, URL, i)
		go reqSteamGameAbbreviation(client, URL, i)
	}
	wg.Wait()
	steamerLog.TimeEnd = time.Now()
	steamerLog.TimeDuration = steamerLog.TimeEnd.Sub(steamerLog.TimeStart)
	err = writeSteamerLogDefault(steamerLog)
	if err != nil {
		log.Println(fmt.Sprintf("[STEAMER] LOG FAILED. ERR(s): CANNOT WRITE %s", err))
	}
}
