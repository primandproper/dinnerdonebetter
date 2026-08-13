package passkey

import (
	"net/http"
	"time"

	"github.com/primandproper/platform-go/v10/cookies"
)

// CookieName is the name the passkey auth cookie is set under.
//
// platform-go v9 dropped cookies.Config.CookieName: BuildCookie there takes the
// name per call, so nothing in that module ever read the field. The name is this
// application's, so it is stated here.
const CookieName = "ddb_auth"

// BuildCookie provides a consistent way of constructing an HTTP cookie for session auth.
// See https://www.calhoun.io/securing-cookies-in-go/
func BuildCookie(cfg *cookies.Config, value string) *http.Cookie {
	expiry := time.Now().Add(cfg.Lifetime)

	return &http.Cookie{
		Name:     CookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.SecureOnly,
		Expires:  expiry,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(time.Until(expiry).Seconds()),
	}
}

// ClearCookie returns a cookie that clears the auth cookie when set.
func ClearCookie(cfg *cookies.Config) *http.Cookie {
	return &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.SecureOnly,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	}
}
