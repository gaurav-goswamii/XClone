package routes

import (
	"net/http"

	"backend/config"
	"backend/controllers"
	"backend/middleware"
	"backend/repository"
	"backend/service"
)

// SetupRouter initializes repositories, services, controllers, routes, and attaches middlewares
func SetupRouter(cfg *config.Config, repos ...repository.UserRepository) http.Handler {
	mux := http.NewServeMux()

	var userRepo repository.UserRepository
	if len(repos) > 0 && repos[0] != nil {
		userRepo = repos[0]
	} else {
		userRepo = repository.NewMemoryUserRepository()
	}

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)

	// Initialize controllers
	authController := controllers.NewAuthController(
		authService,
		cfg.JWTSecret,
	)
	uploadController := controllers.NewUploadController(
		cfg.CloudinaryCloudName,
		cfg.CloudinaryAPIKey,
		cfg.CloudinaryAPISecret,
		cfg.CloudinaryUploadFolder,
		cfg.JWTSecret,
	)

	// Health and Status endpoints
	mux.HandleFunc("/api/health", controllers.HealthCheckHandler)

	// Authentication & Signup endpoints (Single Signup Endpoint)
	mux.HandleFunc("/api/auth/signup", authController.SignupHandler)
	mux.HandleFunc("/api/auth/login", authController.LoginHandler)
	mux.HandleFunc("/api/auth/logout", authController.LogoutHandler)
	mux.HandleFunc("/api/auth/check-username", authController.CheckUsernameHandler)
	mux.HandleFunc("/api/auth/me", authController.MeHandler)

	// User Profile & Media Upload endpoints
	mux.HandleFunc("/api/upload/avatar", uploadController.UploadAvatarHandler)
	mux.HandleFunc("/api/users/profile", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut || r.Method == http.MethodPatch {
			authController.UpdateProfileHandler(w, r)
			return
		}
		if r.Method == http.MethodGet {
			authController.GetUserProfileHandler(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
	mux.HandleFunc("/api/users/by-username", authController.GetUserProfileHandler)
	mux.HandleFunc("/api/users/", authController.GetUserProfileHandler)

	// Apply global middleware chain
	handler := middleware.CORSMiddleware(mux)
	handler = middleware.LoggerMiddleware(handler)

	return handler
}
