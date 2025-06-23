package store_test

import (
	"github.com/lukasmwerner/mark/store"
	"testing"
)

func TestSearchBookmarks(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		db      *store.DB
		query   string
		want    []store.Bookmark
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := store.SearchBookmarks(tt.db, tt.query)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("SearchBookmarks() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("SearchBookmarks() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("SearchBookmarks() = %v, want %v", got, tt.want)
			}
		})
	}
}
