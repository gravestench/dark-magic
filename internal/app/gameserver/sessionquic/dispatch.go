package sessionquic

import (
	"context"
)

// dispatch validates the tagged union once, then routes to the operation that owns its remaining bounds.
func (server *Server) dispatch(ctx context.Context, message request) response {
	if !validShape(message) {
		return response{Error: ErrWire.Error()}
	}

	switch message.Operation {
	case operationJoin:
		return server.dispatchJoin(ctx, message)
	case operationSubmit:
		return server.dispatchSubmit(message)
	case operationRefresh:
		return server.dispatchRefresh(message)
	case operationReconnect:
		return server.dispatchReconnect(message)
	case operationLeave:
		return server.dispatchLeave(message)
	case operationProfileAdmit:
		return server.dispatchProfileAdmission(ctx, message)
	case operationRecipe:
		return server.dispatchRecipe()
	case operationPackageChunk:
		return server.dispatchPackageChunk(ctx, message)
	default:
		return response{Error: ErrWire.Error()}
	}
}

// dispatchJoin bounds the opaque admission ticket before authentication and authoritative admission.
func (server *Server) dispatchJoin(ctx context.Context, message request) response {
	if message.Join == nil || len(message.Join.Credential) > MaxCredentialBytes {
		return response{Error: ErrWire.Error()}
	}

	joined, err := server.endpoint.Join(ctx, *message.Join)

	return joinResponse(joined, err)
}

// dispatchSubmit bounds command JSON before it reaches gameplay validation or the deterministic command log.
func (server *Server) dispatchSubmit(message request) response {
	if message.Command == nil ||
		len(message.Credential) > MaxCredentialBytes ||
		len(message.Command.Payload) > MaxCommandPayloadBytes {
		return response{Error: ErrWire.Error()}
	}

	return errorResponse(server.endpoint.Submit(message.Credential, *message.Command))
}

// dispatchRefresh performs a client-paced correction request after bounding the bearer credential.
func (server *Server) dispatchRefresh(message request) response {
	if len(message.Credential) > MaxCredentialBytes {
		return response{Error: ErrWire.Error()}
	}

	snapshot, err := server.endpoint.Refresh(message.Credential)
	if err != nil {
		return response{Error: err.Error()}
	}

	return response{Snapshot: &snapshot}
}

// dispatchReconnect keeps credential rotation inside Endpoint after enforcing the transport's outer size bound.
func (server *Server) dispatchReconnect(message request) response {
	if message.Reconnect == nil || len(message.Reconnect.Credential) > MaxCredentialBytes {
		return response{Error: ErrWire.Error()}
	}

	joined, err := server.endpoint.Reconnect(*message.Reconnect)

	return joinResponse(joined, err)
}

// dispatchLeave revokes an authenticated membership while preserving Endpoint's error semantics.
func (server *Server) dispatchLeave(message request) response {
	if len(message.Credential) > MaxCredentialBytes {
		return response{Error: ErrWire.Error()}
	}

	return errorResponse(server.endpoint.Leave(message.Credential))
}

// dispatchProfileAdmission permits profile offers only when the host explicitly installed self-hosted admissions.
func (server *Server) dispatchProfileAdmission(ctx context.Context, message request) response {
	if server.profiles == nil ||
		len(message.Credential) > MaxCredentialBytes ||
		len(message.Offer) == 0 ||
		len(message.Offer) > MaxProfileOfferBytes {
		return response{Error: ErrWire.Error()}
	}

	ticket, err := server.profiles.Admit(ctx, message.Credential.String(), message.Offer)
	if err != nil {
		return response{Error: err.Error()}
	}

	return response{Ticket: ticket}
}

// dispatchRecipe validates provider output so a broken host cannot advertise an unusable runtime contract.
func (server *Server) dispatchRecipe() response {
	if server.packages == nil {
		return response{Error: "game session QUIC: package distribution is unavailable"}
	}

	recipe := server.packages.Recipe()
	if err := recipe.Validate(); err != nil {
		return response{Error: err.Error()}
	}

	return response{Recipe: &recipe}
}

// dispatchPackageChunk delegates an already shape-checked bounded range to the immutable package provider.
func (server *Server) dispatchPackageChunk(ctx context.Context, message request) response {
	if server.packages == nil || message.Package == nil || message.Package.Limit > MaxPackageChunkBytes {
		return response{Error: ErrWire.Error()}
	}

	chunk, err := server.packages.ReadChunk(ctx, *message.Package)
	if err != nil {
		return response{Error: err.Error()}
	}

	return response{Package: &chunk}
}
