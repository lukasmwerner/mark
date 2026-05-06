package store

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/fsnotify/fsnotify"
	"github.com/mattn/go-sqlite3"
)

const CRSQLITE_VERSION = "v0.16.3"

type Flag string

const Embedding = Flag("embedding")

type Options struct {
	Flags []Flag
}

type requirement struct {
	name       string
	definition string
	flags      []Flag
}

var Tables = []requirement{
	{
		name: "Bookmarks",
		definition: `CREATE TABLE IF NOT EXISTS Bookmarks (
    id INTEGER PRIMARY KEY NOT NULL,
    url TEXT,
    title TEXT,
    description TEXT,
    tags TEXT
);`,
	},
	{
		name: "Bookmarks_fts",
		definition: `CREATE VIRTUAL TABLE IF NOT EXISTS Bookmarks_fts USING fts5(
    url,
    title,
    description,
    tags
);`,
	},
	{
		name: "Bookmark_Sync_1",
		definition: `CREATE TRIGGER IF NOT EXISTS Bookmarks_insert AFTER INSERT ON Bookmarks
BEGIN
    INSERT INTO Bookmarks_fts (rowid, url, title, description, tags)
    VALUES (new.id, new.url, new.title, new.description, new.tags);
END;`,
	},
	{
		name: "Bookmark_Sync_2",
		definition: `CREATE TRIGGER IF NOT EXISTS Bookmarks_delete AFTER DELETE ON Bookmarks
BEGIN
    DELETE FROM Bookmarks_fts WHERE rowid = old.id;
END;`,
	},
	{
		name: "Bookmark_Sync_3",
		definition: `CREATE TRIGGER IF NOT EXISTS Bookmarks_update AFTER UPDATE ON Bookmarks
BEGIN
    DELETE FROM Bookmarks_fts WHERE rowid = old.id;
    INSERT INTO Bookmarks_fts (rowid, url, title, description, tags)
    VALUES (new.id, new.url, new.title, new.description, new.tags);
END;`,
	},
	{
		name: "Server_Keys",
		definition: `CREATE TABLE IF NOT EXISTS Server_Keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT NOT NULL UNIQUE
);`,
	},
	{
		name:       "Preformance_Tune_WAL",
		definition: `PRAGMA journal_mode=WAL;`,
	},
	{
		name: "Registered_Jobs_Table",
		definition: `CREATE TABLE IF NOT EXISTS jobs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	matcher TEXT NOT NULL,
	action TEXT NOT NULL
);`,
	},
	{
		name: "Pending_jobs_Table",
		definition: `CREATE TABLE IF NOT EXISTS pending_jobs (
	bookmark_id INTEGER,
	action TEXT NOT NULL
);`,
	},
	{
		name: "Bookmark_Embedding",
		definition: `CREATE VIRTUAL TABLE IF NOT EXISTS bookmark_embeddings using vec0(
	document_id int,
	embedding float[768]
		);`,
		flags: []Flag{Embedding},
	},
	{
		name: "Bookmark_embeddings_Sync_1",
		definition: `CREATE TRIGGER IF NOT EXISTS Bookmarks_embeddings_insert AFTER INSERT ON Bookmarks
BEGIN
    INSERT INTO bookmark_embeddings (document_id, embedding)
	VALUES (new.id, embed('embeddinggemma', concat_ws(' ', 'title: ', new.title, ' | text: ', new.description, new.tags, justPath(new.url))));
END;`,
		flags: []Flag{Embedding},
	},
	{
		name: "Bookmark_embeddings_Sync_2",
		definition: `CREATE TRIGGER IF NOT EXISTS Bookmarks_embeddings_update AFTER UPDATE ON Bookmarks
BEGIN
    DELETE FROM bookmark_embeddings WHERE document_id = old.id;
    INSERT INTO bookmark_embeddings (document_id, embedding)
	VALUES (new.id, embed('embeddinggemma', concat_ws(' ', 'title: ', new.title, ' | text: ', new.description, new.tags, justPath(new.url))));
END;`,
		flags: []Flag{Embedding},
	},
	{
		name: "Bookmark_embeddings_Sync_3",
		definition: `CREATE TRIGGER IF NOT EXISTS Bookmarks_embeddings_delete AFTER DELETE ON Bookmarks
BEGIN
    DELETE FROM Bookmarks_fts WHERE rowid = old.id;
END;`,
		flags: []Flag{Embedding},
	},
}

