package routes

import (
	"net/http"

	"backend/controllers"
	"backend/middleware"
)

// SetupRouter initializes routes and attaches middlewares
func SetupRouter() http.Handler {
	mux := http.NewServeMux()

	// Health and Status endpoints
	mux.HandleFunc("/api/health", controllers.HealthCheckHandler)

	// Apply global middleware chain
	handler := middleware.CORSMiddleware(mux)
	handler = middleware.LoggerMiddleware(handler)

	return handler
}
