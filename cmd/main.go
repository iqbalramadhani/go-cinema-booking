package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /movies", listMovies)
	mux.Handle("GET /", http.FileServer(http.Dir("static")))

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

var movie = []movieResponse{
	{
		ID: "inception", Title: "Inception", Rows: 5, SeatPerRow: 8},
	{
		ID: "dune", Title: "Dune Part Two", Rows: 4, SeatPerRow: 6},
}

func listMovies(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, movie)
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

type movieResponse struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Rows       int    `json:"rows"`
	SeatPerRow int    `json:"seats_per_row"`
}
