package http

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lukasmwerner/mark/search"
	"github.com/lukasmwerner/mark/store"
)

func SearchBookmarksHandler(db *store.DB) http.Handler {
	return AuthRequired(db, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			http.Error(w, "Missing query parameter", http.StatusBadRequest)
			return
		}
		results, err := search.FullText5(db, query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		bookmarks := search.ToBookmarks(results)
		w.Header().Set("Content-Type", "application/json")
		jsonBytes, err := json.Marshal(bookmarks)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(jsonBytes)
	}))
}

func CreateBookmarkHandler(db *store.DB) http.Handler {
	return AuthRequired(db, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var bookmark store.Bookmark
		if err := json.NewDecoder(r.Body).Decode(&bookmark); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		id, err := store.InsertBookmark(db, bookmark)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write(fmt.Appendf([]byte{}, `{"id": %d}`, id))
		go db.SyncChanges()
	}))
}

func UpdateBookmarkHandler(db *store.DB) http.Handler {
	return AuthRequired(db, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var submittedBookmark store.Bookmark
		if err := json.NewDecoder(r.Body).Decode(&submittedBookmark); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		originalBookmarkUrl := r.URL.Query().Get("url")

		var originalBookmark store.Bookmark
		originalBookmark.Url = originalBookmarkUrl

		if err := store.UpdateBookmark(db, originalBookmark, submittedBookmark); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		go db.SyncChanges()
	}))
}

func GetBookmarkHandler(db *store.DB) http.Handler {
	return AuthRequired(db, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Query().Get("url")

		bookmarks, err := store.GetBookmark(db, url)
		if err == sql.ErrNoRows {
			http.Error(w, "Bookmark not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(bookmarks)
	}))
}

func StatsHandler(db *store.DB) http.Handler {
	return AuthRequired(db, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stats := struct {
			Count int `json:"bookmark_count"`
		}{}

		stats.Count = store.CountBookmarks(db)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)

	}))
}
