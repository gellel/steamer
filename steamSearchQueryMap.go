package main

type SteamSearchQueryMap map[string]string

func (steamSearchQueryMap *SteamSearchQueryMap) Add(dataLoc, dataValue string) bool {
	ok := steamSearchQueryMap.Has(dataLoc)
	if ok != true {
		(*steamSearchQueryMap)[dataLoc] = dataValue
	}
	return (ok == false)
}

func (steamSearchQueryMap *SteamSearchQueryMap) Fetch(dataParam string) string {
	dataValue, _ := steamSearchQueryMap.Get(dataParam)
	return dataValue
}

func (steamSearchQueryMap *SteamSearchQueryMap) Get(dataParam string) (string, bool) {
	dataValue, ok := (*steamSearchQueryMap)[dataParam]
	return dataValue, ok
}

func (steamSearchQueryMap *SteamSearchQueryMap) Has(dataParam string) bool {
	_, ok := steamSearchQueryMap.Get(dataParam)
	return ok
}
