package middleware

import (
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"pharmacy/config"
	"pharmacy/httperror"
	"strings"

	"github.com/gorilla/csrf"
)

func csrfAuthKey() []byte {
	secret := config.Conf.CSRFKey
	if secret == "" {
		secret = config.Conf.SessionSecret
	}
	if secret == "" {
		panic("CSRF_KEY or SESSION_SECRET must be configured")
	}
	key := sha256.Sum256([]byte("pharmacy/csrf/v1:" + secret))
	return key[:]
}

// csrfErrorHandler answers an app call in JSON with the code the SPA retries on (csrf_failed);
// a page load gets plain text. API calls are told apart the same way as in AuthMiddleware.
func csrfErrorHandler(w http.ResponseWriter, r *http.Request) {
	if !wantsPage(r) {
		refusal := httperror.Forbidden("Your session token expired. Try the request again.", nil)
		refusal.Reason = "csrf_failed"
		refusal.JSONRespond(w)
		return
	}
	http.Error(w, "Forbidden - CSRF token validation failed", http.StatusForbidden)
}

func CSRFTokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(map[string]string{"token": csrf.Token(r)})
}

var csrfProtection = csrf.Protect(
	csrfAuthKey(),
	csrf.Secure(false),
	csrf.HttpOnly(true),
	csrf.Path("/"),
	csrf.ErrorHandler(http.HandlerFunc(csrfErrorHandler)),
)

func CSRFMiddleware(next http.Handler) http.Handler {
	protected := csrfProtection(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwardedProto := strings.ToLower(r.Header.Get("X-Forwarded-Proto"))
		if r.TLS == nil && (forwardedProto == "" || forwardedProto == "http") {
			r = csrf.PlaintextHTTPRequest(r)
		}
		protected.ServeHTTP(w, r)
	})
}
