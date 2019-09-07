package main

type SteamSearchKeyValue struct {
	Key   string
	Value string
}

func NewSteamSearchKeyValue(dataParam, dataValue string) *SteamSearchKeyValue {
	return &SteamSearchKeyValue{
		Key:   dataParam,
		Value: dataValue}
}
