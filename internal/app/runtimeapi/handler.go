package runtimeapi

import (
	"encoding/json"
	"net/http"

	"github.com/gravestench/dark-magic/internal/app/host"
)

// componentStatus is the public HTTP projection of host.Status. Converting explicitly prevents the manager's opaque
// error value from leaking while retaining its diagnostic message for local operators.
type componentStatus struct {
	ID      string     `json:"id"`
	Desired bool       `json:"desired"`
	State   host.State `json:"state"`
	Error   string     `json:"error,omitempty"`
}

// Handler exposes the versioned management surface for tests and embedding. A fresh mux keeps route ownership local
// to each caller and prevents registration on the process-wide default mux.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/components", s.listComponents)
	mux.HandleFunc("POST /v1/components/{id}/{action}", s.transitionComponent)

	return mux
}

// listComponents returns manager snapshots in registration order. Preserving that deterministic order makes the
// development API stable for both operators and automated diagnostics.
func (s *Server) listComponents(writer http.ResponseWriter, _ *http.Request) {
	entries := s.manager.Statuses()

	result := make([]componentStatus, 0, len(entries))
	for _, entry := range entries {
		result = append(result, statusForResponse(entry))
	}

	writeJSON(writer, http.StatusOK, result)
}

// statusForResponse copies a manager snapshot into its wire representation. Errors become strings because manager
// error identities are meaningful inside the process but cannot be preserved across JSON.
func statusForResponse(entry host.Status) componentStatus {
	status := componentStatus{
		ID:      entry.ID,
		Desired: entry.Desired,
		State:   entry.State,
	}
	if entry.Err != nil {
		status.Error = entry.Err.Error()
	}

	return status
}

// transitionComponent applies one synchronous lifecycle request and reports the resulting manager snapshot. Using the
// request context preserves client cancellation and host transition deadlines across the HTTP boundary.
func (s *Server) transitionComponent(writer http.ResponseWriter, request *http.Request) {
	id := request.PathValue("id")
	action := request.PathValue("action")

	recognized, err := s.applyComponentTransition(request, id, action)
	if !recognized {
		writeError(writer, http.StatusBadRequest, "unknown action")
		return
	}

	if err != nil {
		writeError(writer, http.StatusConflict, err.Error())
		return
	}

	entry, _ := s.manager.Status(id)
	writeJSON(writer, http.StatusOK, entry)
}

// applyComponentTransition maps the HTTP action vocabulary onto the manager without introducing a second lifecycle
// policy. Cascade is deliberately accepted only as the exact value "true" to preserve the established API contract.
func (s *Server) applyComponentTransition(request *http.Request, id, action string) (bool, error) {
	switch action {
	case "enable":
		return true, s.manager.Enable(request.Context(), id)
	case "disable":
		// Accepting broader boolean spellings would silently expand an existing wire contract.
		if request.URL.Query().Get("cascade") == "true" {
			return true, s.manager.DisableCascade(request.Context(), id)
		}

		return true, s.manager.Disable(request.Context(), id)
	case "restart":
		return true, s.manager.Restart(request.Context(), id)
	default:
		return false, nil
	}
}

// writeError keeps all error responses on the same one-field JSON contract, avoiding accidental exposure of internal
// error structure when additional endpoints are added.
func writeError(writer http.ResponseWriter, code int, message string) {
	writeJSON(writer, code, map[string]string{"error": message})
}

// writeJSON commits the status before encoding to preserve the existing response contract. Encoding failures are not
// recoverable after headers are written, so the HTTP adapter intentionally retains the established best-effort write.
func writeJSON(writer http.ResponseWriter, code int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(code)

	// Headers are already committed, so an encoding failure cannot be translated into a second HTTP response.
	_ = json.NewEncoder(writer).Encode(value)
}
