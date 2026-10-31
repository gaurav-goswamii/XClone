package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"backend/service"
	"backend/utils"
)

// maxChatBodyBytes caps a chat request body. A text message needs far less
// than a post with photo references, so this is already generous.
const maxChatBodyBytes = 16 << 10

// ChatController handles the direct-message HTTP endpoints
type ChatController struct {
	chatService *service.ChatService
	jwtSecret   string
}

// NewChatController creates an instance of ChatController
func NewChatController(chatService *service.ChatService, jwtSecret string) *ChatController {
	return &ChatController{
		chatService: chatService,
		jwtSecret:   jwtSecret,
	}
}

// requireAuth validates the auth cookie and returns the caller's JWT claims.
// It writes the failure response itself, so callers only need to check ok.
func (c *ChatController) requireAuth(w http.ResponseWriter, r *http.Request) (*utils.JWTClaims, bool) {
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

// StartChatRequest represents the request body for opening a chat. The client
// knows the other person by handle, never by internal ID.
type StartChatRequest struct {
	Username string `json:"username"`
}

// ChatsHandler handles POST /api/chats
func (c *ChatController) ChatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	claims, ok := c.requireAuth(w, r)
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxChatBodyBytes)

	var req StartChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body is too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Handles are stored lowercase and people type them however they like
	username := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(req.Username, "@")))
	if username == "" {
		writeError(w, http.StatusBadRequest, "username is required")
		return
	}

	chat, err := c.chatService.StartChatWithUsername(claims.UserID, username)
	if err != nil {
		writeChatError(w, err, "failed to open chat")
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Status: "success",
		Data: map[string]interface{}{
			"chat": chat,
		},
	})
}

// writeChatError maps chat service errors onto HTTP status codes. It copies
// post_controller.go's switch rather than auth_controller.go's err.Error()
// pass-through, so no internal string ever reaches a client.
func writeChatError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, service.ErrChatNotFound):
		writeError(w, http.StatusNotFound, "chat not found")
	case errors.Is(err, service.ErrRecipientNotFound):
		writeError(w, http.StatusNotFound, "user not found")
	case errors.Is(err, service.ErrCannotMessageSelf):
		writeError(w, http.StatusBadRequest, "you cannot message yourself")
	case errors.Is(err, service.ErrInvalidUser):
		writeError(w, http.StatusBadRequest, "invalid user")
	case errors.Is(err, service.ErrEmptyMessage):
		writeError(w, http.StatusBadRequest, "message cannot be empty")
	case errors.Is(err, service.ErrMessageTooLong):
		writeError(w, http.StatusBadRequest, "message exceeds 2000 characters")
	case errors.Is(err, service.ErrInvalidEncoding):
		writeError(w, http.StatusBadRequest, "message contains invalid characters")
	case errors.Is(err, service.ErrUnauthorized):
		// 403, not 401: the caller is authenticated, they are simply not in
		// this chat. A 401 would make the browser think the session expired.
		writeError(w, http.StatusForbidden, "you are not a participant in this chat")
	default:
		writeError(w, http.StatusInternalServerError, fallback)
	}
}
