package store

type Bookmark struct {
	Url         string
	Tags        []string
	Title       string
	Description string
}

func (b Bookmark) FilterValue() string { return b.Url }

type BookmarkId int64