func Open(options Options) (*DB, error) {
	markStoreLocation := os.Getenv("MARK_STORE_LOCATION")
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.Join(errors.New("unable to get homedir"), err)
	}
	if markStoreLocation == "" {
		markStoreLocation = path.Join(homedir, ".config", "mark")
	}

	if err := EnsureDirExists(markStoreLocation); err != nil {
		return nil, errors.Join(errors.New("unable to make mark store location in: "+markStoreLocation), err)
	}
	changesPath := path.Join(markStoreLocation, "changes")
	if err := EnsureDirExists(changesPath); err != nil {
		return nil, errors.Join(errors.New("unable to make mark store changes location in: "+markStoreLocation), err)
	}

	var ext string
	switch runtime.GOOS {
	case "windows":
		ext = ".dll"
	case "darwin":
		ext = ".dylib"
	default:
		ext = ".so"
	}
	if !DoesFileExist(path.Join(markStoreLocation, "crsqlite"+ext)) {
		err = downloadCrSqlite(markStoreLocation, "crsqlite"+ext)
		if err != nil {
			return nil, errors.Join(errors.New("unable to download crsqlite"), err)
		}
	}
	sql.Register("cr-sqlite", &sqlite3.SQLiteDriver{
		Extensions: []string{path.Join(markStoreLocation, "crsqlite")},
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			err := conn.RegisterFunc("embed", ollama_embedding, true)
			if err != nil {
				return err
			}
			err = conn.RegisterFunc("justPath", justPath, true)
			if err != nil {
				return err
			}
			return nil
		},
	})
	sqlite_vec.Auto()

	sqlDB, err := sql.Open("cr-sqlite", path.Join(markStoreLocation, "data.db"))
	if err != nil {
		return nil, errors.Join(errors.New("unable to open database"), err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	db := &DB{
		DB:              sqlDB,
		StoreLoc:        markStoreLocation,
		ChangesStoreLoc: changesPath,
		Hostname:        hostname,
	}

	err = EnsureTables(db, Tables...)
	if err != nil {
		log.Println(err.Error())
	}

	_, err = db.Exec("select crsql_as_crr('Bookmarks');")
	if err != nil {
		return nil, errors.Join(errors.New("unable to setup crdts"), err)
	}

	err = syncronizeFromHostsToDB(db, hostname, changesPath)
	if err != nil {
		return nil, errors.Join(errors.New("unable to sync fs -> db"), err)
	}

	return db, nil
}

type DB struct {
	*sql.DB

	StoreLoc        string
	ChangesStoreLoc string
	Hostname        string

	syncLock sync.Mutex
}

func (db *DB) SyncChanges() {
	db.syncLock.Lock()
	syncronizeLocalChangesToDisk(db, path.Join(db.ChangesStoreLoc, db.Hostname))
	db.syncLock.Unlock()
}

func (db *DB) Close() error {
	err := syncronizeLocalChangesToDisk(db, path.Join(db.ChangesStoreLoc, db.Hostname))
	if err != nil {
		return err
	}

	_, err = db.Exec(`select crsql_finalize();`) // Clean up after cr-sqlite
	if err != nil {
		return err
	}

	return db.DB.Close()
}

func (db *DB) FSWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(err)
		return
	}
	defer watcher.Close()

	watcher.Add(db.ChangesStoreLoc)

	for {
		select {
		case err := <-watcher.Errors:
			fmt.Println("had error watching", err.Error())
		case event := <-watcher.Events:
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Rename) {
				syncronizeFromHostsToDB(db, db.Hostname, db.ChangesStoreLoc)
				log.Println("sync-event!")
			}
		}
	}
}

func EnsureTables(db *DB, tables ...requirement) error {
	for _, table := range tables {
		if table.flags != nil {
			// TODO: there are flags... need to check if we have them enabled
		}
		_, err := db.Exec(table.definition)
		if err != nil {
			log.Print(table.name, err.Error())
			return err
		}
	}

	return nil
}

func InsertBookmark(db *DB, bookmark Bookmark) (BookmarkId, error) {
	tags := strings.Join(bookmark.Tags, ", ")
	result, err := db.Exec("INSERT INTO Bookmarks (url, title, description, tags) VALUES (?, ?, ?, ?)",
		bookmark.Url, bookmark.Title, bookmark.Description, tags)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return BookmarkId(id), err
}

