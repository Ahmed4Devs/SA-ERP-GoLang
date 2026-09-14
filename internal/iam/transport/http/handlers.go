package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"erp/internal/iam/application/services"
	"erp/internal/iam/domain"
	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	authService  *services.AuthService
	isProduction bool
}

func NewAuthHandler(authService *services.AuthService, isProduction bool) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		isProduction: isProduction,
	}
}

func (h *AuthHandler) Routes() http.Handler {
	r := chi.NewRouter()

	r.Post("/login", h.Login)
	r.Post("/logout", h.Logout)

	// Protected routes
	r.Group(func(protected chi.Router) {
		protected.Use(RequireAuth(h.authService))
		protected.Get("/me", h.Me)
		protected.Post("/logout-all", h.LogoutAll)
	})

	return r
}

type LoginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type LoginResponse struct {
	User      *domain.User `json:"user"`
	ExpiresAt string       `json:"expires_at"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Malformed JSON request body")
		return
	}

	req.Identifier = strings.TrimSpace(req.Identifier)
	if req.Identifier == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Identifier and password are required")
		return
	}

	clientIP := r.RemoteAddr
	userAgent := r.UserAgent()

	session, rawToken, user, err := h.authService.Login(r.Context(), req.Identifier, req.Password, clientIP, userAgent)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid username or password")
			return
		}
		if errors.Is(err, domain.ErrUserDisabled) {
			writeError(w, http.StatusForbidden, "USER_DISABLED", "Your account is disabled or suspended")
			return
		}
		if errors.Is(err, domain.ErrTooManyLoginAttempts) {
			writeError(w, http.StatusTooManyRequests, "TOO_MANY_ATTEMPTS", "Too many failed login attempts, please try again later")
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred during login")
		return
	}

	// Set HttpOnly session cookie
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    rawToken,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   h.isProduction,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(LoginResponse{
		User:      user,
		ExpiresAt: session.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token != "" {
		_ = h.authService.Logout(r.Context(), token)
	}

	// Expire cookie
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.isProduction,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

func (h *AuthHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	_ = h.authService.LogoutAll(r.Context(), user.ID)

	// Expire cookie
	cookie := &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.isProduction,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "All sessions terminated successfully"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"user": user,
	})
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": msg,
		},
	})
}
