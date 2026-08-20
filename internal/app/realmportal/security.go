package realmportal

import (
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"net/http"
)

const portalContentSecurityPolicy = "default-src 'none'; img-src 'self'; style-src 'self'; " +
	"script-src 'self'; form-action 'self'; base-uri 'none'; frame-ancestors 'none'"

// issueActionCSRF creates the double-submit secret used by one account action.
// Failure to obtain cryptographic randomness leaves the rendered form inert.
func (handler *Handler) issueActionCSRF(writer http.ResponseWriter) string {
	var token [32]byte
	if _, err := rand.Read(token[:]); err != nil {
		return ""
	}

	value := fmt.Sprintf("%x", token[:])
	http.SetCookie(writer, newActionCSRFCookie(value, 15*60))

	return value
}

// validActionCSRF compares fixed-length cookie and form secrets in constant
// time, preventing malformed values and timing leaks from authorizing a POST.
func (handler *Handler) validActionCSRF(request *http.Request) bool {
	cookie, err := request.Cookie(actionCSRFCookie)
	form := request.Form.Get("csrf_token")

	return err == nil && len(cookie.Value) == 64 && len(form) == 64 &&
		subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(form)) == 1
}

// clearActionCSRF expires the action cookie after a successful state change so
// a completed browser flow cannot accidentally reuse its authorization secret.
func (handler *Handler) clearActionCSRF(writer http.ResponseWriter) {
	http.SetCookie(writer, newActionCSRFCookie("", -1))
}

// newActionCSRFCookie centralizes the security attributes shared by issued and
// expired cookies, preventing cleanup from weakening the browser boundary.
func newActionCSRFCookie(value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     actionCSRFCookie,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}

// setPortalHeaders denies ambient browser capabilities on every portal-owned
// response, including validation errors and embedded static assets.
func setPortalHeaders(header http.Header) {
	header.Set("Cache-Control", "no-store")
	header.Set("Content-Security-Policy", portalContentSecurityPolicy)
	header.Set("Referrer-Policy", "no-referrer")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Frame-Options", "DENY")
}
