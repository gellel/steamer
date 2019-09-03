package main

// GameTag is a struct that expresses the user-defined genre categories for a Steam game.
type gameTag struct {
	Name  string `json:"name"`  // {Name: "CHOICES-MATTER"}
	Title string `json:"title"` // {Name: "Choices Matter"}
	URL   string `json:"url"`   // {URL: "https://store.steampowered.com/tags/en/Choices%20Matter/?snr=1_5_9__409"}
}
