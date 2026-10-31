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
	case errors.Is(err, service.ErrEmptyPost):
		writeError(w, http.StatusBadRequest, "post must have text or at least one photo")
	case errors.Is(err, service.ErrTooManyPhotos):
		writeError(w, http.StatusBadRequest, "a post can have at most 4 photos")
	case errors.Is(err, service.ErrInvalidMediaRef):
		writeError(w, http.StatusBadRequest, "invalid photo reference")
	case errors.Is(err, service.ErrMediaNotConfigured):
		// Mirrors the upload endpoint's 503, so a deploy missing its Cloudinary
		// settings fails the same recognisable way at both steps
		writeError(w, http.StatusServiceUnavailable, "media uploads are not configured")
	case errors.Is(err, service.ErrUnauthorizedPost):
		writeError(w, http.StatusUnauthorized, "unauthorized")
	default:
		writeError(w, http.StatusInternalServerError, fallback)
	}
}

// maxPostBodyBytes caps the create request body. 280 runes of text plus four
// public IDs sits comfortably under 1KB, so this is already generous.
const maxPostBodyBytes = 64 << 10

// PostMediaRequest references a photo already uploaded through
// POST /api/upload/post-media. Creating and editing a post speak the same
// wire language, so both decode into this type.
//
// There is deliberately no url field. The caller names an asset and the service
// builds the URL, so the wire contract cannot be used to point a post at a host
// we do not control.
//
// The json tags are load-bearing rather than decorative: encoding/json matches
// keys case-insensitively but not across punctuation, so without them an
// incoming "public_id" would never reach PublicID and every photo would arrive
// anonymous.
type PostMediaRequest struct {
	PublicID string `json:"public_id"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

// toMediaRefs translates the wire shape into the service's own type. The
// controller carries the references across; it does not judge them.
func toMediaRefs(media []PostMediaRequest) []service.MediaRef {
	refs := make([]service.MediaRef, 0, len(media))
	for _, item := range media {
		refs = append(refs, service.MediaRef{
			PublicID: item.PublicID,
			Width:    item.Width,
			Height:   item.Height,
		})
	}
	return refs
}

// CreatePostRequest represents request body for creating a post
type CreatePostRequest struct {
	Content string             `json:"content"`
	Media   []PostMediaRequest `json:"media"`
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

	// Bound the body before decoding; this endpoint had no limit at all
	r.Body = http.MaxBytesReader(w, r.Body, maxPostBodyBytes)

	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body is too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	post, err := c.postService.CreatePost(claims.UserID, req.Content, toMediaRefs(req.Media))
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

// UpdatePostRequest represents request body for editing a post.
// Media is a pointer to a slice so the handler can tell three states apart:
// absent (edit the caption, keep the photos), present but empty (remove every
// photo), and present with entries (this is the new photo set).
type UpdatePostRequest struct {
	Content string              `json:"content"`
	Media   *[]PostMediaRequest `json:"media"`
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

	// Same bound as create: an edit carries the same caption and public IDs
	r.Body = http.MaxBytesReader(w, r.Body, maxPostBodyBytes)

	var req UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body is too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// nil refs tell the service to keep the photos the post already has. The
	// service owns that decision because it has the post loaded already, and
	// because it checks ownership first: resolving the existing photos here
	// would mean reading a post before knowing the caller may touch it.
	var refs []service.MediaRef
	if req.Media != nil {
		refs = toMediaRefs(*req.Media)
	}

	post, err := c.postService.UpdatePost(claims.UserID, postID, req.Content, refs)
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
