package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"wallace/internal/config"
	"wallace/internal/database"
	"wallace/internal/router"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.NewDB(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("create database connection: %v", err)
	}
	defer db.Close()

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router.NewRouter(db, cfg),
	}

	log.Printf("Wallace's Server listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
