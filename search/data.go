package search

import "github.com/lukasmwerner/mark/store"

type Result struct {
	store.Bookmark
	Source Source
	Rank   float64
}

type Bookmark interface {
	Get() store.Bookmark
	Score() float64
}
