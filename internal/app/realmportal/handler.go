// Package realmportal owns the browser-facing Realm experience. Realm API and
// business logic stay in package realm; this package only adapts them to HTML.
package realmportal

import (
	"context"
	"embed"
	"errors"
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

// Control exposes the account operations initiated by the browser portal. The
// narrow interface keeps HTTP presentation independent from Realm persistence.
type Control interface {
	VerifyEmail(context.Context, string) (realm.Account, error)
	CompletePasswordRecovery(context.Context, string, string) error
}

// Handler routes browser account actions while delegating all other requests
// to the Realm API. Its embedded assets make the public server self-contained.
type Handler struct {
	control Control
	api     http.Handler
	assets  http.Handler
	web     http.Handler
	verify  *template.Template
	recover *template.Template
}

// NewHandler constructs the browser portal around an existing Realm API. It
// parses all embedded templates up front so serving cannot fail during startup.
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

// ServeHTTP reserves the portal's account routes and delegates the remaining
// surface unchanged, preserving the API handler's routing and error behavior.
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
	case request.Method == http.MethodGet && handler.assets != nil && isPortalAssetPath(request.URL.Path):
		handler.assets.ServeHTTP(writer, request)
	default:
		handler.api.ServeHTTP(writer, request)
	}
}

// isPortalAssetPath recognizes only the generated asset namespaces owned by
// the optional asset handler, leaving unknown account paths to the Realm API.
func isPortalAssetPath(requestPath string) bool {
	return strings.HasPrefix(requestPath, "/account/assets/") ||
		strings.HasPrefix(requestPath, "/account/fonts/") ||
		strings.HasPrefix(requestPath, "/account/roster/")
}

// serveStatic maps a stable public asset route to its embedded filename. The
// cloned request prevents portal routing from mutating state observed upstream.
func (handler *Handler) serveStatic(writer http.ResponseWriter, request *http.Request, name, contentType string) {
	setPortalHeaders(writer.Header())
	writer.Header().Set("Content-Type", contentType)
	writer.Header().Set("Cache-Control", "no-cache")

	clone := request.Clone(request.Context())
	clone.URL.Path = "/" + name
	handler.web.ServeHTTP(writer, clone)
}
