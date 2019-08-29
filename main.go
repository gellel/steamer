package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

var wg sync.WaitGroup

var mutex = sync.RWMutex{}

var searchQueryCatalogue = map[string]map[string]string{}

func main() {
	attrDataParam := "data-param"
	resp, err := http.Get("https://store.steampowered.com/search/")
	if err != nil {
		os.Exit(1)
	}
	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		os.Exit(1)
	}
	s := doc.Find(fmt.Sprintf("[%s]", attrDataParam))
	l := s.Length()
	if ok := l == 0; ok {
		os.Exit(1)
	}
	c := make(chan *goquery.Selection, l)
	s.Each(func(i int, s *goquery.Selection) {
		dataParam, exists := s.Attr(attrDataParam)
		if exists != false {
			wg.Add(1)
			c <- s
			go func(c chan *goquery.Selection, dataParam string) {
				defer wg.Done()
				mutex.Lock()
				if _, ok := searchQueryCatalogue[dataParam]; ok != true {
					searchQueryCatalogue[dataParam] = map[string]string{}
				}
				s := <-c
				queryKey, _ := s.Attr("[data-loc]")
				queryValue, _ := s.Attr("[data-value]")
				searchQueryCatalogue[dataParam][queryKey] = queryValue
				mutex.Unlock()
			}(c, dataParam)
		}
	})
	wg.Wait()
	close(c)
	for dataParam := range searchQueryCatalogue {
		fmt.Println(dataParam)
	}
}
