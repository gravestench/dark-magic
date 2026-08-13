// Package serverapp owns standalone game-server process composition below the
// deliberately thin cmd/server entry point.
package serverapp

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/gameserver/sessionquic"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
)

type QUICConfig struct {
	Address, CertificatePath, PrivateKeyPath, AdmissionKeyPath, SessionID string
	RemoteProfile                                                         *RemoteProfileConfig
}

func StartQUIC(config QUICConfig, host *gameserver.Host) (*sessionquic.Server, error) {
	configured := config.Address != "" || config.CertificatePath != "" || config.PrivateKeyPath != "" || config.AdmissionKeyPath != ""
	if !configured {
		return nil, nil
	}
	if config.Address == "" || config.CertificatePath == "" || config.PrivateKeyPath == "" || config.AdmissionKeyPath == "" {
		return nil, errors.New("server: quic-listen, tls-cert, tls-key, and admission-key must be set together")
	}
	certificate, err := tls.LoadX509KeyPair(config.CertificatePath, config.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("server: load QUIC certificate: %w", err)
	}
	secret, err := ReadAdmissionKey(config.AdmissionKeyPath)
	if err != nil {
		return nil, err
	}
	authenticator, err := gameserver.NewTicketAuthority(secret, config.SessionID)
	if err != nil {
		return nil, err
	}
	endpoint, err := gameserver.NewEndpoint(host, authenticator, playeradapter.ProjectClientView)
	if err != nil {
		return nil, err
	}
	server, err := sessionquic.Listen(config.Address, &tls.Config{Certificates: []tls.Certificate{certificate}}, endpoint)
	if err != nil {
		return nil, err
	}
	if config.RemoteProfile != nil {
		profiles, err := NewRemoteProfileAdmissions(host, authenticator, *config.RemoteProfile)
		if err != nil {
			_ = server.Close()
			return nil, err
		}
		server.SetProfileAdmissions(profiles)
	}
	return server, nil
}

func ReadAdmissionKey(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("server: open admission key: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("server: stat admission key: %w", err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		return nil, errors.New("server: admission key must not be accessible by group or others")
	}
	data, err := io.ReadAll(io.LimitReader(file, 4097))
	if err != nil {
		return nil, fmt.Errorf("server: read admission key: %w", err)
	}
	if len(data) > 4096 {
		return nil, errors.New("server: admission key exceeds 4096 bytes")
	}
	if len(data) > 0 && data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}
	if len(data) < 32 {
		return nil, errors.New("server: admission key must contain at least 32 bytes")
	}
	return data, nil
}
