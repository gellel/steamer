package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Snapshot struct {
	document     *goquery.Document
	request      *http.Request
	response     *http.Response
	ErrDoc       error         `json:"err_document"`
	ErrRes       error         `json:"err_response"`
	ErrReq       error         `json:"err_request"`
	Method       string        `json:"method"`
	RequestOK    bool          `json:"request_OK"`
	ResponseOK   bool          `json:"response_OK"`
	Status       string        `json:"status"`
	StatusCode   int           `json:"status_code"`
	TimeDuration time.Duration `json:"time_duration"`
	TimeEnd      time.Time     `json:"time_end"`
	TimeStart    time.Time     `json:"time_start"`
	URL          string        `json:"URL"`
}

func NewSnapshot(c *http.Client, HTTPMethod, URL string) *Snapshot {
	ok := (strings.HasPrefix(URL, "http://") || strings.HasPrefix(URL, "https://"))
	if ok != true {
		URL = fmt.Sprintf("https://%s", URL)
	}
	req, errReq := http.NewRequest(HTTPMethod, URL, nil)
	timeStart := time.Now()
	res, errRes := c.Do(req)
	timeEnd := time.Now()
	return newSnapshot(HTTPMethod, URL, req, errReq, res, errRes, timeStart, timeEnd)
}

func newSnapshot(HTTPMethod, URL string, req *http.Request, errReq error, res *http.Response, errRes error, timeStart, timeEnd time.Time) *Snapshot {
	if errReq != nil {
		timeStart = time.Time{}
	}
	if errRes != nil {
		timeEnd = time.Time{}
	}
	timeDuration := timeEnd.Sub(timeStart)
	if (errReq != nil) && (errRes != nil) {
		timeDuration = 0
	}
	var status string
	var statusCode int
	if res != nil {
		status = res.Status
		statusCode = res.StatusCode
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	return &Snapshot{
		document:     doc,
		request:      req,
		response:     res,
		ErrDoc:       err,
		ErrReq:       errReq,
		ErrRes:       errRes,
		Method:       HTTPMethod,
		RequestOK:    (errReq == nil),
		ResponseOK:   (errRes == nil),
		Status:       status,
		StatusCode:   statusCode,
		TimeDuration: timeDuration,
		TimeEnd:      timeEnd,
		TimeStart:    timeStart,
		URL:          URL}
}

func writeSnapshot(fullpath string, s *Snapshot) error {
	err := os.MkdirAll(fullpath, os.ModePerm)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	replacer := strings.NewReplacer("https://", "", "/", "", "\\", "", s.request.URL.Host, "", "=", "-", "?", ".")
	filename := replacer.Replace(fmt.Sprintf("%s.json", s.request.URL.String()))
	if strings.HasPrefix(filename, ".") {
		filename = strings.TrimPrefix(filename, ".")
	}
	fullname := filepath.Join(fullpath, filename)
	err = ioutil.WriteFile(fullname, b, os.ModePerm)
	return err
}

func writeSnapshotDefault(s *Snapshot) error {
	user, err := user.Current()
	if err != nil {
		return err
	}
	fullpath := filepath.Join(user.HomeDir, "Desktop", "steambot", s.request.URL.Host)
	return writeSnapshot(fullpath, s)
}
