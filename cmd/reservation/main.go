package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/stockyard-dev/stockyard-reservation/internal/server"
	"github.com/stockyard-dev/stockyard-reservation/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9806"
	}
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./reservation-data"
	}

	db, err := store.Open(dataDir)
	if err != nil {
		log.Fatalf("reservation: %v", err)
	}
	defer db.Close()

	srv := server.New(db, server.DefaultLimits())

	fmt.Printf("\n  Reservation — Self-hosted reservation and table management\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n  Questions? hello@stockyard.dev — I read every message\n\n", port, port)
	log.Printf("reservation: listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, srv))
}
