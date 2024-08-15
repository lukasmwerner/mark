package store

type Bookmark struct {
	Url         string
    Tags        []string
    Title       string
    Description string
}

type BookmarkId int64
