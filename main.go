package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	client := &http.Client{Timeout: time.Second * 10}
	n := 1
	for i := 1; i <= n; i++ {
		URL := fmt.Sprintf("store.steampowered.com/search/?page=%d", i)
		onGetSteamGameAbbreviation(client, URL,
			func(s *Snapshot) {
				writeSnapshotDefault(s)
			},
			func(s *SteamGameAbbreviation) {
				onGetSteamGamePage(client, s.URL,
					func(s *Snapshot) {
						writeSnapshotDefault(s)
					},
					func(s *SteamGamePage) {
						fmt.Println(s.Name)
					},
					func(e error) {
					})
			},
			func(e error) {
			})
	}
}
