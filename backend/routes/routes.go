package routes

import (
	"backend/config"
	"backend/controllers"
	"backend/middleware"
	"backend/repository"
	"backend/service"
	"net/http"
)

// SetupRouter initializes repositories, services, controllers, routes, and attaches middlewares
// like a backend reception
func SetupRouter(
	cfg *config.Config,
	userRepo repository.UserRepository,
	postRepo repository.PostRepository,
) http.Handler {
	mux := http.NewServeMux()

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	// The service validates photo references, so it needs to know which
	// Cloudinary cloud and folder our own uploads live in
	
	photoDeleter := service.NewCloudinaryPhotoDeleter(
		cfg.CloudinaryCloudName,
		cfg.CloudinaryAPIKey,
		cfg.CloudinaryAPISecret,
	)
	
	postService := service.NewPostService(
		postRepo,
		cfg.CloudinaryCloudName,
		cfg.CloudinaryUploadFolder,
		photoDeleter,
	)

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
	postController := controllers.NewPostController(
		postService,
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

	// Posts endpoints
	mux.HandleFunc("/api/posts", postController.CreatePostHandler)
	mux.HandleFunc("/api/posts/feed", postController.FeedHandler)
	mux.HandleFunc("/api/posts/user", postController.UserPostsHandler)
	mux.HandleFunc("/api/posts/{id}", postController.PostByIDHandler)

	// User Profile & Media Upload endpoints
	mux.HandleFunc("/api/upload/avatar", uploadController.UploadAvatarHandler)
	//for instant post media upload, this 2nd route is used.
	mux.HandleFunc("/api/upload/post-media", uploadController.UploadPostMediaHandler)
	// profile update & get route
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

	mux.HandleFunc("/api/users/", authController.GetUserProfileHandler)

	// Apply global middleware chain
	handler := middleware.CORSMiddleware(mux)
	handler = middleware.LoggerMiddleware(handler)

	return handler
}
