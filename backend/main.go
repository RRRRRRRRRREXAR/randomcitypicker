package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"randomcitypicker/db"
	"randomcitypicker/handlers"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/cities.db"
	}

	if err := db.Init(dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init db: %v\n", err)
		os.Exit(1)
	}
	defer db.DB.Close()

	seedCountries := os.Getenv("SEED_COUNTRIES")
	if seedCountries == "" {
		seedCountries = "FR"
	}
	codes := strings.Split(seedCountries, ",")
	for i := range codes {
		codes[i] = strings.TrimSpace(codes[i])
	}
	if err := db.SeedCountries(codes); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to seed db: %v\n", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/cities/random", handlers.EnableCORS(handlers.RandomCity))
	mux.HandleFunc("/api/cities/confirm", handlers.EnableCORS(handlers.ConfirmPick))
	mux.HandleFunc("/api/cities/picked", handlers.EnableCORS(handlers.PickedCities))
	mux.HandleFunc("/api/cities/reset", handlers.EnableCORS(handlers.ResetPicks))
	mux.HandleFunc("/api/countries", handlers.EnableCORS(handlers.Countries))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
