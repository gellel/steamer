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

var hrefGroup []string

var client *http.Client

var scanner *bufio.Scanner

var wg sync.WaitGroup

func scrapeGameCategory(d *goquery.Document) {
	d.Find("div.game_area_details_specs a.name").Each(func(i int, s *goquery.Selection) {
		fmt.Println("category:", strings.TrimSpace(s.Text()))
	})
}

func scrapeGameDate(d *goquery.Document) {
	date := strings.TrimSpace(d.Find("div.release_date div.date").First().Text())
	fmt.Println("release date:", date)
}

func scrapeGameDescription(d *goquery.Document) {
	description := strings.TrimSpace(d.Find("div.game_description_snippet").First().Text())
	fmt.Println("descripton:", description)
}

func scrapeGameDevelopers(d *goquery.Document) {
	d.Find("#developers_list a").Each(func(i int, s *goquery.Selection) {
		fmt.Println("developer:", strings.TrimSpace(s.Text()))
	})
}

func scrapeGameLanguages(d *goquery.Document) {
	d.Find("table.game_language_options tr[class='']").Each(func(i int, s *goquery.Selection) {
		var (
			lang      = strings.TrimSpace(s.Find("td:nth-child(1)").Text())
			audio     = strings.TrimSpace(s.Find("td:nth-child(2)").Text())
			subtitles = strings.TrimSpace(s.Find("td:nth-child(3)").Text())
		)
		fmt.Println("language:", lang, "audio:", (len(audio) != 0), "subtitles:", (len(subtitles) != 0))
	})
}

func scrapeGameMeta(d *goquery.Document) {
	d.Find("meta").Each(func(i int, s *goquery.Selection) {
		var (
			content  = s.AttrOr("content", "NIL")
			name     = s.AttrOr("name", "NIL")
			property = s.AttrOr("property", "NIL")
		)
		fmt.Println("meta:", "content:", content, "name:", name, "property:", property)
	})
}

func scrapeGameTags(d *goquery.Document) {
	d.Find("a.app_tag").Each(func(i int, s *goquery.Selection) {
		fmt.Println("tag:", strings.TrimSpace(s.Text()))
	})
}

func scrapeGameTitle(d *goquery.Document) {
	title := strings.TrimSpace(d.Find("div.apphub_AppName").First().Text())
	fmt.Println("title:", title)
}

func scrapeGamePage(d *goquery.Document) {
	scrapeGameCategory(d)
	scrapeGameDate(d)
	scrapeGameDescription(d)
	scrapeGameDevelopers(d)
	scrapeGameLanguages(d)
	scrapeGameMeta(d)
	scrapeGameTags(d)
	scrapeGameTitle(d)
	fmt.Println("-")
}

func netrunnerGamePages(c chan string) {
	defer wg.Done()
	req, err := http.NewRequest(http.MethodGet, <-c, nil)
	if err != nil {
		return
	}
	res, err := client.Do(req)
	if res.StatusCode != http.StatusOK {
		return
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return
	}
	scrapeGamePage(doc)
}

func netrunnerStorePages(c chan string) {
	defer wg.Done()
	req, err := http.NewRequest(http.MethodGet, <-c, nil)
	if err != nil {
		return
	}
	res, err := client.Do(req)
	if err != nil {
		return
	}
	if res.StatusCode != http.StatusOK {
		return
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return
	}
	doc.Find("a.search_result_row").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists != true {
			return
		}
		hrefGroup = append(hrefGroup, href)
	})
}

func main() {
	scanner = bufio.NewScanner(os.Stdin)
	if ok := scanner.Scan(); !ok {
		return
	}
	n, err := strconv.Atoi(scanner.Text())
	if err != nil {
		return
	}
	client = (&http.Client{Timeout: (time.Second * 1)})
	c := make(chan string, n)
	hrefGroup = []string{}
	for i := 1; i < n+1; i++ {
		wg.Add(1)
		c <- fmt.Sprintf("%s?page=%d", steamSearchURL, i)
		netrunnerStorePages(c)
	}
	wg.Wait()
	close(c)
	c = make(chan string, len(hrefGroup))
	for _, href := range hrefGroup {
		wg.Add(1)
		c <- href
		netrunnerGamePages(c)
	}
	wg.Wait()
	close(c)
}
