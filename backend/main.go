package main

import (
	"fmt"
	"log"
	"net/http"

	"backend/config"
	"backend/database"
	"backend/repository"
	"backend/routes"
)

func main() {
	//loads config
	cfg := config.LoadConfig()

	// Connect to PostgreSQL and run migrations
	db, err := database.ConnectPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()
	log.Printf("✅ Connected to PostgreSQL at %s", cfg.DatabaseURL)

	if err := database.Migrate(db); err != nil {
		log.Fatalf("❌ Failed to run database migrations: %v", err)
	}
	log.Printf("✅ PostgreSQL migrations applied successfully")

	userRepo := repository.NewPostgresUserRepository(db)
	router := routes.SetupRouter(cfg, userRepo)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 Starting XClone Backend on http://localhost%s (Environment: %s)\n", addr, cfg.Environment)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server exited with error: %v", err)
	}
}
