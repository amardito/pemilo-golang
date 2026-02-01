package config

import (
	"os"

	"github.com/gorilla/sessions"
)

var SessionStore *sessions.CookieStore

func InitSession() {
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		secret = "default-secret-change-in-production"
	}

	SessionStore = sessions.NewCookieStore([]byte(secret))
	SessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   os.Getenv("GIN_MODE") == "release",
		SameSite: 4, // SameSiteStrictMode
	}
}

func init() {
	InitSession()
}
