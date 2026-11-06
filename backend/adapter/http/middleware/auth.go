package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"pharmacy/config"
	"pharmacy/httperror"
	"pharmacy/internal/constant"
	"strings"
)

// wantsPage reports whether the browser is loading a page rather than the app calling an API.
func wantsPage(r *http.Request) bool {
	return r.Method == http.MethodGet && strings.Contains(r.Header.Get("Accept"), "text/html")
}

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
			// A call from the app gets 401: a redirect would be followed to the 200 login
			// page, which the caller would take for success. Only page loads are remembered
			// and sent to sign in.
			if !wantsPage(r) {
				httperror.Unauthorized("Your session has ended; sign in again.", nil).
					JSONRespond(w)
				return
			}
			session.Values["next"] = r.URL.RequestURI()
			session.Save(r, w)
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), constant.UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// forbid refuses a request without the permission. The app shows the server's message, so an API
// call gets one in JSON; a page load keeps the plain refusal.
func forbid(w http.ResponseWriter, r *http.Request) {
	if wantsPage(r) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	httperror.Forbidden(
		"You don't have permission to do this. If it was just given to you, sign in again.", nil,
	).JSONRespond(w)
}

func RequirePermissions(mode constant.RequirePermissionMode, requiredPermissions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			store := config.NewSessionStore()
			session, err := store.Get(r, "session")
			if err != nil {
				session, _ = store.New(r, "session")
			}

			permJSON, ok := session.Values["permissions"].(string)
			if !ok {
				forbid(w, r)
				return
			}

			var userPermissions map[string]bool
			if err := json.Unmarshal([]byte(permJSON), &userPermissions); err != nil {
				forbid(w, r)
				return
			}

			switch mode {
			case constant.RequireAllPermissions:
				for _, p := range requiredPermissions {
					if !userPermissions[p] {
						forbid(w, r)
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
				// None of the permissions matched: refuse, don't fall through to the handler.
				forbid(w, r)
				return
			default:
				http.Error(w, "Server Error", http.StatusInternalServerError)
				return
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

		permJSON, ok := session.Values[constant.PermissionsSessionKey].(string)
		if !ok {
			// If missing or invalid, treat as empty
			userPermissions := make(map[string]bool)
			ctx := context.WithValue(r.Context(), constant.PermissionsSessionKey, userPermissions)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Unmarshal into map[string]bool
		var userPermissions map[string]bool
		if err := json.Unmarshal([]byte(permJSON), &userPermissions); err != nil {
			userPermissions = make(map[string]bool)
		}

		// Store in context
		ctx := context.WithValue(r.Context(), constant.PermissionsSessionKey, userPermissions)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
