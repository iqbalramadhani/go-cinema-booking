package main

import (
	"log"
	"net/http"

	"github.com/iqbalramadhani/go-cinema-booking/internal/adapters/redis"
	"github.com/iqbalramadhani/go-cinema-booking/internal/booking"
	"github.com/iqbalramadhani/go-cinema-booking/internal/utils"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /movies", listMovies)
	mux.Handle("GET /", http.FileServer(http.Dir("static")))

	store := booking.NewRedistore(redis.NewClient("localhost:6379"))
	svc := booking.NewService(store)

	bookinghandler := booking.NewHandler(svc)

	mux.HandleFunc("GET /movies/{movieID}/seats", bookinghandler.ListSeats)
	mux.HandleFunc("POST /movies/{movieID}/seats/{seatID}/hold", bookinghandler.HoldSeat)

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
	utils.WriteJSON(w, http.StatusOK, movie)
}

type movieResponse struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Rows       int    `json:"rows"`
	SeatPerRow int    `json:"seats_per_row"`
}
