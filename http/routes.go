package http

import (
	"net/http"

	"github.com/lukasmwerner/mark/store"
)

type Handler func(*store.DB) http.Handler

type Route struct {
	Pattern string
	Handler Handler
}

var routes = []Route{
	{
		Pattern: "GET /api/bookmarks/search",
		Handler: SearchBookmarksHandler,
	},
	{
		Pattern: "POST /api/bookmarks",
		Handler: CreateBookmarkHandler,
	},
	{
		Pattern: "PATCH /api/bookmarks",
		Handler: UpdateBookmarkHandler,
	},
	{
		Pattern: "GET /api/bookmarks",
		Handler: GetBookmarkHandler,
	},
	{
		Pattern: "GET /api/stats",
		Handler: StatsHandler,
	},
}

func RegisterRoutes(db *store.DB, mux *http.ServeMux) {
	for _, route := range routes {
		mux.Handle(route.Pattern, route.Handler(db))
	}
}
