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

type game struct {
	AppID         string
	BundleID      string
	Categories    []gameCategory
	CrtrID        string
	DescriptionID string
	Description   string
	Developer     []gameDeveloper
	Genre         []gameGenre
	Languages     []gameLanguage
	Meta          []gameMeta
	Name          string
	PackageID     string
	Publisher     string
	ReleaseDate   string
	TagID         string
	Tags          []string
	Title         string
	URL           string
}

type gameCategory struct {
	Name string
	URL  string
}

type gameDeveloper struct {
	Name string
	URL  string
}

type gameGenre struct {
	Name string
	URL  string
}

type gameLanguage struct {
	Audio     bool
	Interface bool
	Name      string
	Subtitles bool
}

type gameMeta struct {
	Content  string
	Name     string
	Property string
}

const steamSearchURL string = "https://store.steampowered.com/search/"

var gameMap map[string]game

var hrefGroup []string

var client *http.Client

var scanner *bufio.Scanner

var wg sync.WaitGroup

func scrapeGameCategory(d *goquery.Document) []gameCategory {
	a := d.Find("div.game_area_details_specs a.name")
	gameCategories := make([]gameCategory, a.Length())
	a.Each(func(i int, s *goquery.Selection) {
		gameCategories[i] = gameCategory{URL: strings.TrimSpace(s.Text())}
	})
	return gameCategories
}

func scrapeGameDate(d *goquery.Document) string {
	date := strings.TrimSpace(d.Find("div.release_date div.date").First().Text())
	return date
}

func scrapeGameDescription(d *goquery.Document) string {
	description := strings.TrimSpace(d.Find("div.game_description_snippet").First().Text())
	return description
}

func scrapeGameDevelopers(d *goquery.Document) []gameDeveloper {
	a := d.Find("#developers_list a")
	gameDevelopers := make([]gameDeveloper, a.Length())
	a.Each(func(i int, s *goquery.Selection) {
		gameDevelopers[i] = gameDeveloper{Name: strings.TrimSpace(s.Text())}
	})
	return gameDevelopers
}

func scrapeGameLanguages(d *goquery.Document) []gameLanguage {
	tr := d.Find("table.game_language_options tr[class='']")
	gameLanguages := make([]gameLanguage, tr.Length())
	tr.Each(func(i int, s *goquery.Selection) {
		var (
			lang      = strings.TrimSpace(s.Find("td:nth-child(1)").Text())
			inter     = strings.TrimSpace(s.Find("td:nth-child(2)").Text())
			audio     = strings.TrimSpace(s.Find("td:nth-child(3)").Text())
			subtitles = strings.TrimSpace(s.Find("td:nth-child(4)").Text())
		)
		gameLanguage := gameLanguage{
			Audio:     (len(audio) != 0),
			Interface: (len(inter) != 0),
			Name:      lang,
			Subtitles: (len(subtitles) != 0)}
		gameLanguages[i] = gameLanguage
	})
	return gameLanguages
}

func scrapeGameMeta(d *goquery.Document) []gameMeta {
	m := d.Find("meta")
	gameMetaTags := make([]gameMeta, m.Length())
	m.Each(func(i int, s *goquery.Selection) {
		var (
			content  = s.AttrOr("content", "NIL")
			name     = s.AttrOr("name", "NIL")
			property = s.AttrOr("property", "NIL")
		)
		gameMeta := gameMeta{
			Content:  content,
			Name:     name,
			Property: property}
		gameMetaTags[i] = gameMeta
	})
	return gameMetaTags
}

func scrapeGamePublisher(d *goquery.Document) string {
	publisher := strings.TrimSpace(d.Find("div.dev_row > b:first-child + a").First().Text())
	return publisher
}

func scrapeGameTags(d *goquery.Document) []string {
	a := d.Find("a.app_tag")
	gameTags := make([]string, a.Length())
	a.Each(func(i int, s *goquery.Selection) {
		gameTags[i] = strings.TrimSpace(s.Text())
	})
	return gameTags
}

func scrapeGameTitle(d *goquery.Document) string {
	title := strings.TrimSpace(d.Find("div.apphub_AppName").First().Text())
	return title
}

func scrapeGamePage(d *goquery.Document) game {
	game := gameMap[d.Url.String()]
	game.Categories = scrapeGameCategory(d)
	game.Description = scrapeGameDescription(d)
	game.Developer = scrapeGameDevelopers(d)
	game.Languages = scrapeGameLanguages(d)
	game.Meta = scrapeGameMeta(d)
	game.Publisher = scrapeGamePublisher(d)
	game.ReleaseDate = scrapeGameDate(d)
	game.Title = scrapeGameTitle(d)
	game.Tags = scrapeGameTags(d)
	return game
}

func scrapePageItemHrefAttribute(s *goquery.Selection) string {
	href, exists := s.Attr("href")
	if exists == true {
		hrefGroup = append(hrefGroup, href)
	}
	return href
}

func scrapePageItemAppIDAttribute(s *goquery.Selection) string {
	ID := strings.TrimSpace(s.AttrOr("data-ds-appid", "NIL"))
	return ID
}

func scrapePageItemBundleIDAttribute(s *goquery.Selection) string {
	ID := strings.TrimSpace(s.AttrOr("data-ds-bundleid", "NIL"))
	return ID
}

func scrapePageItemCrtrIDAttribute(s *goquery.Selection) string {
	ID := strings.TrimSpace(s.AttrOr("data-ds-crtrids", "NIL"))
	return ID
}

func scrapePageItemDescIDAttribute(s *goquery.Selection) string {
	ID := strings.TrimSpace(s.AttrOr("data-ds-descids", "NIL"))
	return ID
}

func scrapePageItemPackageIDAttribute(s *goquery.Selection) string {
	ID := strings.TrimSpace(s.AttrOr("data-ds-packageid", "NIL"))
	return ID
}

func scrapePageItemTagIDAttribute(s *goquery.Selection) string {
	ID := strings.TrimSpace(s.AttrOr("data-ds-tagids", "NIL"))
	return ID
}

func scrapePageItemTitle(s *goquery.Selection) string {
	title := strings.TrimSpace(s.Find("div.search_name span.title").Text())
	return title
}

func scrapePageItem(s *goquery.Selection) game {
	game := game{
		AppID:         scrapePageItemAppIDAttribute(s),
		BundleID:      scrapePageItemBundleIDAttribute(s),
		CrtrID:        scrapePageItemCrtrIDAttribute(s),
		DescriptionID: scrapePageItemDescIDAttribute(s),
		Name:          scrapePageItemTitle(s),
		PackageID:     scrapePageItemPackageIDAttribute(s),
		TagID:         scrapePageItemTagIDAttribute(s),
		URL:           scrapePageItemHrefAttribute(s)}
	gameMap[game.URL] = game
	return game
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
		scrapePageItem(s)
	})
}

func main() {
	gameMap = make(map[string]game)
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
	for _, game := range gameMap {

		fmt.Println(game)
	}
}
