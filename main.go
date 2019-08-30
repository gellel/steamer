package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

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

//var writer = new(tabwriter.Writer).Init(os.Stdout, 0, 8, 0, '\t', 0)

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
	/*err = writer.Flush()
	if err != nil {
		os.Exit(1)
	}*/
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
	steamSearchCaption := strings.ToUpper(doc.Find(".search_pagination_left").Text())
	r, err = regexp.Compile(`\r?\n`)
	if err != nil {
		os.Exit(1)
	}
	steamSearchCaption = r.ReplaceAllString(steamSearchCaption, "")
	fmt.Println(strings.Split(steamSearchCaption, "OF"))

	/*
		// find the available steam store game options.
		s = doc.Find(fmt.Sprintf("[%s]", "data-ds-appid"))
		l = s.Length()
		// confirm that there are at least some items to process.
		if ok := l > 0; !ok {
			os.Exit(1)
		}
		fmt.Println("games:", l)*/
}
