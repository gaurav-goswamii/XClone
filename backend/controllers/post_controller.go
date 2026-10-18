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

// requireAuth validates the auth cookie and returns the caller's JWT claims.
// It writes the failure response itself, so callers only need to check ok.
func (c *PostController) requireAuth(w http.ResponseWriter, r *http.Request) (*utils.JWTClaims, bool) {
	cookie, err := r.Cookie(AuthCookieName)
	if err != nil || cookie.Value == "" {
		writeError(w, http.StatusUnauthorized, "missing authentication token")
		return nil, false
	}

	claims, err := utils.ValidateJWT(cookie.Value, c.jwtSecret)
	if err != nil {
		clearAuthCookie(w)
		writeError(w, http.StatusUnauthorized, "invalid or expired authentication token")
		return nil, false
	}

	return claims, true
}

// writePostError maps post service errors onto HTTP status codes
func writePostError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, service.ErrPostNotFound):
		writeError(w, http.StatusNotFound, "post not found")
	case errors.Is(err, service.ErrNotPostOwner):
		writeError(w, http.StatusForbidden, "you can only modify your own posts")
	case errors.Is(err, service.ErrEmptyContent):
		writeError(w, http.StatusBadRequest, "post content cannot be empty")
	case errors.Is(err, service.ErrContentTooLong):
		writeError(w, http.StatusBadRequest, "post content exceeds 280 characters")
	case errors.Is(err, service.ErrUnauthorizedPost):
		writeError(w, http.StatusUnauthorized, "unauthorized")
	default:
		writeError(w, http.StatusInternalServerError, fallback)
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

	claims, ok := c.requireAuth(w, r)
	if !ok {
		return
	}

	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	post, err := c.postService.CreatePost(claims.UserID, req.Content)
	if err != nil {
		writePostError(w, err, "failed to create post")
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

// UpdatePostRequest represents request body for editing a post
type UpdatePostRequest struct {
	Content string `json:"content"`
}

// PostByIDHandler handles GET, PUT and DELETE on /api/posts/{id}
func (c *PostController) PostByIDHandler(w http.ResponseWriter, r *http.Request) {
	postID := strings.TrimSpace(r.PathValue("id"))
	if postID == "" {
		writeError(w, http.StatusBadRequest, "post id is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		c.getPost(w, postID)
	case http.MethodPut:
		c.updatePost(w, r, postID)
	case http.MethodDelete:
		c.deletePost(w, r, postID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// getPost serves a single post publicly, like the feed endpoints
func (c *PostController) getPost(w http.ResponseWriter, postID string) {
	post, err := c.postService.GetPost(postID)
	if err != nil {
		writePostError(w, err, "failed to fetch post")
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Status: "success",
		Data: map[string]interface{}{
			"post": post,
		},
	})
}

// updatePost rewrites the content of a post owned by the caller
func (c *PostController) updatePost(w http.ResponseWriter, r *http.Request, postID string) {
	claims, ok := c.requireAuth(w, r)
	if !ok {
		return
	}

	var req UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	post, err := c.postService.UpdatePost(claims.UserID, postID, req.Content)
	if err != nil {
		writePostError(w, err, "failed to update post")
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Status:  "success",
		Message: "post updated successfully",
		Data: map[string]interface{}{
			"post": post,
		},
	})
}

// deletePost permanently removes a post owned by the caller
func (c *PostController) deletePost(w http.ResponseWriter, r *http.Request, postID string) {
	claims, ok := c.requireAuth(w, r)
	if !ok {
		return
	}

	if err := c.postService.DeletePost(claims.UserID, postID); err != nil {
		writePostError(w, err, "failed to delete post")
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Status:  "success",
		Message: "post deleted successfully",
		Data: map[string]interface{}{
			"id": postID,
		},
	})
}
