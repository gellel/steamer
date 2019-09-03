package main

import "fmt"

// GameCategory is a struct that expresses the individual attribute of the Steam game. Unlike a GameGenre, a GameCategory
// describes the unique attributes at a feature level that help distinguish the options one game offers over another.
type gameCategory struct {
	Name  string `json:"name"`  // {Name: "SINGLE-PLAYER"}
	Title string `json:"title"` // {Title: "Single Player"}
	URL   string `json:"url"`   // {URL: "https://store.steampowered.com/search/?category2=2"}
}

func (gameCategory gameCategory) String() string {
	return fmt.Sprintf("%s", gameCategory.Name)
}
