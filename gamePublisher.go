package main

import "fmt"

// GamePublisher is a struct that expresses the individual or group that was responsible for publishing the Steam game.
type gamePublisher struct {
	Name  string `json:"name"`  // {Name: "ROCKSTAR-GAMES"}
	Title string `json:"title"` // {Title: "Rockstar Games"}
	URL   string `json:"url"`   // {URL: "https://store.steampowered.com/publisher/rockstargames"}
}

func (gamePublisher gamePublisher) String() string {
	return fmt.Sprintf("%s", gamePublisher.Name)
}
