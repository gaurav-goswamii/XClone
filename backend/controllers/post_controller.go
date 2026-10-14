package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"backend/service"
	"backend/utils"
)

// PostController handles incoming post HTTP endpoints
type PostController struct {
	postService *service.PostService
	jwtSecret   string
}

// NewPostController creates an instance of PostController
func NewPostController(postService *service.PostService, jwtSecret string) *PostController {
	return &PostController{
		postService: postService,
		jwtSecret:   jwtSecret,
	}
}

// CreatePostRequest represents request body for creating a post
type CreatePostRequest struct {
	Content string `json:"content"`
}

// CreatePostHandler handles POST /api/posts
func (c *PostController) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	cookie, err := r.Cookie(AuthCookieName)
	if err != nil || cookie.Value == "" {
		writeError(w, http.StatusUnauthorized, "missing authentication token")
		return
	}

	claims, err := utils.ValidateJWT(cookie.Value, c.jwtSecret)
	if err != nil {
		clearAuthCookie(w)
		writeError(w, http.StatusUnauthorized, "invalid or expired authentication token")
		return
	}

	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	post, err := c.postService.CreatePost(claims.UserID, req.Content)
	if err != nil {
		if errors.Is(err, service.ErrEmptyContent) {
			writeError(w, http.StatusBadRequest, "post content cannot be empty")
			return
		}
		if errors.Is(err, service.ErrContentTooLong) {
			writeError(w, http.StatusBadRequest, "post content exceeds 280 characters")
			return
		}
		if errors.Is(err, service.ErrUnauthorizedPost) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create post")
		return
	}

	writeJSON(w, http.StatusCreated, APIResponse{
		Status:  "success",
		Message: "post created successfully",
		Data: map[string]interface{}{
			"post": post,
		},
	})
}

// FeedHandler handles GET /api/posts/feed
func (c *PostController) FeedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	limit := 20
	offset := 0

	query := r.URL.Query()
	if l := query.Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := query.Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	posts, err := c.postService.GetFeed(limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch feed")
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Status: "success",
		Data: map[string]interface{}{
			"posts": posts,
		},
	})
}

// UserPostsHandler handles GET /api/posts/user?username=...
func (c *PostController) UserPostsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	username := strings.TrimSpace(r.URL.Query().Get("username"))
	if username == "" {
		writeError(w, http.StatusBadRequest, "username parameter is required")
		return
	}

	limit := 20
	offset := 0

	query := r.URL.Query()
	if l := query.Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := query.Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	posts, err := c.postService.GetUserPosts(username, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch user posts")
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Status: "success",
		Data: map[string]interface{}{
			"posts": posts,
		},
	})
}
