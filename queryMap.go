package main

type queryMap map[string]string

func (queryMap queryMap) Add(key, value string) bool {
	queryMap[key] = value
	return queryMap.Has(key)
}

func (queryMap queryMap) Get(key string) (string, bool) {
	value, ok := queryMap[key]
	return value, ok
}

func (queryMap queryMap) Has(key string) bool {
	_, ok := queryMap.Get(key)
	return ok
}
