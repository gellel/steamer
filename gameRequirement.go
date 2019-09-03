package main

// GameRequirement expresses the benchmark that the Steam game should meet for performance metrics.
// A GameRequirement can express either a minimum or recommended specification.
type gameRequirement struct {
	DirectX   string `json:"directx"`
	Graphics  string `json:"graphics"`
	Memory    string `json:"memory"`
	Name      string `json:"name"`
	Network   string `json:"network"`
	OS        string `json:"os"`
	Processor string `json:"processor"`
	SoundCard string `json:"soundcard"`
	Storage   string `json:"storage"`
}
