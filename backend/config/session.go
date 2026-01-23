package config

import (
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

func NewSessionStore() *sessions.CookieStore {
	now := time.Now()
	expiry := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
	if !now.Before(expiry) {
		expiry = expiry.Add(24 * time.Hour)
	}

	s := sessions.NewCookieStore([]byte(Conf.SessionSecret))
	s.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   int(expiry.Sub(now).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteDefaultMode,
	}
	return s
}
