package main

import (
	"bufio"
	"fmt"
	"math"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/PuerkitoBio/goquery"
)

const attrDataParam string = "data-param"

const attrDataLoc string = "data-loc"

const attrDataValue string = "data-value"

const steamStoreSearchURL string = "https://store.steampowered.com/search/"

var wg sync.WaitGroup

var mutex = sync.RWMutex{}

var searchQueryCatalogue = map[string]map[string]string{}

var searchQueryReverse = map[string]string{}

var queryMap = map[string][]string{}

var maxPages int

var currentPage int

var writer = new(tabwriter.Writer).Init(os.Stdout, 0, 8, 0, '\t', 0)

var scanner = bufio.NewScanner(os.Stdin)

func main() {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		os.Exit(1)
	}
	termXY := strings.Split(strings.TrimSuffix(string(out), "\n"), " ")
	_, err = strconv.Atoi(termXY[0])
	if err != nil {
		os.Exit(1)
	}
	sY, err := strconv.Atoi(termXY[1])
	if err != nil {
		os.Exit(1)
	}
	r, err := regexp.Compile(`[^a-zA-Z0-9]+`)
	if err != nil {
		os.Exit(1)
	}
	// http.Get to base steam store
	resp, err := http.Get(steamStoreSearchURL)
	if err != nil {
		os.Exit(1)
	}
	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
	// tokenize http response.
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		os.Exit(1)
	}
	// find the html tokens that contain the steam filters.
	s := doc.Find(fmt.Sprintf("[%s]", attrDataLoc))
	// count the number of filters that were parsed.
	l := s.Length()
	if ok := l == 0; ok {
		os.Exit(1)
	}
	var x int
	c := make(chan *goquery.Selection, l)
	// iterate across the html tokens and attempt to find desired meta data
	s.Each(func(i int, s *goquery.Selection) {
		dataParam, exists := s.Attr(attrDataParam)
		if exists != false {
			// block the channel
			wg.Add(1)
			// lock the searchQueryCatalogue
			mutex.Lock()
			c <- s
			// process the current html token within the channel
			go func(c chan *goquery.Selection, dataParam string) {
				defer wg.Done()
				if _, ok := searchQueryCatalogue[dataParam]; !ok {
					searchQueryCatalogue[dataParam] = map[string]string{}
				}
				s := <-c
				queryKey, ok := s.Attr(attrDataLoc)
				if ok != true {
					mutex.Unlock()
					return
				}
				queryValue, ok := s.Attr(attrDataValue)
				if ok != true {
					mutex.Unlock()
					return
				}
				queryKey = strings.ToUpper(r.ReplaceAllString(queryKey, "-"))
				searchQueryCatalogue[dataParam][queryKey] = queryValue
				searchQueryReverse[queryKey] = dataParam
				mutex.Unlock()
				x = x + 1
			}(c, dataParam)
		}
	})
	wg.Wait()
	close(c)
	// confirm that the program found at least a single filter
	if ok := len(searchQueryCatalogue) > 0; !ok {
		os.Exit(1)
	}
	// confirm the program found the same number of filters as expected
	if ok := x == l; !ok {
		os.Exit(1)
	}
	var i int
	// write to os.STDOUT the available filters for the user to select
	for dataParam := range searchQueryCatalogue {
		fmt.Println(fmt.Sprintf("%v\t|%s", i, strings.ToUpper(dataParam)))
		i = i + 1
		fmt.Println("")
		for queryKey := range searchQueryCatalogue[dataParam] {
			fmt.Println(fmt.Sprintf("\t|%s", queryKey))
		}
		fmt.Println(strings.Repeat("-", sY))
	}
	fmt.Println("STEAMER.EXE\t>>>\tPLEASE INPUT ALL REQUIRED FILTERS:")
	// wait for user input to collect the desired number of search query filters.
	if ok := scanner.Scan(); !ok {
		os.Exit(1)
	}
	searchOptions := strings.ToUpper(scanner.Text())
	if ok := len(searchOptions) > 0; !ok {
		os.Exit(1)
	}
	r, err = regexp.Compile(`(\s{2,}|[,_|/\\]+)`)
	if err != nil {
		os.Exit(1)
	}
	searchOptions = r.ReplaceAllString(searchOptions, " ")
	userOptions := strings.Split(searchOptions, " ")
	// attempt to match the users search options back to the common queryMap.
	for _, optionKey := range userOptions {
		searchKey, ok := searchQueryReverse[optionKey]
		if ok != true {
			continue
		}
		searchValue, ok := searchQueryCatalogue[searchKey][optionKey]
		if ok != true {
			continue
		}
		if querySet, ok := queryMap[searchKey]; ok {
			queryMap[searchKey] = append(querySet, searchValue)
		} else {
			queryMap[searchKey] = []string{searchValue}
		}
	}
	queryStringSet := []string{}
	// build the steam store queryString for the next search action.
	for key, querySet := range queryMap {
		// concatenate the keyValue for each queryParam with the parentTag
		queryStringSet = append(queryStringSet, fmt.Sprintf("%s=%s", key, strings.Join(querySet, "%2C")))
	}
	// concatenate the steamStoreURL with the queryMap for upcoming search action.
	steamStoreSearchURLwithQuery := (steamStoreSearchURL + "?" + strings.Join(queryStringSet, "&"))
	// os.STDOUT program action.
	fmt.Println(fmt.Sprintf("STEAMER.EXE\t>>>\tSEARCHING: %s", steamStoreSearchURLwithQuery))
	// http.Get the steam store but using the formatted queryString.
	resp, err = http.Get(steamStoreSearchURLwithQuery)
	if err != nil {
		fmt.Println("STEAMER.EXE\t>>>\tCANNOT REACH STEAM STORE.")
		os.Exit(1)
	}
	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
	// tokenize http response.
	doc, err = goquery.NewDocumentFromResponse(resp)
	if err != nil {
		os.Exit(1)
	}
	// find the number of pages that are required to be processed
	s = doc.Find(".search_pagination_right>a:not(.pagebtn)")
	if ok := s.Length() > 0; !ok {
		fmt.Println("STEAMER.EXE\t>>>\tNO RESULTS WERE FOUND FOR THAT QUERY.")
		os.Exit(1)
	}
	s.Each(func(i int, s *goquery.Selection) {
		nthPage, err := strconv.Atoi(strings.TrimSpace(s.Text()))
		if err != nil {
			return
		}
		maxPages = int(math.Max(float64(nthPage), float64(maxPages)))
	})
	s = doc.Find("[data-ds-appid]")
	l = s.Length()
	if ok := l > 0; !ok {
		fmt.Println("STEAMER.EXE\t>>>\tPROGRAM COULD NOT FIND RECORDS TO PROCESS.")
		os.Exit(1)
	}
	// os.STDOUT program action.
	fmt.Println(fmt.Sprintf("STEAMER.EXE\t>>>\tPROGRAM FOUND A TOTAL OF %v PAGES TO PROCESS.", l))
	//for i := currentPage + 1; i <= maxPages; i++ {
	//fmt.Println(fmt.Sprintf("%s&page=%v", steamStoreSearchURLwithQuery, i))
	//}
	fmt.Println(strings.Repeat("-", sY))
	//c := make(chan *goquery.Selection, l)
	s.Each(func(i int, s *goquery.Selection) {
		// get the game title from the current html token
		gTitle := strings.TrimSpace(s.Find(".title").Text())
		// get the release date
		gReleaseDate := strings.TrimSpace(s.Find(".search_released").Text())
		// get the game sentiment
		gReviewSentiment := strings.TrimSpace(s.Find(".search_review_summary").AttrOr("data-tooltip-html", "NIL"))
		gReviewSentiment = strings.ReplaceAll(gReviewSentiment, "<br>", " ")
		// get the direct link to the game from the html token.
		gHref := s.AttrOr("href", "NIL")
		// get the appID
		gAppID := s.AttrOr("data-ds-appid", "NIL")
		// get the bundleID
		gBundleID := s.AttrOr("data-ds-bundleid", "NIL")
		// get the ctrlID
		gCtrlID := s.AttrOr("data-ds-crtrids", "NIL")
		// get the descID
		gDescID := s.AttrOr("data-ds-descids", "NIL")
		// get the package ID from the current html token
		gPkgID := s.AttrOr("data-ds-packageid", "NIL")
		// get the tagIDs from the html token
		gTagID := s.AttrOr("data-ds-tagids", "NIL")
		gPriceDiscount := strings.TrimSpace(s.Find(".search_discount>span").Text())
		gPriceCurrent := strings.TrimSpace(s.Find(".search_price").Text())
		if len(gPriceDiscount) == 0 {
			gPriceDiscount = "0%"
		}
		//, strings.Repeat("-", (sY-len(title)+2))
		fmt.Fprintln(writer, fmt.Sprintf("[%s]%s", strings.ToUpper(gTitle), strings.Repeat("-", (sY-len(gTitle)+2))))
		fmt.Fprintln(writer, fmt.Sprintf("RELEASE DATE\t|%s", gReleaseDate))
		fmt.Fprintln(writer, fmt.Sprintf("SENTIMENT\t|%s", gReviewSentiment))
		fmt.Fprintln(writer, fmt.Sprintf("HREF\t|%s", gHref))
		fmt.Fprintln(writer, fmt.Sprintf("ID-APP\t|%s", gAppID))
		fmt.Fprintln(writer, fmt.Sprintf("ID-BUNDLE\t|%s", gBundleID))
		fmt.Fprintln(writer, fmt.Sprintf("ID-CTRL\t|%s", gCtrlID))
		fmt.Fprintln(writer, fmt.Sprintf("ID-DESC\t|%s", gDescID))
		fmt.Fprintln(writer, fmt.Sprintf("ID-PACKAGE\t|%s", gPkgID))
		fmt.Fprintln(writer, fmt.Sprintf("ID-TAG\t|%s", gTagID))
		fmt.Fprintln(writer, fmt.Sprintf("PRICE-CURRENT\t|%s", gPriceCurrent))
		fmt.Fprintln(writer, fmt.Sprintf("PRICE-DISCOUNT\t|%s", gPriceDiscount))
		fmt.Fprintln(writer)
	})
	writer.Flush()
}
