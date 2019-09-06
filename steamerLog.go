package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

type SteamerLog struct {
	PagesFrom     int               `json:"pages_from"`
	PagesTo       int               `json:"pages_to"`
	PagesOK       *SteamerLogPageOK `json:"pages_ok"`
	TerminateZero bool              `json:"terminate_zero"`
	TimeDuration  time.Duration     `json:"time_duration"`
	TimeEnd       time.Time         `json:"time_end"`
	TimeStart     time.Time         `json:"time_start"`
}

func writeSteamerLog(fullpath string, s *SteamerLog) error {
	err := os.MkdirAll(fullpath, os.ModePerm)
	if err != nil {
		return err
	}
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	filename := fmt.Sprintf("%d-%d-log.json", s.PagesFrom, s.PagesTo)
	fullname := filepath.Join(fullpath, filename)
	err = ioutil.WriteFile(fullname, b, os.ModePerm)
	return err
}

func writeSteamerLogDefault(s *SteamerLog) error {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fullpath := filepath.Join(user.HomeDir, "Desktop", "steambot")
	err = writeSteamerLog(fullpath, s)
	return err
}
