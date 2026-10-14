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
	//loads config
	cfg := config.LoadConfig()

	// Connect to PostgreSQL and run migrations
	db, err := database.ConnectPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}

	//close db when the function is returned/server is closed
	defer db.Close()

	log.Printf("✅ Connected to PostgreSQL at %s", cfg.DatabaseURL)
	//loading database tables
	if err := database.Migrate(db); err != nil {
		log.Fatalf("❌ Failed to run database migrations: %v", err)
	}
	log.Printf("✅ PostgreSQL migrations applied successfully")
	// a tool which knows How to get user data from Postgresql
	//(db= connection to database, NewPostgresUserRepository= name of the tool,
	// userRepo= variable that holds the tool)
	userRepo := repository.NewPostgresUserRepository(db)
	postRepo := repository.NewPostgresPostRepository(db)
	//router is the main handler for all HTTP requests to handlers
	router := routes.SetupRouter(cfg, userRepo, postRepo)
	//address of the server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 Starting XClone Backend on http://localhost%s (Environment: %s)\n", addr, cfg.Environment)
	//starts the server and listens for requests
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server exited with error: %v", err)
	}
}
