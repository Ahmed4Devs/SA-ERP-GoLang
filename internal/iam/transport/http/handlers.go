package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"erp/internal/iam/application/services"
	"erp/internal/iam/domain"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService  *services.AuthService
	mfaService   *services.MFAService
	isProduction bool
}

func NewAuthHandler(authService *services.AuthService, mfaService *services.MFAService, isProduction bool) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		mfaService:   mfaService,
		isProduction: isProduction,
	}
}

func (h *AuthHandler) Routes() http.Handler {
	r := chi.NewRouter()

	r.Post("/login", h.Login)
	r.Post("/login/mfa", h.CompleteMFALogin)
	r.Post("/logout", h.Logout)

	// Protected routes
	r.Group(func(protected chi.Router) {
		protected.Use(RequireAuth(h.authService))
		protected.Get("/me", h.Me)
		protected.Post("/logout-all", h.LogoutAll)
		protected.Get("/mfa/status", h.MFAStatus)
		protected.Post("/mfa/enroll", h.EnrollMFA)
		protected.Post("/mfa/confirm", h.ConfirmMFA)
		protected.Post("/mfa/disable", h.DisableMFA)
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
		if errors.Is(err, domain.ErrMFARequired) {
			challengeToken, expiresAt, challengeErr := h.mfaService.CreateLoginChallenge(r.Context(), user.ID, clientIP, userAgent)
			if challengeErr != nil {
				writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to create MFA challenge")
				return
			}
			writeJSON(w, http.StatusAccepted, map[string]any{
				"mfa_required":    true,
				"challenge_token": challengeToken,
				"expires_at":      expiresAt.Format(time.RFC3339),
			})
			return
		}
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

type MFALoginRequest struct {
	ChallengeToken string `json:"challenge_token"`
	Code           string `json:"code"`
}

func (h *AuthHandler) CompleteMFALogin(w http.ResponseWriter, r *http.Request) {
	var req MFALoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.ChallengeToken) == "" || strings.TrimSpace(req.Code) == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "Challenge token and MFA code are required")
		return
	}
	challenge, err := h.mfaService.CompleteLogin(r.Context(), req.ChallengeToken, req.Code)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "INVALID_MFA_CODE", "Invalid, expired, or already used MFA code")
		return
	}
	session, rawToken, user, err := h.authService.CreateSessionForUser(r.Context(), challenge.UserID, challenge.IPAddress, challenge.UserAgent)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to create authenticated session")
		return
	}
	h.setSessionCookie(w, rawToken, session.ExpiresAt)
	writeJSON(w, http.StatusOK, LoginResponse{User: user, ExpiresAt: session.ExpiresAt.Format(time.RFC3339)})
}

func (h *AuthHandler) MFAStatus(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}
	enabled, err := h.mfaService.IsEnabled(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to determine MFA status")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"enabled": enabled})
}

func (h *AuthHandler) EnrollMFA(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}
	enrollment, err := h.mfaService.BeginEnrollment(r.Context(), user)
	if err != nil {
		if errors.Is(err, domain.ErrMFAAlreadyEnabled) {
			writeError(w, http.StatusConflict, "MFA_ALREADY_ENABLED", "MFA is already enabled")
			return
		}
		writeError(w, http.StatusServiceUnavailable, "MFA_UNAVAILABLE", "MFA is not configured on this server")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusCreated, enrollment)
}

type MFACodeRequest struct {
	Code string `json:"code"`
}

func (h *AuthHandler) ConfirmMFA(w http.ResponseWriter, r *http.Request) {
	h.handleMFACode(w, r, h.mfaService.ConfirmEnrollment)
}

func (h *AuthHandler) DisableMFA(w http.ResponseWriter, r *http.Request) {
	h.handleMFACode(w, r, h.mfaService.Disable)
}

func (h *AuthHandler) handleMFACode(w http.ResponseWriter, r *http.Request, action func(context.Context, uuid.UUID, string) error) {
	user := UserFromContext(r.Context())
	var req MFACodeRequest
	if user == nil || json.NewDecoder(r.Body).Decode(&req) != nil || strings.TrimSpace(req.Code) == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_FAILED", "MFA code is required")
		return
	}
	if err := action(r.Context(), user.ID, req.Code); err != nil {
		writeError(w, http.StatusUnauthorized, "INVALID_MFA_CODE", "Invalid or already used MFA code")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandler) setSessionCookie(w http.ResponseWriter, rawToken string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{Name: SessionCookieName, Value: rawToken, Path: "/", Expires: expiresAt, HttpOnly: true, Secure: h.isProduction, SameSite: http.SameSiteLaxMode})
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
	writeJSON(w, status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": msg,
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
