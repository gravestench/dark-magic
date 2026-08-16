// Package realmportal owns the browser-facing Realm experience. Realm API and
// business logic stay in package realm; this package only adapts them to HTML.
package realmportal

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gravestench/dark-magic/internal/app/realm"
)

const maximumFormBytes = 8 << 10
const actionCSRFCookie = "__Host-dark_magic_realm_action"

//go:embed web/*
var embeddedWeb embed.FS

type Control interface {
	VerifyEmail(context.Context, string) (realm.Account, error)
	CompletePasswordRecovery(context.Context, string, string) error
}

type Handler struct {
	control Control
	api     http.Handler
	assets  http.Handler
	web     http.Handler
	verify  *template.Template
	recover *template.Template
}

type accountActionPage struct {
	Token   string
	CSRF    string
	Error   string
	Success bool
}

func NewHandler(control Control, api, assets http.Handler) (http.Handler, error) {
	if control == nil || api == nil {
		return nil, errors.New("realm portal requires a control plane and API handler")
	}
	verify, err := template.ParseFS(embeddedWeb, "web/verify.html")
	if err != nil {
		return nil, err
	}
	recover, err := template.ParseFS(embeddedWeb, "web/recover.html")
	if err != nil {
		return nil, err
	}
	static, err := fs.Sub(embeddedWeb, "web")
	if err != nil {
		return nil, err
	}
	handler := &Handler{
		control: control,
		api:     api,
		assets:  assets,
		web:     http.FileServer(http.FS(static)),
		verify:  verify,
		recover: recover,
	}
	return handler, nil
}

func (handler *Handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	switch {
	case request.Method == http.MethodGet && request.URL.Path == "/verify":
		handler.verificationPage(writer, request)
	case request.Method == http.MethodPost && request.URL.Path == "/verify":
		handler.verifyEmail(writer, request)
	case request.Method == http.MethodGet && request.URL.Path == "/recover":
		handler.recoveryPage(writer, request)
	case request.Method == http.MethodPost && request.URL.Path == "/recover":
		handler.completeRecovery(writer, request)
	case request.Method == http.MethodGet && request.URL.Path == "/account/portal.css":
		handler.serveStatic(writer, request, "portal.css", "text/css; charset=utf-8")
	case request.Method == http.MethodGet && request.URL.Path == "/account/portal.js":
		handler.serveStatic(writer, request, "portal.js", "text/javascript; charset=utf-8")
	case request.Method == http.MethodGet && handler.assets != nil &&
		(strings.HasPrefix(request.URL.Path, "/account/assets/") ||
			strings.HasPrefix(request.URL.Path, "/account/fonts/") ||
			strings.HasPrefix(request.URL.Path, "/account/roster/")):
		handler.assets.ServeHTTP(writer, request)
	default:
		handler.api.ServeHTTP(writer, request)
	}
}

func (handler *Handler) verificationPage(writer http.ResponseWriter, request *http.Request) {
	token := strings.TrimSpace(request.URL.Query().Get("token"))
	page := accountActionPage{Token: token}
	status := http.StatusOK
	if token == "" {
		status, page.Error = http.StatusBadRequest, "This verification link is invalid."
	} else {
		page.CSRF = handler.issueActionCSRF(writer)
	}
	handler.renderAccountAction(writer, status, handler.verify, "verify.html", page)
}

func (handler *Handler) verifyEmail(writer http.ResponseWriter, request *http.Request) {
	request.Body = http.MaxBytesReader(writer, request.Body, maximumFormBytes)
	if err := request.ParseForm(); err != nil || !handler.validActionCSRF(request) {
		handler.renderAccountAction(writer, http.StatusForbidden, handler.verify, "verify.html",
			accountActionPage{Error: "This verification request is invalid."})
		return
	}
	token := strings.TrimSpace(request.Form.Get("token"))
	page := accountActionPage{}
	status := http.StatusOK
	if token == "" {
		status, page.Error = http.StatusBadRequest, "This verification link is invalid."
	} else if _, err := handler.control.VerifyEmail(request.Context(), token); err != nil {
		status, page.Error = http.StatusBadRequest, "This verification link has expired or was already used."
	} else {
		page.Success = true
		handler.clearActionCSRF(writer)
	}
	handler.renderAccountAction(writer, status, handler.verify, "verify.html", page)
}

