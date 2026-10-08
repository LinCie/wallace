package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL        string
	Port               int
	JWTSecret          string
	JWTExpiry          time.Duration
	RefreshTokenExpiry time.Duration
	CookieSecure       bool
	CORSAllowedOrigin  string
}

func NewConfig() (Config, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	port := 8080
	if value := os.Getenv("PORT"); value != "" {
		parsedPort, err := strconv.Atoi(value)
		if err != nil {
			return Config{}, err
		}
		port = parsedPort
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-me"
	}
	minutes := 5
	if v := os.Getenv("JWT_EXPIRES_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			minutes = n
		}
	}
	jwtExpiry := time.Duration(minutes) * time.Minute

	days := 30
	if v := os.Getenv("REFRESH_TOKEN_EXPIRES_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			days = n
		}
	}
	refreshTokenExpiry := time.Duration(days) * 24 * time.Hour

	// Secure by default; set COOKIE_SECURE=false only for local development
	// over plain HTTP.
	cookieSecure := true
	if v := os.Getenv("COOKIE_SECURE"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			cookieSecure = b
		}
	}

	corsAllowedOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if corsAllowedOrigin == "" {
		corsAllowedOrigin = "http://localhost:5173"
	}

	return Config{
		DatabaseURL:        databaseURL,
		Port:               port,
		JWTSecret:          jwtSecret,
		JWTExpiry:          jwtExpiry,
		RefreshTokenExpiry: refreshTokenExpiry,
		CookieSecure:       cookieSecure,
		CORSAllowedOrigin:  corsAllowedOrigin,
	}, nil
}
