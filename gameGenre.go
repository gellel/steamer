package main

import "fmt"

// GameGenre is a struct that expresses the individual genre of the Steam game. Unlike a GameCategory, a GameGenre
// describes the unique qualities at a gameplay level that help distinguish one type of game from another.
type gameGenre struct {
	Name  string `json:"name"`  // {Name: "FIRST-PERSON-SHOOTER"}
	Title string `json:"title"` // {Name: "First Person Shooter"}
	URL   string `json:"url"`   // {URL: "https://store.steampowered.com/tags/en/Action"}
}

func (gameGenre gameGenre) String() string {
	return fmt.Sprintf("%s", gameGenre.Name)
}
