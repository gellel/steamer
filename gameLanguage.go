package main

import "fmt"

// GameLanguage is a struct that expresses the provided language support for a Steam game. Audio represents whether the game's
// languages provides full audio translation for that language. Interface represents the whether the game's user interface
// has the current language supported. Subtitles is whether the game has subtitle support for foreign audio.
type gameLanguage struct {
	Audio     bool   `json:"audio"`     // {Audio: true}
	Interface bool   `json:"interface"` // {Interface: true}
	Name      string `json:"name"`      // {Name: "ENGLISH"}
	Subtitles bool   `json:"subtitles"` // {Subtitles: true}
}

func (gameLanguage gameLanguage) String() string {
	return fmt.Sprintf("{%s Audio %t Interface %t Subtitles %t}", gameLanguage.Name, gameLanguage.Audio, gameLanguage.Interface, gameLanguage.Subtitles)
}
