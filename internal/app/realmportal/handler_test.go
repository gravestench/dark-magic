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

func (control *fakeControl) VerifyEmail(_ context.Context, token string) (realm.Account, error) {
	control.verifiedToken = token
	return realm.Account{Name: "Alice", EmailVerified: true}, nil
}

func (control *fakeControl) CompletePasswordRecovery(_ context.Context, token, password string) error {
	control.recoveryToken, control.recoveryPassword = token, password
	return nil
}

func TestPortalDelegatesRealmAPIRequests(t *testing.T) {
	api := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/v1/status" {
			t.Fatalf("delegated path = %q", request.URL.Path)
		}
		writer.WriteHeader(http.StatusNoContent)
	})
	handler, err := NewHandler(&fakeControl{}, api, nil)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/status", nil))
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestVerificationLinkConsumesChallengeAndReturnsToExplicitLogin(t *testing.T) {
	control := &fakeControl{}
	handler, err := NewHandler(control, http.NotFoundHandler(), nil)
	if err != nil {
		t.Fatal(err)
	}
	get := httptest.NewRecorder()
	handler.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/verify?token=verification-token", nil))
	if get.Code != http.StatusOK || control.verifiedToken != "" {
		t.Fatalf("GET status=%d token=%q", get.Code, control.verifiedToken)
	}
	csrf, cookie := actionCSRF(t, get)
	form := url.Values{"token": {"verification-token"}, "csrf_token": {csrf}}
	request := httptest.NewRequest(http.MethodPost, "/verify", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
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

func TestRecoveryPageCompletesPasswordReplacement(t *testing.T) {
	control := &fakeControl{}
	handler, err := NewHandler(control, http.NotFoundHandler(), nil)
	if err != nil {
		t.Fatal(err)
	}
	get := httptest.NewRecorder()
	handler.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/recover?token=recovery-token", nil))
	if get.Code != http.StatusOK || !strings.Contains(get.Body.String(), `value="recovery-token"`) {
		t.Fatalf("recovery page status=%d body=%s", get.Code, get.Body.String())
	}
	csrf, cookie := actionCSRF(t, get)
	form := url.Values{"token": {"recovery-token"}, "csrf_token": {csrf}, "password": {"replacement-password"},
		"confirm_password": {"replacement-password"}}
	request := httptest.NewRequest(http.MethodPost, "/recover", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.AddCookie(cookie)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || control.recoveryToken != "recovery-token" ||
		control.recoveryPassword != "replacement-password" {
		t.Fatalf("status=%d token=%q password=%q", response.Code, control.recoveryToken, control.recoveryPassword)
	}
	if body := response.Body.String(); !strings.Contains(body, "Return to Dark Magic and log in") {
		t.Fatalf("recovery response = %s", body)
	}
}

func TestRecoveryRejectsMissingActionCSRF(t *testing.T) {
	control := &fakeControl{}
	handler, err := NewHandler(control, http.NotFoundHandler(), nil)
	if err != nil {
		t.Fatal(err)
	}
	form := url.Values{"token": {"recovery-token"}, "password": {"replacement-password"},
		"confirm_password": {"replacement-password"}}
	request := httptest.NewRequest(http.MethodPost, "/recover", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden || control.recoveryToken != "" {
		t.Fatalf("status=%d token=%q", response.Code, control.recoveryToken)
	}
}

var csrfInputPattern = regexp.MustCompile(`name="csrf_token" value="([a-f0-9]{64})"`)

func actionCSRF(t *testing.T, response *httptest.ResponseRecorder) (string, *http.Cookie) {
	t.Helper()
	match := csrfInputPattern.FindStringSubmatch(response.Body.String())
	if len(match) != 2 {
		t.Fatalf("CSRF input missing from response: %s", response.Body.String())
	}
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == actionCSRFCookie {
			if !cookie.Secure || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode || cookie.Path != "/" {
				t.Fatalf("CSRF cookie = %#v", cookie)
			}
			return match[1], cookie
		}
	}
	t.Fatal("CSRF cookie missing")
	return "", nil
}
