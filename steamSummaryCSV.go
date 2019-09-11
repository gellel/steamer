package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"strings"
)

type SteamSummaryCSV struct {
	Heading []string
	Values  []string
}

func NewSteamSummaryCSV(s *SteamGameSummary) SteamSummaryCSV {
	r := reflect.ValueOf(s).Elem()
	t := r.Type()
	var (
		headings []string
		values   []string
	)
	for i := 0; i < r.NumField(); i++ {
		var v string
		headings = append(headings, t.Field(i).Name)
		x := r.Field(i).Interface()
		switch x.(type) {
		case []string:
			v = strings.Join(x.([]string), ",")
		default:
			v = fmt.Sprintf("%v", x)
		}
		values = append(values, v)
	}
	return SteamSummaryCSV{
		Heading: headings,
		Values:  values}
}

func writeSteamSummaryCSV(fullpath string, name string, s *[]SteamSummaryCSV) error {
	if len(*s) == 0 {
		return errors.New("s cannot be empty")
	}
	err := os.MkdirAll(fullpath, os.ModePerm)
	if err != nil {
		return err
	}
	if ok := strings.HasSuffix(name, ".csv"); ok != true {
		name = fmt.Sprintf("%s.csv", name)
	}
	file, err := os.Create(filepath.Join(fullpath, name))
	defer file.Close()
	if err != nil {
		return err
	}
	writer := csv.NewWriter(file)
	defer writer.Flush()
	if err := writer.Write((*s)[0].Heading); err != nil {
		return err
	}
	for _, steamSummaryCSV := range *s {
		if err := writer.Write(steamSummaryCSV.Values); err != nil {
			return err
		}
	}
	return nil
}

func writeSteamSummaryCSVDefault(name string, s *[]SteamSummaryCSV) error {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fullpath := filepath.Join(user.HomeDir, "Desktop", "steambot")
	err = writeSteamSummaryCSV(fullpath, name, s)
	return err
}
