package main

import (
	"fmt"
	"log"
	"net/http"

	"backend/config"
	"backend/routes"
)

func main() {
	cfg := config.LoadConfig()

	router := routes.SetupRouter()

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Starting XClone Backend on http://localhost%s (Environment: %s)\n", addr, cfg.Environment)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server exited with error: %v", err)
	}
}
