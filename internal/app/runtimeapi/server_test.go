package runtimeapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gravestench/dark-magic/internal/app/host"
)

type component struct{}

func (*component) Start(context.Context) error { return nil }
func (*component) Stop(context.Context) error  { return nil }

func TestHandlerControlsRegisteredComponents(t *testing.T) {
	manager := host.NewManager()
	if err := manager.Register(host.ManagedDefinition{ID: "scripts", New: func(context.Context) (host.Component, error) { return &component{}, nil }}); err != nil {
		t.Fatal(err)
	}
	server := New("", manager)
	request := httptest.NewRequest(http.MethodPost, "/v1/components/scripts/enable", nil)
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"state":"enabled"`) {
		t.Fatalf("code=%d body=%s", response.Code, response.Body.String())
	}
	request = httptest.NewRequest(http.MethodGet, "/v1/components", nil)
	response = httptest.NewRecorder()
	server.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"id":"scripts"`) {
		t.Fatalf("code=%d body=%s", response.Code, response.Body.String())
	}
}
