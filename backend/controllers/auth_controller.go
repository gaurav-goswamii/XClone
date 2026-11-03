package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"backend/service"
	"backend/utils"
)

type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type AuthController struct {
	authService *service.AuthService
	jwtSecret   string
}

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
		MaxAge:   30 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   false,
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

func (c *AuthController) SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req service.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
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

func (c *AuthController) MeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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

type UpdateProfileRequest struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Bio       string `json:"bio"`
}

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

func (c *AuthController) GetUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	username := r.URL.Query().Get("username")
	if strings.TrimSpace(username) == "" {
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
