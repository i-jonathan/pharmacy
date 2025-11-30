package middleware

import (
	"context"
	"net/http"
	"pharmacy/config"
	"pharmacy/internal/constant"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		store := config.NewSessionStore()
		session, err := store.Get(r, "session")
		if err != nil {
			session, _ = store.New(r, "session")
		}

		val, ok := session.Values[constant.UserSessionKey]
		userID, castOk := val.(int)
		if !ok || !castOk {
			session.Values["next"] = r.URL.RequestURI()
			session.Save(r, w)
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), constant.UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
