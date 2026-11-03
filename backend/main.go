package main

import (
	"backend/config"
	"backend/database"
	"backend/repository"
	"backend/routes"
	"fmt"
	"log"
	"net/http"
)

func main() {
	cfg := config.LoadConfig()

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
	postRepo := repository.NewPostgresPostRepository(db)
	chatRepo := repository.NewPostgresChatRepository(db)
	messageRepo := repository.NewPostgresMessageRepository(db)
	router := routes.SetupRouter(cfg, userRepo, postRepo, chatRepo, messageRepo)
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 Starting XClone Backend on http://localhost%s (Environment: %s)\n", addr, cfg.Environment)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server exited with error: %v", err)
	}
}
