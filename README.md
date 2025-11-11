# mark

![mark tui demo](./images/tui-demo.gif)

Mark is a simple bookmark manager that allows you to save and recall bookmarks.
It also allows you to sync those changes across all your devices using a
file sync service. This is sort-of explained the following blog post:
https://lukaswerner.com/post/2024-08-13@Sqlite-Local-First

### Installation

1. Download GitHub repo
`git clone https://github.com/lukasmwerner/mark.git`
2. Install program
`go install --tags fts5 .`
3. Enjoy!

### Sync

By default `mark` stores its data in `~/.config/mark/data.db` where the
important sync data is stored in `~/.config/mark/changes/`. Configure whatever
your sync engine is (Dropbox, Google Drive, Syncthing etc) to synchronize that
changes folder.

### Usage
```
Available Commands:
  add         Add a bookmark
  completion  Generate the autocompletion script for the specified shell
  delete      deletes bookmarks based on given search query
  dump        dumps the database to a sqlite3 file
  edit        Edit a bookmark
  help        Help about any command
  import      Lets you import bookmarks from an existing csv
  keys        List all allowed api keys for the local http server
  open        Opens the matching search link in the user's browser
  search      [EXPERIMENTAL] google search like tui intended for hosting via gotty or ttyd
  server      Local HTTP server for managing bookmarks (used for chrome extension)
  show        Shows the entry of a bookmark
  sql         Lets you run sql queries on your database
```

