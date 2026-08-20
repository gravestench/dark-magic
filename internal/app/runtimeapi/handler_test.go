package runtimeapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gravestench/dark-magic/internal/app/host"
)

// testComponent exposes lifecycle counts and a controlled start failure without adding synchronization that could hide
// whether the HTTP adapter preserves the manager's synchronous behavior.
type testComponent struct {
	startErr error
	starts   int
	stops    int
}

// Start records lifecycle calls so routing tests can prove that actions reach the manager exactly once. A configured
// failure exercises the API's conflict and diagnostic response paths without replacing manager behavior.
func (component *testComponent) Start(context.Context) error {
	component.starts++

	return component.startErr
}

// Stop records lifecycle calls so restart and disable tests can distinguish those actions from a status-only response.
func (component *testComponent) Stop(context.Context) error {
	component.stops++

	return nil
}

// newInstance returns the same instrumented component for each manager construction. Tests rely on that stable identity
// to observe lifecycle call counts across disable, enable, and restart transitions.
func (component *testComponent) newInstance(context.Context) (host.Component, error) {
	return component, nil
}

// TestHandlerControlsRegisteredComponents preserves the original end-to-end contract: enabling a registered component
// returns its enabled status, and listing then exposes the same component.
func TestHandlerControlsRegisteredComponents(t *testing.T) {
	manager := host.NewManager()
	component := &testComponent{}
	registerTestComponent(t, manager, "scripts", nil, component)
	handler := New("", manager).Handler()

	enableResponse := serveRequest(handler, http.MethodPost, "/v1/components/scripts/enable")
	if enableResponse.Code != http.StatusOK || !strings.Contains(enableResponse.Body.String(), `"state":"enabled"`) {
		t.Fatalf("enable: code=%d body=%s", enableResponse.Code, enableResponse.Body.String())
	}

	listResponse := serveRequest(handler, http.MethodGet, "/v1/components")
	if listResponse.Code != http.StatusOK || !strings.Contains(listResponse.Body.String(), `"id":"scripts"`) {
		t.Fatalf("list: code=%d body=%s", listResponse.Code, listResponse.Body.String())
	}
}

// TestTransitionRestartsEnabledComponent verifies that restart reaches both lifecycle phases exactly once and does not
// respond before the manager returns the component to its enabled state.
func TestTransitionRestartsEnabledComponent(t *testing.T) {
	manager := host.NewManager()
	component := &testComponent{}
	registerTestComponent(t, manager, "scripts", nil, component)
	handler := New("", manager).Handler()

	requireResponseCode(t, serveRequest(handler, http.MethodPost, "/v1/components/scripts/enable"), http.StatusOK)
	restartResponse := serveRequest(handler, http.MethodPost, "/v1/components/scripts/restart")

	requireResponseCode(t, restartResponse, http.StatusOK)

	if component.starts != 2 || component.stops != 1 {
		t.Fatalf("restart lifecycle calls: starts=%d stops=%d", component.starts, component.stops)
	}

	if !strings.Contains(restartResponse.Body.String(), `"state":"enabled"`) {
		t.Fatalf("restart body=%s", restartResponse.Body.String())
	}
}

// TestTransitionRequiresCascadeForActiveDependents verifies that a normal disable preserves dependency safety while
// the explicit cascade query selects the manager operation that also shuts down dependents.
func TestTransitionRequiresCascadeForActiveDependents(t *testing.T) {
	manager := host.NewManager()
	dependency := &testComponent{}
	dependent := &testComponent{}

	registerTestComponent(t, manager, "renderer", nil, dependency)
	registerTestComponent(t, manager, "scripts", []string{"renderer"}, dependent)
	handler := New("", manager).Handler()

	requireResponseCode(t, serveRequest(handler, http.MethodPost, "/v1/components/scripts/enable"), http.StatusOK)
	blockedResponse := serveRequest(handler, http.MethodPost, "/v1/components/renderer/disable")

	requireResponseCode(t, blockedResponse, http.StatusConflict)

	if !strings.Contains(blockedResponse.Body.String(), "active dependents") {
		t.Fatalf("non-cascade disable body=%s", blockedResponse.Body.String())
	}

	cascadeResponse := serveRequest(handler, http.MethodPost, "/v1/components/renderer/disable?cascade=true")
	requireResponseCode(t, cascadeResponse, http.StatusOK)

	if dependency.stops != 1 || dependent.stops != 1 {
		t.Fatalf("cascade lifecycle calls: renderer stops=%d scripts stops=%d", dependency.stops, dependent.stops)
	}
}

