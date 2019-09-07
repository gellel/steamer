package main

type SteamerLogPageOK map[int]bool

func (steamerLogPageOK *SteamerLogPageOK) Add(page int, status bool) bool {
	ok := steamerLogPageOK.Has(page)
	if ok != true {
		(*steamerLogPageOK)[page] = status
	}
	return (ok == false)
}

func (steamerLogPageOK *SteamerLogPageOK) Get(page int) (bool, bool) {
	pageStatus, ok := (*steamerLogPageOK)[page]
	return pageStatus, ok
}

func (steamerLogPageOK *SteamerLogPageOK) Has(page int) bool {
	_, ok := steamerLogPageOK.Get(page)
	return ok
}
