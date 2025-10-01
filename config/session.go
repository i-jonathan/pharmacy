package config

import (
	"time"

	"github.com/gorilla/sessions"
)

var SessionStore = func() *sessions.CookieStore {
	now := time.Now()
	expiry := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
	if !now.Before(expiry) {
		expiry = expiry.Add(24 * time.Hour)
	}
	var s *sessions.CookieStore
	s = sessions.NewCookieStore([]byte(Conf.SessionSecret))
	s.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   int(expiry.Sub(now).Seconds()),
		HttpOnly: true,
		SameSite: 1,
		// Secure: true, // enable if HTTPS
	}
	return s
}()