func GetBookmark(db *DB, query_url string) (Bookmark, error) {
	var b Bookmark
	var tags string

	u, err := url.Parse(query_url)
	if err != nil {
		return b, err
	}

	rows, err := db.Query(`SELECT url, title, description, tags 
	FROM Bookmarks 
	WHERE url like ? 
	AND url like ?`, fmt.Sprintf("%%%v%%", u.Host), fmt.Sprintf("%%%v%%", u.Path))
	if err != nil {
		return b, err
	}

	i := 0
	for rows.Next() {
		if i > 0 {
			return b, errors.New("multiple bookmarks found")
		}
		err := rows.Scan(&b.Url, &b.Title, &b.Description, &tags)
		if err != nil {
			return b, err
		}
		if len(tags) > 0 {
			b.Tags = strings.Split(tags, ", ")
		} else {
			b.Tags = []string{}
		}
		i++
	}

	if i == 0 {
		return b, sql.ErrNoRows
	}

	db_url, err := url.Parse(b.Url)
	if err != nil {
		return b, errors.New("unable to parse bookmark url")
	}

	if db_url.Path != u.Path {
		return b, sql.ErrNoRows
	}

	return b, nil
}

func SemanticSearchBookmarks(db *DB, query string) ([]Bookmark, error) {
	bookmarks := []Bookmark{}
	rows, err := db.Query(`SELECT b.url, b.title, b.description, b.tags
				FROM bookmark_embeddings e
				JOIN Bookmarks b ON e.document_id = b.id
				WHERE e.embedding MATCH embed('embeddinggemma', concat_ws(' ', 'task: search result | query: ', ?)) and k = 100 and distance <= 1.21
				ORDER BY distance;`, query)
	if err != nil {
		return bookmarks, err
	}
	defer rows.Close()

	for rows.Next() {
		var b Bookmark
		var tags string
		err := rows.Scan(&b.Url, &b.Title, &b.Description, &tags)
		if err != nil {
			return bookmarks, err
		}
		b.Tags = strings.Split(tags, ", ")
		bookmarks = append(bookmarks, b)
	}

	return bookmarks, nil
}

func SearchBookmarks(db *DB, query string) ([]Bookmark, error) {
	bookmarks := []Bookmark{}
	query = strings.Join(strings.Fields(query), "* ") + "*"
	rows, err := db.Query(`SELECT url, title, description, tags 
		FROM Bookmarks_fts 
		WHERE Bookmarks_fts MATCH ? 
		ORDER BY bm25(Bookmarks_fts) DESC;`, query)
	if err != nil {
		return bookmarks, err
	}
	defer rows.Close()

	for rows.Next() {
		var b Bookmark
		var tags string
		err := rows.Scan(&b.Url, &b.Title, &b.Description, &tags)
		if err != nil {
			return bookmarks, err
		}
		b.Tags = strings.Split(tags, ", ")
		bookmarks = append(bookmarks, b)
	}

	return bookmarks, nil
}

func UpdateBookmark(db *DB, original Bookmark, updated Bookmark) error {
	_, err := db.Exec(`UPDATE Bookmarks SET
		url = ?,
		title = ?,
		description = ?,
		tags = ?
	WHERE
		url = ?;`,
		updated.Url,
		updated.Title,
		updated.Description,
		strings.Join(updated.Tags, ", "),
		original.Url,
	)

	return err
}

func DeleteBookmark(db *DB, bookmark Bookmark) error {
	_, err := db.Exec(`DELETE FROM Bookmarks
	WHERE
		url = ? AND
		title = ? AND
		description = ? AND
		tags = ?;`,
		bookmark.Url,
		bookmark.Title,
		bookmark.Description,
		strings.Join(bookmark.Tags, ", "),
	)
	return err
}

func CountBookmarks(db *DB) (count int) {
	rows, err := db.Query(`SELECT count(*) FROM Bookmarks;`)
	if err != nil {
		return count
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&count)
		return count
	}
	return count

}

func AddKey(db *DB, key string) error {
	_, err := db.Exec("INSERT INTO Server_Keys (key) VALUES (?)", key)
	if err != nil {
		return err
	}
	return nil
}

func HasKey(db *DB, key string) (bool, error) {
	var value string
	err := db.QueryRow("SELECT key FROM Server_Keys WHERE key = ?", key).Scan(&value)
	if err != nil || value == "" {
		return false, err
	}
	return true, nil
}

func GetKeys(db *DB) ([]string, error) {
	keys := []string{}
	rows, err := db.Query("SELECT key FROM Server_Keys")
	if err != nil {
		return keys, err
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		err := rows.Scan(&key)
		if err != nil {
			return keys, err
		}
		keys = append(keys, key)
	}

	return keys, nil
}

func DeleteKey(db *DB, key string) error {
	_, err := db.Exec("DELETE FROM Server_Keys WHERE key = ?", key)
	if err != nil {
		return err
	}
	return nil
}
