package main

import (
	"database/sql"
	"net/http"

	"sync"

	_ "github.com/mattn/go-sqlite3"
)

const schema = `
CREATE TABLE IF NOT EXISTS comment (
	id INTEGER PRIMARY KEY,
	ip TEXT NOT NULL,
	name TEXT NOT NULL,
	comment TEXT NOT NULL,
	timestamp INTEGER NOT NULL
)
`

// Comment holds the data associated with a comment
type Comment struct {
	ip      string
	name    string
	comment string
}

// Cache is the bundled caching and and concurrancy primitive
type Cache struct {
	data string
	sync.RWMutex
}

var cache = &Cache{data: ""}
var newcomments = make(chan Comment)

func init() {
	// Init the DB connection and if need be create the table
	db, _ := sql.Open("sqlite3", "./app.db")
	db.Exec(schema)

}

func (c *Cache) updateCache() {
	c.Lock()
	defer c.Unlock()

	// TODO: query database and build HTML or JSON blob
	// c.data = blob
}

func main() {
	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		// return template rendered with cached data
	})

	http.HandleFunc("/ws", func(rw http.ResponseWriter, r *http.Request) {
		// initiate websocket connection and send results to client as the come through the channel
		// e.g., send( <-newcomments)

	})

	http.HandleFunc("/post", func(rw http.ResponseWriter, r *http.Request) {
		// TODO: Add to database

		go cache.updateCache()

		// TODO: send the results to the ws goroutine. E.g.,
		newcomments <- Comment{}

	})

}
