package middleware

import (
	"pharmacy/config"

	"github.com/gorilla/csrf"
)

var CSRFMiddleware = csrf.Protect(
	[]byte(config.Conf.CSRFKey),
	csrf.Secure(false),
)