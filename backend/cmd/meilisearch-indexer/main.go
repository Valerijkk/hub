package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/valerii/hub/backend/internal/platform/config"
	"github.com/valerii/hub/backend/internal/platform/database"
	searchmod "github.com/valerii/hub/backend/internal/search"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	db, _, err := database.NewPostgresClient(ctx, cfg.Postgres.URL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	indexer := searchmod.NewMeiliIndexerFromEnv()
	if indexer == nil {
		log.Fatal("MEILI_HOST not set")
	}
	n, err := indexer.IndexAllIssues(ctx, db)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("indexed %d issues into Meilisearch", n)
	os.Exit(0)
}
