package routes

import (
	"net/http"
	"backend/config"
	"backend/controllers"
	"backend/middleware"
	"backend/realtime"
	"backend/repository"
	"backend/service"
)

func SetupRouter(
	cfg *config.Config,
	userRepo repository.UserRepository,
	postRepo repository.PostRepository,
	commentRepo repository.CommentRepository,
	chatRepo repository.ChatRepository,
	messageRepo repository.MessageRepository,
) http.Handler {
	mux := http.NewServeMux()

	authService := service.NewAuthService(userRepo, cfg.JWTSecret)

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

	hub := realtime.NewHub()
	chatService := service.NewChatService(chatRepo, messageRepo, userRepo, hub)

	commentService := service.NewCommentService(commentRepo, postRepo)

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
	chatController := controllers.NewChatController(
		chatService,
		cfg.JWTSecret,
		hub,
	)
	commentController := controllers.NewCommentController(
		commentService,
		cfg.JWTSecret,
	)

	mux.HandleFunc("/api/health", controllers.HealthCheckHandler)

	mux.HandleFunc("/api/auth/signup", authController.SignupHandler)
	mux.HandleFunc("/api/auth/login", authController.LoginHandler)
	mux.HandleFunc("/api/auth/logout", authController.LogoutHandler)
	mux.HandleFunc("/api/auth/check-username", authController.CheckUsernameHandler)
	mux.HandleFunc("/api/auth/me", authController.MeHandler)

	mux.HandleFunc("/api/posts", postController.CreatePostHandler)
	mux.HandleFunc("/api/posts/feed", postController.FeedHandler)
	mux.HandleFunc("/api/posts/user", postController.UserPostsHandler)
	mux.HandleFunc("/api/posts/{id}/like", postController.PostLikeHandler)
	mux.HandleFunc("/api/posts/{id}", postController.PostByIDHandler)
	mux.HandleFunc("/api/posts/{id}/comments", commentController.PostCommentsHandler)
	mux.HandleFunc("/api/comments/{id}", commentController.DeleteCommentHandler)

	mux.HandleFunc("/api/upload/avatar", uploadController.UploadAvatarHandler)
	mux.HandleFunc("/api/upload/post-media", uploadController.UploadPostMediaHandler)
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

	mux.HandleFunc("/api/chats", chatController.ChatsHandler)
 	mux.HandleFunc("/api/chats/{id}/messages", chatController.MessagesHandler)
    mux.HandleFunc("/api/chats/{id}/ws", chatController.ChatWebSocketHandler)
 
	handler := middleware.CORSMiddleware(mux)
	handler = middleware.LoggerMiddleware(handler)

	return handler
}
