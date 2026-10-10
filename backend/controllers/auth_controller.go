package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"backend/service"
	"backend/utils"
)

// APIResponse represents standard JSON response envelope
type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// AuthController handles incoming authentication HTTP endpoints
type AuthController struct {
	authService *service.AuthService
	jwtSecret   string
}

// NewAuthController creates an instance of AuthController
func NewAuthController(authService *service.AuthService, jwtSecret string) *AuthController {
	return &AuthController{
		authService: authService,
		jwtSecret:   jwtSecret,
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, APIResponse{
		Status:  "error",
		Message: message,
	})
}

const AuthCookieName = "x_auth_token"

func setAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     AuthCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   30 * 24 * 60 * 60, // 30 days
		HttpOnly: true,
		Secure:   false, // Set to false for local development over HTTP
		SameSite: http.SameSiteLaxMode,
	})
}

func clearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     AuthCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

// SignupHandler handles POST /api/auth/signup
func (c *AuthController) SignupHandler(w http.ResponseWriter, r *http.Request) {
	//checks for post method and restricts
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	//checks the json format and validates the data
	var req service.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	//actual signup function in service
	user, err := c.authService.Signup(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, APIResponse{
		Status:  "success",
		Message: "Account created successfully",
		Data: map[string]interface{}{
			"user": user,
		},
	})
}

// LoginHandler handles POST /api/auth/login
func (c *AuthController) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req service.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, token, err := c.authService.Login(req)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	setAuthCookie(w, token)

	writeJSON(w, http.StatusOK, APIResponse{
		Status:  "success",
		Message: "Login successful",
		Data: map[string]interface{}{
			"user": user,
		},
	})
}

// LogoutHandler handles POST /api/auth/logout and clears the authentication cookie
func (c *AuthController) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	clearAuthCookie(w)

	writeJSON(w, http.StatusOK, APIResponse{
		Status:  "success",
		Message: "Logged out successfully",
	})
}

// CheckUsernameHandler handles GET /api/auth/check-username?username=xyz
func (c *AuthController) CheckUsernameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	username := r.URL.Query().Get("username")
	if strings.TrimSpace(username) == "" {
		writeError(w, http.StatusBadRequest, "username parameter is required")
		return
	}

	available, msg, err := c.authService.CheckUsernameAvailability(username)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "success",
		"available": available,
		"message":   msg,
	})
}

// MeHandler handles GET /api/auth/me (validates cookie or Authorization Bearer token)
func (c *AuthController) MeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var tokenStr string

	// 1. Check HTTP-only cookie first
	cookie, err := r.Cookie(AuthCookieName)

if err != nil || cookie.Value == "" {
    writeError(w, http.StatusUnauthorized, "missing authentication token")
    return
}

tokenStr = cookie.Value

	if tokenStr == "" {
		writeError(w, http.StatusUnauthorized, "missing authentication token or cookie")
		return
	}

	claims, err := utils.ValidateJWT(tokenStr, c.jwtSecret)
	if err != nil {
		clearAuthCookie(w)
		writeError(w, http.StatusUnauthorized, "invalid or expired authentication token")
		return
	}

	user, err := c.authService.GetUserByID(claims.UserID)
	if err != nil {
		clearAuthCookie(w)
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Status: "success",
		Data: map[string]interface{}{
			"user": user,
		},
	})
}

// UpdateProfileRequest represents request body for updating user profile
type UpdateProfileRequest struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Bio       string `json:"bio"`
}

// UpdateProfileHandler handles PUT /api/users/profile
func (c *AuthController) UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var tokenStr string
	cookie, err := r.Cookie(AuthCookieName)

if err != nil || cookie.Value == "" {
    writeError(w, http.StatusUnauthorized, "missing authentication token")
    return
}

tokenStr = cookie.Value

	if tokenStr == "" {
		writeError(w, http.StatusUnauthorized, "missing authentication token or cookie")
		return
	}

	claims, err := utils.ValidateJWT(tokenStr, c.jwtSecret)
	if err != nil {
		clearAuthCookie(w)
		writeError(w, http.StatusUnauthorized, "invalid or expired authentication token")
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	updatedUser, err := c.authService.UpdateUserProfile(claims.UserID, req.Name, req.AvatarURL, req.Bio)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Status:  "success",
		Message: "Profile updated successfully",
		Data: map[string]interface{}{
			"user": updatedUser,
		},
	})
}

// GetUserProfileHandler handles GET /api/users/profile?username=xyz or GET /api/users/{username}
func (c *AuthController) GetUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	username := r.URL.Query().Get("username")
	if strings.TrimSpace(username) == "" {
		// Try parsing from path suffix: /api/users/{username}
		path := strings.TrimPrefix(r.URL.Path, "/api/users/")
		path = strings.Trim(path, "/")
		if path != "" && path != "profile" {
			username = path
		}
	}

	if strings.TrimSpace(username) == "" {
		writeError(w, http.StatusBadRequest, "username parameter is required")
		return
	}

	user, err := c.authService.GetUserByUsername(username)
	if err != nil || user == nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Status: "success",
		Data: map[string]interface{}{
			"user": user,
		},
	})
}
