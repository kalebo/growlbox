package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tjgq/broadcast"
)

const schema = `
CREATE TABLE IF NOT EXISTS comment (
	id INTEGER PRIMARY KEY,
	ip TEXT,
	name TEXT NOT NULL,
	comment TEXT NOT NULL,
	timestamp INTEGER
)
`

// Comment holds the data associated with a comment
type Comment struct {
	IP      string
	Name    string `json:name`
	Comment string `json:comment`
}

// Cache is the bundled caching and and concurrancy primitive
type Cache struct {
	data []byte
	sync.RWMutex
}

var db *sql.DB
var cache Cache
var newcomments = make(chan Comment)
var broadcaster broadcast.Broadcaster

func init() {
	// Init the DB connection and if need be create the table
	var err error
	db, err = sql.Open("sqlite3", "./app.db")
	if err != nil {
		log.Fatalf("Error on initializing database connection: %s", err.Error())
	}

	db.Ping()

	db.Exec(schema)

}

func (c *Cache) updateCache() {
	// Query DB for all comments
	rows, err := db.Query("SELECT name, comment FROM comment")
	if err != nil {
		log.Fatalf("Could not get DB records: %s", err)
	}

	defer rows.Close()

	// Build the HTML blob
	var comment Comment
	var blob CommentBlob

	blob.Open()
	for rows.Next() {
		rows.Scan(&comment.Name, &comment.Comment)
		blob.Append(comment.Name, comment.Comment)
	}
	blob.Close()
	fmt.Println(blob.String())

	// Lock the cache and update it
	c.Lock()
	c.data = blob.Buffer.Bytes()
	c.Unlock()

}

func handleRoot(rw http.ResponseWriter, r *http.Request) {
	cache.RLock()
	// return template rendered with cached data
	rw.Write(cache.data)
	cache.RUnlock()

}

func handleWebsocket(rw http.ResponseWriter, r *http.Request) {
	// Attach to broadcast system as listener
	l := broadcaster.Listen()
	defer l.Close()

	// Initiate websocket connection
	conn, err := websocket.Upgrade(rw, r, rw.Header(), 1024, 1024)
	if err != nil {
		http.Error(rw, "Could not open websocket connection", http.StatusBadRequest)
	}

	// and send results to client as the come across the broadcast channel
	for v := range l.Ch {
		comment := v.(Comment)
		msg, err := json.Marshal(&comment)
		if err != nil {
			log.Fatal(err)
			return
		}

		conn.WriteMessage(websocket.TextMessage, msg)
	}
	conn.Close()

}

func handlePost(rw http.ResponseWriter, r *http.Request) {
	// Parse JSON
	var comment Comment
	err := json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		log.Fatalf("Cannot decode JSON: %s", err.Error())
	}
	defer r.Body.Close()

	// Insert into database
	stmt, err := db.Prepare("INSERT INTO comment (ip, name, comment) VALUES (?,?,?)")
	if err != nil {
		log.Fatalf("Cannot prepare SQL statement: %s", err)
	}

	// check x forwarded for and r.RemoteAddr
	_, err = stmt.Exec(r.RemoteAddr, comment.Name, comment.Comment)

	if err != nil {
		log.Fatalf("Cannot insert value into DB: %s", err)
	}

	// Update the cached version of all the comments
	go cache.updateCache()

	// Broadcast the new comment to the ws goroutines
	broadcaster.Send(comment)

}

func main() {
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/ws", handleWebsocket)
	http.HandleFunc("/post", handlePost)

	http.ListenAndServe(":4758", nil)
}
