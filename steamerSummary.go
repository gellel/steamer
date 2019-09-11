package main

type SteamerSummary struct {
	Developers map[string][]string `json:"developers"`
	Games      int                 `json:"games"`
	Genres     map[string]int      `json:"genres"`
	Publishers map[string][]string `json:"publishers"`
	Sentiments map[string]int      `json:"sentiments"`
}
