package config

import "github.com/gorilla/sessions"

var SessionStore = func() *sessions.CookieStore {
	var s *sessions.CookieStore
	s = sessions.NewCookieStore([]byte(Conf.SessionSecret))
	s.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   18 * 60 * 60,
		HttpOnly: true,
		SameSite: 1,
		// Secure: true, // enable if HTTPS
	}
	return s
}()
