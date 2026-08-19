package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/gameserver/sessionquic"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/app/serverapp"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

// startQUICTransport creates the optional authenticated game transport.
func startQUICTransport(
	host *gameserver.Host,
	prepared *serverContent,
	admissions admissionState,
	config serverConfig,
) (*sessionquic.Server, error) {
	server, err := serverapp.StartQUIC(serverapp.QUICConfig{
		Address:          config.quicListen,
		CertificatePath:  config.tlsCertificate,
		PrivateKeyPath:   config.tlsKey,
		AdmissionKeyPath: config.admissionKey,
		SessionID:        config.sessionID,
		RemoteProfile:    admissions.remoteProfile,
		ModCache:         prepared.mods.Cache,
		Tickets:          admissions.workerTickets,
		RealmMemberships: admissions.workerMemberships,
	}, host)
	if err != nil {
		return nil, fmt.Errorf("start QUIC game-session transport: %w", err)
	}

	if server != nil {
		slog.Info("serving authenticated game sessions", "address", server.Addr())
	}

	return server, nil
}

// startWorkerControl creates the private Realm supervision endpoint.
func startWorkerControl(
	host *gameserver.Host,
	quicServer *sessionquic.Server,
	destination playeradapter.Destination,
	admissions admissionState,
	config serverConfig,
) (*serverapp.WorkerControlServer, <-chan struct{}, error) {
	if !config.workerConfigured() {
		return nil, nil, nil
	}

	drain := make(chan struct{}, 1)

	control, err := serverapp.StartWorkerControl(serverapp.WorkerControlConfig{
		Address:         config.workerControlListen,
		CertificatePath: config.tlsCertificate,
		PrivateKeyPath:  config.tlsKey,
		TokenPath:       config.workerControlToken,
		Tickets:         admissions.workerTickets,
		Destination:     destination,
		Memberships:     admissions.workerMemberships,
		Drain:           drainWorker(drain, quicServer),
	}, host)
	if err != nil {
		return nil, nil, fmt.Errorf("start Realm worker control: %w", err)
	}

	return control, drain, nil
}

// drainWorker fences public traffic after notifying the worker coordinator.
func drainWorker(drain chan<- struct{}, quicServer *sessionquic.Server) func() {
	return func() {
		select {
		case drain <- struct{}{}:
		default:
		}

		_ = quicServer.Close()
	}
}

// publishWorkerReadiness writes the owner-only rendezvous consumed by the Realm.
func publishWorkerReadiness(
	control *serverapp.WorkerControlServer,
	quicServer *sessionquic.Server,
	config serverConfig,
) (string, error) {
	path, err := darkpaths.ExpandHost(config.workerReadyFile)
	if err != nil {
		return "", fmt.Errorf("expand Realm worker readiness path: %w", err)
	}

	ready := realm.WorkerProcessReady{
		GameID:         config.sessionID,
		AllocationID:   config.allocationID,
		ProcessID:      os.Getpid(),
		ControlAddress: control.Addr().String(),
		GameEndpoint: realm.GameEndpoint{
			Address:        quicServer.Addr(),
			TLSFingerprint: control.TLSFingerprint(),
		},
	}
	if err := realm.WriteWorkerProcessReady(path, ready); err != nil {
		return "", fmt.Errorf("write Realm worker readiness: %w", err)
	}

	slog.Info(
		"Realm worker ready",
		"control_address", control.Addr(),
		"game_address", quicServer.Addr(),
		"tls_fingerprint", control.TLSFingerprint(),
	)

	return path, nil
}

// closeWorkerControl bounds the private control server shutdown.
func closeWorkerControl(control *serverapp.WorkerControlServer) {
	if control != nil {
		_ = control.Close(context.Background())
	}
}
