package main

import "fmt"

// GameDeveloper is a struct that expresses the individual or group that was responsible for creating the Steam game.
type gameDeveloper struct {
	Name  string `json:"name"`  // {Name: "ROCKSTAR-NORTH"}
	Title string `json:"title"` // {Title: "Rockstar North"}
	URL   string `json:"url"`   // {URL: "https://store.steampowered.com/developer/rockstarnorth"}
}

func (gameDeveloper gameDeveloper) String() string {
	return fmt.Sprintf("%s", gameDeveloper.Name)
}
