package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"slack-off-Backend/db"
	"slack-off-Backend/endpoints"

	_ "github.com/lib/pq"
)

func getPort() (string, error) {
	port := os.Getenv("PORT")
	if port == "" {
		return "", fmt.Errorf("$PORT not set")
	}
	return ":" + port, nil
}

func main() {
	port, err := getPort()
	if err != nil {
		log.Fatal(err)
	}

	serv := endpoints.NewEndpoints(db.NewDB())

	http.HandleFunc("/new_pairing", serv.NewPairing)
	http.HandleFunc("/submit_winner", serv.SubmitWinner)

	err = http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
