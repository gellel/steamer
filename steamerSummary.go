package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

type SteamerSummary struct {
	Developers map[string][]string `json:"developers"`
	Games      int                 `json:"games"`
	Genres     map[string]int      `json:"genres"`
	PagesFrom  int                 `json:"pages_from"`
	PagesTo    int                 `json:"pages_to"`
	Publishers map[string][]string `json:"publishers"`
	Sentiments map[string]int      `json:"sentiments"`
}

func writeSteamerSummary(fullpath string, s *SteamerSummary) error {
	err := os.MkdirAll(fullpath, os.ModePerm)
	if err != nil {
		return err
	}
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	filename := fmt.Sprintf("%d-%d-summary.json", s.PagesFrom, s.PagesTo)
	fullname := filepath.Join(fullpath, filename)
	err = ioutil.WriteFile(fullname, b, os.ModePerm)
	return err
}

func writeSteamerSummaryDefault(s *SteamerSummary) error {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fullpath := filepath.Join(user.HomeDir, "Desktop", "steambot")
	err = writeSteamerSummary(fullpath, s)
	return err
}