// TestTransitionRejectsUnknownActions verifies that unsupported action names are client errors and never reach the
// manager, which protects the lifecycle vocabulary from silently expanding.
func TestTransitionRejectsUnknownActions(t *testing.T) {
	manager := host.NewManager()
	component := &testComponent{}
	registerTestComponent(t, manager, "scripts", nil, component)
	handler := New("", manager).Handler()

	response := serveRequest(handler, http.MethodPost, "/v1/components/scripts/remove")

	requireResponseCode(t, response, http.StatusBadRequest)

	if response.Body.String() != "{\"error\":\"unknown action\"}\n" {
		t.Fatalf("unknown action body=%q", response.Body.String())
	}

	if component.starts != 0 || component.stops != 0 {
		t.Fatalf("unexpected lifecycle calls: starts=%d stops=%d", component.starts, component.stops)
	}
}

// TestListReportsComponentFailures verifies that manager errors become diagnostic strings in list responses without
// exposing Go error internals or losing registration order.
func TestListReportsComponentFailures(t *testing.T) {
	manager := host.NewManager()
	working := &testComponent{}
	failing := &testComponent{startErr: errors.New("fixture start failure")}

	registerTestComponent(t, manager, "working", nil, working)
	registerTestComponent(t, manager, "failing", nil, failing)
	handler := New("", manager).Handler()

	failureResponse := serveRequest(handler, http.MethodPost, "/v1/components/failing/enable")
	requireResponseCode(t, failureResponse, http.StatusConflict)

	listResponse := serveRequest(handler, http.MethodGet, "/v1/components")
	requireResponseCode(t, listResponse, http.StatusOK)

	wantBody := "[" +
		"{\"id\":\"working\",\"desired\":false,\"state\":\"disabled\"}," +
		"{\"id\":\"failing\",\"desired\":true,\"state\":\"failed\"," +
		"\"error\":\"host: enable \\\"failing\\\": fixture start failure\"}" +
		"]\n"
	if listResponse.Body.String() != wantBody {
		t.Fatalf("list body=%s want=%s", listResponse.Body.String(), wantBody)
	}
}

// registerTestComponent fails at the fixture boundary so every scenario can assume its manager graph is valid.
// Keeping registration failures out of action assertions makes the behavioral phase of each test easier to follow.
func registerTestComponent(
	t *testing.T,
	manager *host.Manager,
	id string,
	dependencies []string,
	component *testComponent,
) {
	t.Helper()

	err := manager.Register(host.ManagedDefinition{
		ID:        id,
		DependsOn: dependencies,
		New:       component.newInstance,
	})
	if err != nil {
		t.Fatalf("register component %q: %v", id, err)
	}
}

// serveRequest exercises the actual mux with a fresh recorder, preserving route matching and path-value extraction in
// tests instead of calling endpoint methods directly.
func serveRequest(handler http.Handler, method, target string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	return response
}

// requireResponseCode centralizes the status assertion while retaining the response body in failures, where manager
// diagnostics are most useful.
func requireResponseCode(t *testing.T, response *httptest.ResponseRecorder, want int) {
	t.Helper()

	if response.Code != want {
		t.Fatalf("response code=%d want=%d body=%s", response.Code, want, response.Body.String())
	}
}
