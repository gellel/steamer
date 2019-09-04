package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type netSnapshot struct {
	req          *http.Request
	res          *http.Response
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

func getNetSnapshot(c chan *netSnapshot, r *http.Client, m, URL string) {
	defer wg.Done()
	if ok := (strings.HasPrefix(URL, "http://") || strings.HasPrefix(URL, "https://")); !ok {
		URL = fmt.Sprintf("https://%s", URL)
	}
	req, errReq := http.NewRequest(m, URL, nil)
	timeStart := time.Now()
	if errReq != nil {
		timeStart = time.Time{}
	}
	res, errRes := r.Do(req)
	timeEnd := time.Now()
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
	c <- &netSnapshot{
		req:          req,
		res:          res,
		ErrReq:       errReq,
		ErrRes:       errRes,
		Method:       m,
		RequestOK:    (errReq == nil),
		ResponseOK:   (errRes == nil),
		Status:       status,
		StatusCode:   statusCode,
		TimeDuration: timeDuration,
		TimeEnd:      timeEnd,
		TimeStart:    timeStart,
		URL:          URL}
}
