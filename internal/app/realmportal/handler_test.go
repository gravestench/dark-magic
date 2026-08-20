package realmportal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/gravestench/dark-magic/internal/app/realm"
)

type fakeControl struct {
	verifiedToken    string
	recoveryToken    string
	recoveryPassword string
}

// VerifyEmail records the challenge handed across the HTTP boundary so tests
// can distinguish page rendering from a state-changing verification request.
func (control *fakeControl) VerifyEmail(_ context.Context, token string) (realm.Account, error) {
	control.verifiedToken = token
	return realm.Account{Name: "Alice", EmailVerified: true}, nil
}

// CompletePasswordRecovery records both sensitive inputs without introducing
// persistence behavior that could obscure the portal contract under test.
func (control *fakeControl) CompletePasswordRecovery(_ context.Context, token, password string) error {
	control.recoveryToken, control.recoveryPassword = token, password
	return nil
}

// TestPortalDelegatesRealmAPIRequests verifies that the portal reserves only
// its browser routes and leaves unrelated API requests unchanged.
func TestPortalDelegatesRealmAPIRequests(t *testing.T) {
	// Assert at the delegated boundary so an accidental rewrite cannot be hidden
	// by a response-only check.
	api := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/status" {
			t.Fatalf("delegated path = %q", request.URL.Path)
		}

		writer.WriteHeader(http.StatusNoContent)
	})
	handler := newTestHandler(t, &fakeControl{}, api)

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/status", nil))

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}
}

// TestVerificationLinkConsumesChallengeAndReturnsToExplicitLogin verifies that
// GET is inert and the CSRF-authenticated POST consumes the challenge once.
func TestVerificationLinkConsumesChallengeAndReturnsToExplicitLogin(t *testing.T) {
	control := &fakeControl{}
	handler := newTestHandler(t, control, http.NotFoundHandler())

	page, csrf, cookie := requestActionPage(t, handler, "/verify?token=verification-token")
	if page.Code != http.StatusOK || control.verifiedToken != "" {
		t.Fatalf("GET status=%d token=%q", page.Code, control.verifiedToken)
	}

	form := url.Values{"token": {"verification-token"}, "csrf_token": {csrf}}
	response := submitActionForm(handler, "/verify", form, cookie)

	if response.Code != http.StatusOK || control.verifiedToken != "verification-token" {
		t.Fatalf("POST status=%d token=%q", response.Code, control.verifiedToken)
	}

	if body := response.Body.String(); !strings.Contains(body, "Return to Dark Magic and log in") {
		t.Fatalf("verification response = %s", body)
	}

	csp := response.Header().Get("Content-Security-Policy")
	if strings.Contains(csp, "unsafe-inline") || !strings.Contains(csp, "script-src 'self'") {
		t.Fatalf("CSP = %q", csp)
	}
}

// TestRecoveryPageCompletesPasswordReplacement verifies that a matching pair
// reaches the control plane only after the browser proves the CSRF challenge.
func TestRecoveryPageCompletesPasswordReplacement(t *testing.T) {
	control := &fakeControl{}
	handler := newTestHandler(t, control, http.NotFoundHandler())

	page, csrf, cookie := requestActionPage(t, handler, "/recover?token=recovery-token")
	if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), `value="recovery-token"`) {
		t.Fatalf("recovery page status=%d body=%s", page.Code, page.Body.String())
	}

	form := url.Values{
		"token":            {"recovery-token"},
		"csrf_token":       {csrf},
		"password":         {"replacement-password"},
		"confirm_password": {"replacement-password"},
	}
	response := submitActionForm(handler, "/recover", form, cookie)

	if response.Code != http.StatusOK || control.recoveryToken != "recovery-token" ||
		control.recoveryPassword != "replacement-password" {
		t.Fatalf("status=%d token=%q password=%q", response.Code, control.recoveryToken, control.recoveryPassword)
	}

	if body := response.Body.String(); !strings.Contains(body, "Return to Dark Magic and log in") {
		t.Fatalf("recovery response = %s", body)
	}
}

// TestRecoveryRejectsMissingActionCSRF verifies that a plausible reset form is
// still inert when it was not initiated by the portal's own GET response.
func TestRecoveryRejectsMissingActionCSRF(t *testing.T) {
	control := &fakeControl{}
	handler := newTestHandler(t, control, http.NotFoundHandler())
	form := url.Values{
		"token":            {"recovery-token"},
		"password":         {"replacement-password"},
		"confirm_password": {"replacement-password"},
	}

	response := submitActionForm(handler, "/recover", form, nil)

	if response.Code != http.StatusForbidden || control.recoveryToken != "" {
		t.Fatalf("status=%d token=%q", response.Code, control.recoveryToken)
	}
}

var csrfInputPattern = regexp.MustCompile(`name="csrf_token" value="([a-f0-9]{64})"`)

// newTestHandler constructs a portal with no generated asset routes, keeping
// each test focused on account actions and API delegation.
func newTestHandler(t *testing.T, control Control, api http.Handler) http.Handler {
	t.Helper()

	handler, err := NewHandler(control, api, nil)
	if err != nil {
		t.Fatal(err)
	}

	return handler
}

// requestActionPage starts a browser action and returns the paired form and
// cookie secrets required for an authenticated POST.
func requestActionPage(
	t *testing.T,
	handler http.Handler,
	target string,
) (*httptest.ResponseRecorder, string, *http.Cookie) {
	t.Helper()

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, target, nil))

	if response.Code != http.StatusOK {
		t.Fatalf("GET %s status=%d body=%s", target, response.Code, response.Body.String())
	}

	csrf, cookie := actionCSRF(t, response)

	return response, csrf, cookie
}

// submitActionForm reproduces a browser form POST and optionally attaches the
// action cookie, allowing rejection tests to omit that credential explicitly.
func submitActionForm(
	handler http.Handler,
	target string,
	form url.Values,
	cookie *http.Cookie,
) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if cookie != nil {
		request.AddCookie(cookie)
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	return response
}

// actionCSRF extracts and validates both halves of the double-submit token so
// callers cannot accidentally build authenticated requests from a weak cookie.
func actionCSRF(t *testing.T, response *httptest.ResponseRecorder) (string, *http.Cookie) {
	t.Helper()

	match := csrfInputPattern.FindStringSubmatch(response.Body.String())
	if len(match) != 2 {
		t.Fatalf("CSRF input missing from response: %s", response.Body.String())
	}

	for _, cookie := range response.Result().Cookies() {
		if cookie.Name != actionCSRFCookie {
			continue
		}

		if !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode || cookie.Path != "/" {
			t.Fatalf("CSRF cookie = %#v", cookie)
		}

		return match[1], cookie
	}

	t.Fatal("CSRF cookie missing")

	return "", nil
}
