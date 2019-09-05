package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var wg sync.WaitGroup

func mainRequestSteamGamePage(c chan *Snapshot, client *http.Client, URL string) {
	wg.Add(1)
	go chanSnapshot(c, client, http.MethodGet, URL)
	steamGameSnapshot := <-c
	if steamGameSnapshot.StatusCode != http.StatusOK {
		return
	}
	if steamGameSnapshot.document == nil {
		return
	}
	s := steamGameSnapshot.document.Find("html")
	SteamSearchGame := NewSteamGamePage(s)
	err := writeSteamGamePageDefault(SteamSearchGame)
	if err != nil {
		panic(err)
	}
}

func mainRequestSteamGameAbbreviation(c chan *Snapshot, client *http.Client, URL string) {
	wg.Add(1)
	go chanSnapshot(c, client, http.MethodGet, URL)
	steamPageSnapshot := <-c
	if steamPageSnapshot.StatusCode != http.StatusOK {
		return
	}
	if steamPageSnapshot.document == nil {
		return
	}
	s := steamPageSnapshot.document.Find("a.search_result_row[href]")
	steamGameSnapshotChan := make(chan *Snapshot)
	s.Each(func(i int, s *goquery.Selection) {
		steamGameAbbreviation := NewSteamGameAbbreviation(s)
		if steamPageSnapshot.URL == "NIL" {
			return
		}
		err := writeSteamGameAbbreviationDefault(steamGameAbbreviation)
		if err != nil {
			panic(err)
		}
		mainRequestSteamGamePage(steamGameSnapshotChan, client, steamGameAbbreviation.URL)
	})
}

func main() {
	client := &http.Client{Timeout: time.Second * 10}
	steamPageSnapshotChan := make(chan *Snapshot)
	for i := 1; i < 3; i++ {
		mainRequestSteamGameAbbreviation(steamPageSnapshotChan, client, fmt.Sprintf("store.steampowered.com/search/?page=%d", i))
	}
	wg.Wait()
}
