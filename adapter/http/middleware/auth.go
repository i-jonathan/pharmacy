package middleware

import (
	"context"
	"net/http"
	"pharmacy/config"
	"pharmacy/constant"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := config.SessionStore.Get(r, "session")
		if err != nil {
			session, _ = config.SessionStore.New(r, "session")
		}
		
		val, ok := session.Values[constant.UserSessionKey]
		userID, castOk := val.(int)
		if !ok || !castOk {
		    http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		    return
		}
		
		ctx := context.WithValue(r.Context(), constant.UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}