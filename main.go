package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type game struct {
	AppID              string            `json:"appid"`
	BundleID           string            `json:"bundleid"`
	Categories         []gameCategory    `json:"categories"`
	CrtrID             string            `json:"crtrid"`
	DescriptionID      string            `json:"descriptionid"`
	Description        string            `json:"description"`
	DescriptionVerbose string            `json:"descriptionverbose"`
	Developer          []gameDeveloper   `json:"developer"`
	Genre              []gameGenre       `json:"genre"`
	Languages          []gameLanguage    `json:"languages"`
	Meta               []gameMeta        `json:"meta"`
	Name               string            `json:"name"`
	PackageID          string            `json:"packageid"`
	Publisher          []gamePublisher   `json:"publisher"`
	ReleaseDate        string            `json:"releasedate"`
	Requirements       []gameRequirement `json:"requirements"`
	TagID              string            `json:"tagid"`
	Tags               []string          `json:"tags"`
	Title              string            `json:"title"`
	URL                string            `json:"url"`
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

type gamePublisher struct {
	Name string
	URL  string
}

type gameRequirement struct {
	DirectX   string `json:"directx"`
	Graphics  string `json:"graphics"`
	Memory    string `json:"memory"`
	Name      string `json:"name"`
	Network   string `json:"network"`
	OS        string `json:"os"`
	Processor string `json:"processor"`
	SoundCard string `json:"soundcard"`
	Storage   string `json:"storage"`
}

const steamSearchURL string = "https://store.steampowered.com/search/"

var filterMap map[string]map[string]string

var optionMap map[string]string

var gameMap map[string]game

var hrefGroup []string

var client *http.Client

var scanner *bufio.Scanner

var wg sync.WaitGroup

func scrapeGameCategory(d *goquery.Document) []gameCategory {
	a := d.Find("div.game_area_details_specs a.name")
	gameCategories := make([]gameCategory, a.Length())
	a.Each(func(i int, s *goquery.Selection) {
		gameCategories[i] = gameCategory{
			Name: strings.TrimSpace(s.Text()),
			URL:  strings.TrimSpace(s.AttrOr("href", "NIL"))}
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

func scrapeGameDescriptionVerbose(d *goquery.Document) string {
	descriptionVerbose := strings.TrimSpace(d.Find("#game_area_description").First().Text())
	return descriptionVerbose
}

func scrapeGameDevelopers(d *goquery.Document) []gameDeveloper {
	a := d.Find("#developers_list a")
	gameDevelopers := make([]gameDeveloper, a.Length())
	a.Each(func(i int, s *goquery.Selection) {
		gameDevelopers[i] = gameDeveloper{
			Name: strings.TrimSpace(s.Text()),
			URL:  strings.TrimSpace(s.AttrOr("href", "NIL"))}
	})
	return gameDevelopers
}

func scrapeGameGenre(d *goquery.Document) []gameGenre {
	a := d.Find("div.game_details div.details_block:first-child > a")
	gameGenres := make([]gameGenre, a.Length())
	a.Each(func(i int, s *goquery.Selection) {
		gameGenres[i] = gameGenre{
			Name: strings.TrimSpace(s.Text()),
			URL:  strings.TrimSpace(s.AttrOr("href", "NIL"))}
	})
	return gameGenres
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

func scrapeGamePublisher(d *goquery.Document) []gamePublisher {
	a := d.Find("div.dev_row > b:first-child + a")
	gamePublishers := make([]gamePublisher, a.Length())
	a.Each(func(i int, s *goquery.Selection) {
		gamePublishers[i] = gamePublisher{
			Name: strings.TrimSpace(s.Text()),
			URL:  s.AttrOr("href", "NIL")}
	})
	return gamePublishers
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

func scrapeGameRequirements(d *goquery.Document) []gameRequirement {
	gameRequirements := []gameRequirement{}
	d.Find("div.game_area_sys_req[data-os]").Each(func(_ int, s *goquery.Selection) {
		reg := regexp.MustCompile(`[^a-zA-Z]+`)
		gameRequirement := gameRequirement{
			Name: strings.TrimSpace(s.AttrOr("data-os", "NIL"))}
		s.Find("ul.bb_ul").First().Each(func(i int, s *goquery.Selection) {
			m := map[string]string{}
			s.Find("li").Each(func(j int, s *goquery.Selection) {
				key := s.Find("strong").First().Text()
				key = reg.ReplaceAllString(key, "")
				key = strings.ToLower(key)
				m[key] = strings.TrimSpace(s.Text())
			})
			b, err := json.Marshal(m)
			if err != nil {
				panic(err)
			}
			if err := json.Unmarshal(b, &gameRequirement); err != nil {
				panic(err)
			}
			gameRequirements = append(gameRequirements, gameRequirement)
		})
	})
	return gameRequirements
}

func scrapeGamePage(d *goquery.Document) game {
	ID := d.Url.String()
	game, ok := gameMap[ID]
	if ok != true {
		panic(fmt.Sprintf("game not found! %s", ID))
	}
	game.Categories = scrapeGameCategory(d)
	game.Description = scrapeGameDescription(d)
	game.DescriptionVerbose = scrapeGameDescriptionVerbose(d)
	game.Developer = scrapeGameDevelopers(d)
	game.Genre = scrapeGameGenre(d)
	game.Languages = scrapeGameLanguages(d)
	game.Meta = scrapeGameMeta(d)
	game.Publisher = scrapeGamePublisher(d)
	game.ReleaseDate = scrapeGameDate(d)
	game.Requirements = scrapeGameRequirements(d)
	game.Title = scrapeGameTitle(d)
	game.Tags = scrapeGameTags(d)
	gameMap[ID] = game
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

func scrapeStoreCategories(s *goquery.Selection) {
	tag, ok := s.Attr("data-param")
	if ok != true {
		return
	}
	if _, ok := filterMap[tag]; !ok {
		filterMap[tag] = map[string]string{}
	}
	value, ok := s.Attr("data-value")
	if ok != true {
		return
	}
	loc, ok := s.Attr("data-loc")
	if ok != true {
		return
	}
	filter, ok := filterMap[tag]
	if ok != true {
		panic(fmt.Sprintf("tag: %s", tag))
	}
	filter[loc] = value
	filterMap[tag] = filter
	optionMap[loc] = tag
}

func netrunnerGamePages(c chan string) {
	defer wg.Done()
	req, err := http.NewRequest(http.MethodGet, <-c, nil)
	if err != nil {
		return
	}
	req.Header.Set("Cookie", "birthtime=-949485599; lastagecheckage=1-0-1900; wants_mature_content=1")
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

func netrunnerStoreCategories(URL string) {
	req, err := http.NewRequest(http.MethodGet, URL, nil)
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
	doc.Find("#additional_search_options div.tab_filter_control").Each(func(i int, s *goquery.Selection) {
		scrapeStoreCategories(s)
	})
}

func fGamePrintln(w *tabwriter.Writer, game game) {
	s := reflect.ValueOf(&game).Elem()
	typeOfT := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fmt.Fprintln(w, fmt.Sprintf("%s\t|%v", typeOfT.Field(i).Name, f.Interface()))
	}
	//"%d: %s %s = %v\n", i, typeOfT.Field(i).Name, f.Type(), f.Interface()
}

func main() {
	filterMap = map[string]map[string]string{}
	optionMap = map[string]string{}
	gameMap = map[string]game{}
	scanner = bufio.NewScanner(os.Stdin)
	if ok := scanner.Scan(); !ok {
		return
	}
	n, err := strconv.Atoi(scanner.Text())
	if err != nil {
		return
	}
	fmt.Println(fmt.Sprintf("Steamer.exe\t>\tcollecting %d pages", n))
	client = (&http.Client{Timeout: (time.Second * 1)})
	netrunnerStoreCategories(steamSearchURL)
	fmt.Println(filterMap)
	c := make(chan string, n)
	hrefGroup = []string{}
	for i := 1; i < n+1; i++ {
		wg.Add(1)
		c <- fmt.Sprintf("%s?page=%d", steamSearchURL, i)
		netrunnerStorePages(c)
	}
	wg.Wait()
	close(c)
	fmt.Println(fmt.Sprintf("Steamer.exe\t>\tfound %d games", len(hrefGroup)))
	c = make(chan string, len(hrefGroup))
	for _, href := range hrefGroup {
		wg.Add(1)
		c <- href
		netrunnerGamePages(c)
	}
	wg.Wait()
	close(c)
	fmt.Println(fmt.Sprintf("Steamer.exe\t>\tbuilt %d games", len(gameMap)))
	w := new(tabwriter.Writer).Init(os.Stdout, 0, 8, 0, '\t', 0)
	for _, game := range gameMap {
		fGamePrintln(w, game)
		fmt.Fprintln(w, "")
	}
	w.Flush()
}
