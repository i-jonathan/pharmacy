package config

import (
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

// SessionExpiry is when a session cookie saved at now stops working: the next 03:00.
func SessionExpiry(now time.Time) time.Time {
	expiry := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
	if !now.Before(expiry) {
		expiry = expiry.Add(24 * time.Hour)
	}
	return expiry
}

func NewSessionStore() *sessions.CookieStore {
	now := time.Now()
	expiry := SessionExpiry(now)

	s := sessions.NewCookieStore([]byte(Conf.SessionSecret))
	s.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   int(expiry.Sub(now).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteDefaultMode,
	}
	return s
}
