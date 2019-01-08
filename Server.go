package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

type message struct {
	Channel string `json:"channel"`
	ID      string `json:"id"`
	Created int    `json:"created"`
	Author  string `json:"author"`
	Text    string `json:"text"`
}

type webServer struct {
	db           *sql.DB
	messageCount int
}

func (serv *webServer) getMessageCount() {
	rows, _ := serv.db.Query("SELECT COUNT(*) FROM messages")
	var count int
	rows.Scan(&count)
	serv.messageCount = count
}

func (serv *webServer) newPairing(w http.ResponseWriter, r *http.Request) {
	message1 := serv.getRandomMessage()

	var message2 *message
	for {
		message2 = serv.getRandomMessage()
		if message1.ID != message2.ID {
			break
		}
	}

	pairing, _ := json.Marshal([]message{*message1, *message2})
	sendResponse(w, r, string(pairing))
}

func (serv *webServer) getRandomMessage() *message {
	var channel string
	var id string
	var created int
	var author string
	var text string

	command := `SELECT channel, id, created, author, text FROM messages OFFSET floor(random()*$1) LIMIT 1;`
	row := serv.db.QueryRow(command, serv.messageCount)

	switch err := row.Scan(&channel, &id, &created, &author, &text); err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
	case nil:
		fmt.Println(channel, id, created, author, text)
	default:
		panic(err)
	}

	return &message{
		Channel: channel,
		ID:      id,
		Created: created,
		Author:  author,
		Text:    text,
	}
}

func (serv *webServer) submitWinner(w http.ResponseWriter, r *http.Request) {
	winners, ok := r.URL.Query()["winner"]

	if !ok || len(winners[0]) < 1 {
		log.Println("Url Param 'winner' is missing")
		return
	}

	losers, ok := r.URL.Query()["loser"]

	if !ok || len(losers[0]) < 1 {
		log.Println("Url Param 'losers' is missing")
		return
	}

	winnerID := winners[0]
	loserID := losers[0]

	serv.recordVictory(winnerID, loserID)
}

func (serv *webServer) recordVictory(winnerID string, loserID string) {
	winnerRating := serv.getCurrentRating(winnerID)
	loserRating := serv.getCurrentRating(loserID)

	newWinnerRating, newLoserRating := serv.calculateNewRatings(winnerRating, loserRating)

	serv.updateRating(winnerID, newWinnerRating)
	serv.updateRating(loserID, newLoserRating)
}

func (serv *webServer) getCurrentRating(id string) float64 {
	command := `SELECT rating FROM messages WHERE id = $1;`
	row := serv.db.QueryRow(command, id)

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

func (serv *webServer) calculateNewRatings(winnerRating float64, loserRating float64) (float64, float64) {
	ratingDifference := loserRating - winnerRating

	expectedScoreForWinner := serv.calculateExpectedScore(ratingDifference)
	expectedScoreForLoser := serv.calculateExpectedScore(-ratingDifference)

	newWinnerRating := serv.calculateNewElo(winnerRating, expectedScoreForWinner, 1)
	newLoserRating := serv.calculateNewElo(loserRating, expectedScoreForLoser, 0)

	return newWinnerRating, newLoserRating
}

func (serv *webServer) calculateExpectedScore(ratingDifference float64) float64 {
	return 1 / (1 + math.Pow(10, (ratingDifference/400)))
}

func (serv *webServer) calculateNewElo(baseRating float64, expectedScore float64, actualScore float64) float64 {
	return baseRating + 8*(actualScore-expectedScore)
}

func (serv *webServer) updateRating(id string, newRating float64) {
	command := `UPDATE messages SET rating = $2 WHERE id = $1;`
	_, err := serv.db.Exec(command, id, newRating)
	if err != nil {
		panic(err)
	}
}

func getPort() (string, error) {
	port := os.Getenv("PORT")
	if port == "" {
		return "", fmt.Errorf("$PORT not set")
	}
	return ":" + port, nil
}

func main() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	port, err := getPort()
	if err != nil {
		log.Fatal(err)
	}

	serv := &webServer{
		db: db,
	}

	serv.getMessageCount()

	http.HandleFunc("/new_pairing", serv.newPairing)
	http.HandleFunc("/submit_winner", serv.submitWinner)

	err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	// dbInit(db)
}

func sendResponse(w http.ResponseWriter, r *http.Request, response string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

func dbInit(db *sql.DB) {
	messageBodies := []byte(`[{"channel": "circle-hell", "id": "1510020154.000024", "created": 1510020825, "user": "connor", "author": "Schmaron", "text": "<https://www.youtube.com/watch?v=O2BgJUHYWxA&amp;feature=youtu.be>"}, {"channel": "circle-hell", "id": "1511818419.000080", "created": 1511818441, "user": "connor", "author": "ashna", "text": "Heads up- HCC hwk is due Wednesday, not tomorrow "}, {"channel": "circle-hell", "id": "1513481492.000061", "created": 1513481537, "user": "connor", "author": "connor", "text": "i need it to be more millenial"}, {"channel": "circle-hell", "id": "1513710972.000154", "created": 1513710985, "user": "connor", "author": "Schmaron", "text": "What the hell is going on?"}]`)
	messages := make([]message, 0)
	json.Unmarshal(messageBodies, &messages)

	fmt.Println(messages)

	for _, message := range messages {
		insertMessage(db, message)
	}
}

func insertMessage(db *sql.DB, m message) {
	sqlStatement := `
	INSERT INTO messages (id, channel, created, author, text, rating)
	VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.Exec(sqlStatement, m.ID, m.Channel, m.Created, m.Author, m.Text, 2500.0)
	if err != nil {
		panic(err)
	}
}
