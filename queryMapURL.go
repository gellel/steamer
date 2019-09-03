package main

import (
	"fmt"
	"strings"
)

type queryMapURL map[string][]string

func (queryMapURL queryMapURL) Add(tag, key string) bool {
	if ok := queryMapURL.Has(tag); ok != true {
		queryMapURL.Set(tag)
	}
	querySet, _ := queryMapURL.Get(tag)
	querySet = append(querySet, key)
	queryMapURL[tag] = querySet
	return queryMapURL.Has(tag)
}

func (queryMapURL queryMapURL) Get(tag string) ([]string, bool) {
	querySet, ok := queryMapURL[tag]
	return querySet, ok
}

func (queryMapURL queryMapURL) Has(tag string) bool {
	_, ok := queryMapURL[tag]
	return ok
}

func (queryMapURL queryMapURL) Set(tag string) bool {
	_, ok := queryMapURL[tag]
	if ok != true {
		queryMapURL[tag] = []string{}
	}
	return (ok == false)
}

func (queryMapURL queryMapURL) URL() string {
	var properties []string
	for tag, querySet := range queryMapURL {
		properties = append(properties, fmt.Sprintf("%s=%s", tag, strings.Join(querySet, "%2C")))
	}
	return strings.Join(properties, "&")
}
