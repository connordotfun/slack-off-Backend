package endpoints

import (
	"fmt"
	"log"
	"net/http"

	"github.com/connordotfun/slack-off-Backend/db"
	"github.com/connordotfun/slack-off-Backend/elo"
	"github.com/connordotfun/slack-off-Backend/marshaller"
)

// Endpoints houses the endpoints
type Endpoints struct {
	db *db.DB
}

// NewEndpoints is a constructor the Endpoints
func NewEndpoints(db *db.DB) *Endpoints {
	return &Endpoints{
		db: db,
	}
}

// NewPairing sends a new, random pairing as a response
func (serv *Endpoints) NewPairing(w http.ResponseWriter, r *http.Request) {
	pairing := serv.db.NewPairing()
	serv.sendResponse(w, r, marshaller.ToJSON(pairing))
}

// SubmitWinner accepts matchup results
func (serv *Endpoints) SubmitWinner(w http.ResponseWriter, r *http.Request) {
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
	serv.sendResponse(w, r, "ok")
}

func (serv *Endpoints) recordVictory(winnerID string, loserID string) {
	winnerRating := serv.db.GetCurrentRating(winnerID)
	loserRating := serv.db.GetCurrentRating(loserID)

	newWinnerRating, newLoserRating := elo.CalculateNewRatings(winnerRating, loserRating)

	serv.db.UpdateRating(winnerID, newWinnerRating)
	serv.db.UpdateRating(loserID, newLoserRating)
}

func (serv *Endpoints) sendResponse(w http.ResponseWriter, r *http.Request, response string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprintln(w, response)
	if err != nil {
		panic(err)
	}
}
