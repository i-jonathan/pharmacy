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

func RequirePermissions(mode constant.RequirePermissionMode, requiredPermissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			store := config.NewSessionStore()
			session, err := store.Get(r, "session")
			if err != nil {
				session, _ = store.New(r, "session")
			}

			userPermissions, ok := session.Values["permissions"].(map[string]bool)
			if !ok {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			switch mode {
			case constant.RequireAllPermissions:
				for _, p := range requiredPermissions {
					if !userPermissions[p] {
						http.Error(w, "Forbidden", http.StatusForbidden)
						return
					}
				}
			case constant.RequireAnyPermissions:
				for _, p := range requiredPermissions {
					if userPermissions[p] {
						ctx := context.WithValue(r.Context(), constant.PermissionsSessionKey, userPermissions)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			default:
				http.Error(w, "Server Error", http.StatusInternalServerError)
			}

			ctx := context.WithValue(r.Context(), constant.PermissionsSessionKey, userPermissions)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AddPermissionsToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		store := config.NewSessionStore()
		session, err := store.Get(r, "session")
		if err != nil {
			session, _ = store.New(r, "session")
		}

		userPermissions, _ := session.Values["permissions"].(map[string]bool)
		ctx := context.WithValue(r.Context(), constant.PermissionsSessionKey, userPermissions)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
