package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/lukasmwerner/mark/store"
)

func AuthRequired(db *store.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		bearer := r.Header.Get("Authorization")
		if bearer == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(bearer, "Bearer ")

		if token == "" {
			http.Error(w, "Invalid Authorization header", http.StatusUnauthorized)
			return
		}

		exists, err := store.HasKey(db, token)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if exists {
			next.ServeHTTP(w, r)
		} else {
			fmt.Println("Unauthorized, token not found: ", token[:10])
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
	})
}
