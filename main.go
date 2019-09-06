package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {
	i := 1
	n := 1000
	wg := &sync.WaitGroup{}
	client := &http.Client{Timeout: time.Second * 10}
	steamerLog := &SteamerLog{
		PagesFrom: i,
		PagesTo:   n,
		PagesOK:   &SteamerLogPageOK{},
		TimeStart: time.Now()}
	fmt.Println("timeStart", "\t", "->", steamerLog.TimeStart)
	for i := 1; i <= n; i++ {
		URL := fmt.Sprintf("store.steampowered.com/search/?page=%d", i)
		wg.Add(1)
		go func(client *http.Client, URL string) {
			defer wg.Done()
			onGetSteamGameAbbreviation(client, URL,
				func(s *Snapshot) {
					writeSnapshotDefault(s)
				},
				func(s *SteamGameAbbreviation) {

					writeSteamGameAbbreviationDefault(s)

					wg.Add(1)
					go func(client *http.Client, URL string) {
						defer wg.Done()
						onGetSteamGamePage(client, URL,
							func(s *Snapshot) {
								writeSnapshotDefault(s)
							},
							func(s *SteamGamePage) {

								writeSteamGamePageDefault(s)

								wg.Add(1)
								go func(client *http.Client, URL string) {
									defer wg.Done()
									onGetSteamChartPage(client, fmt.Sprintf("https://steamcharts.com/app/%d", s.AppID),
										func(s *Snapshot) {
											writeSnapshotDefault(s)
										},
										func(s *SteamChartPage) {

											writeSteamChartPageDefault(s)
										},
										func(e error) {

										})
								}(client, fmt.Sprintf("https://steamcharts.com/app/%d", s.AppID))
							},
							func(e error) {
							})
					}(client, s.URL)
				},
				func(e error) {
				})
		}(client, URL)
	}
	wg.Wait()
	steamerLog.TimeEnd = time.Now()
	fmt.Println("timeEnd", "\t", "->", steamerLog.TimeEnd)
	steamerLog.TimeDuration = steamerLog.TimeEnd.Sub(steamerLog.TimeStart)
	writeSteamerLogDefault(steamerLog)
	fmt.Println("timeDuration", "\t", "->", steamerLog.TimeDuration)
}
