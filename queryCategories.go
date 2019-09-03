package main

import "fmt"

type queryCategories map[string]queryMap

func (queryCategories queryCategories) Add(tag, key, value string) bool {
	if ok := queryCategories.Has(tag); ok != true {
		queryCategories.Set(tag)
	}
	queryMap, ok := queryCategories.Get(tag)
	if ok != true {
		panic(fmt.Sprintf("cannot find %s", tag))
	}
	queryMap.Add(key, value)
	queryCategories[tag] = queryMap
	return queryCategories.Has(tag)
}

func (queryCategories queryCategories) Get(tag string) (queryMap, bool) {
	queryMap, ok := queryCategories[tag]
	return queryMap, ok
}

func (queryCategories queryCategories) Has(tag string) bool {
	_, ok := queryCategories[tag]
	return ok
}

func (queryCategories queryCategories) Set(tag string) bool {
	_, ok := queryCategories[tag]
	if ok != true {
		queryCategories[tag] = queryMap{}
	}
	return (ok == false)
}