func (handler *Handler) recoveryPage(writer http.ResponseWriter, request *http.Request) {
	token := strings.TrimSpace(request.URL.Query().Get("token"))
	page := accountActionPage{Token: token}
	status := http.StatusOK
	if token == "" {
		status, page.Error = http.StatusBadRequest, "This password-reset link is invalid."
	} else {
		page.CSRF = handler.issueActionCSRF(writer)
	}
	handler.renderAccountAction(writer, status, handler.recover, "recover.html", page)
}

func (handler *Handler) completeRecovery(writer http.ResponseWriter, request *http.Request) {
	request.Body = http.MaxBytesReader(writer, request.Body, maximumFormBytes)
	if err := request.ParseForm(); err != nil || !handler.validActionCSRF(request) {
		handler.renderAccountAction(writer, http.StatusForbidden, handler.recover, "recover.html",
			accountActionPage{Error: "Invalid password-reset request."})
		return
	}
	token := strings.TrimSpace(request.Form.Get("token"))
	password := request.Form.Get("password")
	page := accountActionPage{Token: token, CSRF: request.Form.Get("csrf_token")}
	if password != request.Form.Get("confirm_password") {
		page.Error = "The passwords do not match."
		handler.renderAccountAction(writer, http.StatusBadRequest, handler.recover, "recover.html", page)
		return
	}
	if err := handler.control.CompletePasswordRecovery(request.Context(), token, password); err != nil {
		page.Error = "This password-reset link is invalid, expired, or the password was not accepted."
		handler.renderAccountAction(writer, http.StatusBadRequest, handler.recover, "recover.html", page)
		return
	}
	page.Token, page.Success = "", true
	handler.clearActionCSRF(writer)
	handler.renderAccountAction(writer, http.StatusOK, handler.recover, "recover.html", page)
}

func (handler *Handler) issueActionCSRF(writer http.ResponseWriter) string {
	var token [32]byte
	if _, err := rand.Read(token[:]); err != nil {
		return ""
	}
	value := fmt.Sprintf("%x", token[:])
	http.SetCookie(writer, &http.Cookie{Name: actionCSRFCookie, Value: value, Path: "/", MaxAge: 15 * 60,
		Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode})
	return value
}

func (handler *Handler) validActionCSRF(request *http.Request) bool {
	cookie, err := request.Cookie(actionCSRFCookie)
	form := request.Form.Get("csrf_token")
	return err == nil && len(cookie.Value) == 64 && len(form) == 64 &&
		subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(form)) == 1
}

func (handler *Handler) clearActionCSRF(writer http.ResponseWriter) {
	http.SetCookie(writer, &http.Cookie{Name: actionCSRFCookie, Path: "/", MaxAge: -1,
		Secure: true, HttpOnly: true, SameSite: http.SameSiteStrictMode})
}

func (handler *Handler) renderAccountAction(writer http.ResponseWriter, status int, pageTemplate *template.Template, name string, page accountActionPage) {
	setPortalHeaders(writer.Header())
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(status)
	_ = pageTemplate.ExecuteTemplate(writer, name, page)
}

func (handler *Handler) serveStatic(writer http.ResponseWriter, request *http.Request, name, contentType string) {
	setPortalHeaders(writer.Header())
	writer.Header().Set("Content-Type", contentType)
	writer.Header().Set("Cache-Control", "no-cache")
	clone := request.Clone(request.Context())
	clone.URL.Path = "/" + name
	handler.web.ServeHTTP(writer, clone)
}

func setPortalHeaders(header http.Header) {
	header.Set("Cache-Control", "no-store")
	header.Set("Content-Security-Policy", "default-src 'none'; img-src 'self'; style-src 'self'; script-src 'self'; form-action 'self'; base-uri 'none'; frame-ancestors 'none'")
	header.Set("Referrer-Policy", "no-referrer")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("X-Frame-Options", "DENY")
}
