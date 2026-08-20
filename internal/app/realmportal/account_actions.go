package realmportal

import (
	"html/template"
	"net/http"
	"strings"
)

// accountActionPage is the complete template state for either browser action;
// empty tokens intentionally select the terminal success or error views.
type accountActionPage struct {
	Token   string
	CSRF    string
	Error   string
	Success bool
}

// verificationPage presents a verification challenge without consuming it.
// Separating GET from POST prevents link scanners from confirming accounts.
func (handler *Handler) verificationPage(writer http.ResponseWriter, request *http.Request) {
	page, status := handler.beginAccountAction(writer, request, "This verification link is invalid.")
	handler.renderAccountAction(writer, status, handler.verify, "verify.html", page)
}

// verifyEmail consumes a verification challenge only after validating the
// bounded form and its CSRF token, so cross-site requests cannot change state.
func (handler *Handler) verifyEmail(writer http.ResponseWriter, request *http.Request) {
	if !handler.parseValidActionForm(writer, request) {
		handler.renderAccountAction(
			writer,
			http.StatusForbidden,
			handler.verify,
			"verify.html",
			accountActionPage{Error: "This verification request is invalid."},
		)

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

// recoveryPage presents a password-reset challenge without consuming it. The
// issued CSRF token remains paired with the challenge across validation errors.
func (handler *Handler) recoveryPage(writer http.ResponseWriter, request *http.Request) {
	page, status := handler.beginAccountAction(writer, request, "This password-reset link is invalid.")
	handler.renderAccountAction(writer, status, handler.recover, "recover.html", page)
}

// completeRecovery validates both password fields before handing the change to
// the control plane, retaining the challenge only when the user can retry.
func (handler *Handler) completeRecovery(writer http.ResponseWriter, request *http.Request) {
	if !handler.parseValidActionForm(writer, request) {
		handler.renderAccountAction(
			writer,
			http.StatusForbidden,
			handler.recover,
			"recover.html",
			accountActionPage{Error: "Invalid password-reset request."},
		)

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

// beginAccountAction validates the link token and issues the CSRF secret needed
// by a subsequent POST. Invalid links never receive an actionable form.
func (handler *Handler) beginAccountAction(
	writer http.ResponseWriter,
	request *http.Request,
	invalidTokenMessage string,
) (accountActionPage, int) {
	token := strings.TrimSpace(request.URL.Query().Get("token"))
	if token == "" {
		return accountActionPage{Error: invalidTokenMessage}, http.StatusBadRequest
	}

	return accountActionPage{
		Token: token,
		CSRF:  handler.issueActionCSRF(writer),
	}, http.StatusOK
}

// parseValidActionForm bounds attacker-controlled input before parsing and then
// authenticates the form against its secure cookie before callers inspect it.
func (handler *Handler) parseValidActionForm(writer http.ResponseWriter, request *http.Request) bool {
	request.Body = http.MaxBytesReader(writer, request.Body, maximumFormBytes)
	if err := request.ParseForm(); err != nil {
		return false
	}

	return handler.validActionCSRF(request)
}

// renderAccountAction applies the portal's security headers before committing
// the response status, ensuring error pages receive the same browser policy.
func (handler *Handler) renderAccountAction(
	writer http.ResponseWriter,
	status int,
	pageTemplate *template.Template,
	name string,
	page accountActionPage,
) {
	setPortalHeaders(writer.Header())
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(status)
	_ = pageTemplate.ExecuteTemplate(writer, name, page)
}
