package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"wallace/internal/config"
	"wallace/internal/httpx"
	"wallace/internal/modules/users"
)

const refreshCookieName = "refresh_token"
const refreshCookiePath = "/v1/auth"

type Service interface {
	Register(ctx context.Context, name, email, password string) (*User, error)
	Login(ctx context.Context, email, password string) (*Tokens, error)
	Refresh(ctx context.Context, refreshToken string) (*Tokens, error)
	Logout(ctx context.Context, refreshToken string) error
}

type Handler struct {
	service Service
	cfg     config.Config
}

func NewHandler(service Service, cfg config.Config) *Handler {
	return &Handler{service: service, cfg: cfg}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var body RegisterRequest
	if err := httpx.BindBody(r, &body); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	user, err := h.service.Register(r.Context(), body.Name, body.Email, body.Password)
	if err != nil {
		h.respondError(w, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, UserResponse{ID: user.ID, Name: user.Name, Email: user.Email})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var body LoginRequest
	if err := httpx.BindBody(r, &body); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	tokens, err := h.service.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		h.respondError(w, err)
		return
	}
	h.respondTokens(w, tokens)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshCookieName)
	if err != nil {
		h.respondError(w, ErrInvalidRefreshToken)
		return
	}
	tokens, err := h.service.Refresh(r.Context(), cookie.Value)
	if err != nil {
		h.respondError(w, err)
		return
	}
	h.respondTokens(w, tokens)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var refreshToken string
	if cookie, err := r.Cookie(refreshCookieName); err == nil {
		refreshToken = cookie.Value
	}
	if err := h.service.Logout(r.Context(), refreshToken); err != nil {
		h.respondError(w, err)
		return
	}

	cookie := h.refreshCookie("", time.Unix(1, 0))
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) respondTokens(w http.ResponseWriter, tokens *Tokens) {
	w.Header().Set("Cache-Control", "no-store")
	http.SetCookie(w, h.refreshCookie(tokens.RefreshToken, tokens.RefreshExpiresAt))
	httpx.RespondJSON(w, http.StatusOK, TokenResponse{
		AccessToken: tokens.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   tokens.ExpiresIn,
	})
}

func (h *Handler) refreshCookie(value string, expiresAt time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     refreshCookieName,
		Value:    value,
		Path:     refreshCookiePath,
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
		MaxAge:   int(h.cfg.RefreshTokenExpiry / time.Second),
	}
}

func (h *Handler) respondError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrHasherBusy) {
		w.Header().Set("Retry-After", "1")
		httpx.RespondError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "please try again shortly")
		return
	}
	if errors.Is(err, users.ErrUserEmailAlreadyExists) {
		httpx.RespondError(w, http.StatusConflict, "EMAIL_ALREADY_EXISTS", err.Error())
		return
	}
	if errors.Is(err, ErrInvalidCredentials) || errors.Is(err, ErrInvalidRefreshToken) {
		httpx.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}
	httpx.RespondError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "authentication request failed")
}
