package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/connordotfun/slack-off-Backend/message"
)

// DB is an abstraction of the database connection
type DB struct {
	sqldb        *sql.DB
	messageCount int
}

// NewDB is a database constructor
func NewDB() *DB {
	sqldb, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	db := &DB{
		sqldb: sqldb,
	}
	db.getMessageCount()

	return db
}

func (db *DB) getMessageCount() {
	rows, _ := db.sqldb.Query("SELECT COUNT(*) FROM messages")
	var count int
	for rows.Next() {
		rows.Scan(&count)
	}
	db.messageCount = count
}

func (db DB) NewPairing() [2]message.Message {
	message1 := db.getRandomMessage()

	var message2 *message.Message
	for {
		message2 = db.getRandomMessage()
		if message1.ID != message2.ID {
			break
		}
	}

	return [2]message.Message{*message1, *message2}
}

func (db *DB) getRandomMessage() *message.Message {
	var channel string
	var id string
	var author string
	var text string
	var file string

	command := `SELECT id, channel, author, text, file FROM messages OFFSET floor(random()*$1) LIMIT 1;`
	row := db.sqldb.QueryRow(command, db.messageCount)

	switch err := row.Scan(&id, &channel, &author, &text, &file); err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
	case nil:
		fmt.Println(channel, id, author, text, file)
	default:
		panic(err)
	}

	return &message.Message{
		Channel: channel,
		Author:  author,
		ID:      id,
		Text:    text,
		File:    file,
	}
}

// GetCurrentRating returns the rating associated with the ID
func (db *DB) GetCurrentRating(id string) float64 {
	command := `SELECT rating FROM messages WHERE id = $1;`
	row := db.sqldb.QueryRow(command, id)

	var rating float64
	switch err := row.Scan(&rating); err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
	case nil:
		fmt.Println(rating)
	default:
		panic(err)
	}
	return rating
}

// UpdateRating updates the rating associated with the ID
func (db *DB) UpdateRating(id string, newRating float64) {
	command := `UPDATE messages SET rating = $2 WHERE id = $1;`
	_, err := db.sqldb.Exec(command, id, newRating)
	if err != nil {
		panic(err)
	}
}
