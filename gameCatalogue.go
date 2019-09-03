package main

// GameCatalogue is a map-like struct that contains a collection of Game structs assigned by their Steam store page URL.
type gameCatalogue map[string]game

func (gameCatalogue gameCatalogue) Add(game game) bool {
	gameCatalogue[game.URL] = game
	return gameCatalogue.Has(game.URL)
}

func (gameCatalogue gameCatalogue) Get(key string) (game, bool) {
	game, ok := gameCatalogue[key]
	return game, ok
}

func (gameCatalogue gameCatalogue) Has(key string) bool {
	_, ok := gameCatalogue.Get(key)
	return ok
}
