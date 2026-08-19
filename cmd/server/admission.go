package main

import (
	"fmt"
	"time"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
	"github.com/gravestench/dark-magic/internal/app/realm"
	"github.com/gravestench/dark-magic/internal/app/serverapp"
	playeradapter "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/player"
	darkpaths "github.com/gravestench/dark-magic/internal/paths"
)

// admissionState collects local, remote, and Realm-worker admission policy.
type admissionState struct {
	localProfile      serverapp.ProfileAdmission
	remoteProfile     *serverapp.RemoteProfileConfig
	workerTickets     *gameserver.TicketAuthority
	workerMemberships *realm.WorkerMemberships
}

// prepareAdmissions configures every admission mode before transports start.
func prepareAdmissions(
	host *gameserver.Host,
	restoredPlayerIDs []string,
	config serverConfig,
) (admissionState, error) {
	localProfile, profilePath, err := prepareLocalProfile(host, config)
	if err != nil {
		return admissionState{}, err
	}

	remoteProfile, err := prepareRemoteProfile(profilePath, config)
	if err != nil {
		return admissionState{}, err
	}

	tickets, memberships, err := prepareWorkerAdmissions(restoredPlayerIDs, config)
	if err != nil {
		return admissionState{}, err
	}

	return admissionState{
		localProfile:      localProfile,
		remoteProfile:     remoteProfile,
		workerTickets:     tickets,
		workerMemberships: memberships,
	}, nil
}

// prepareLocalProfile admits the explicitly selected self-hosted profile.
func prepareLocalProfile(
	host *gameserver.Host,
	config serverConfig,
) (serverapp.ProfileAdmission, string, error) {
	path, err := darkpaths.ExpandHost(config.playerProfile)
	if err != nil {
		return serverapp.ProfileAdmission{}, "", fmt.Errorf("expand player profile path: %w", err)
	}

	if path == "" {
		return serverapp.ProfileAdmission{}, "", nil
	}

	destination, err := profileDestination(config)
	if err != nil {
		return serverapp.ProfileAdmission{}, "", err
	}

	admission := serverapp.ProfileAdmission{
		Path:        path,
		PlayerID:    config.profilePlayer,
		Destination: destination,
	}
	if err := serverapp.AdmitSelectedProfile(host, admission); err != nil {
		return serverapp.ProfileAdmission{}, "", fmt.Errorf("admit selected player profile: %w", err)
	}

	return admission, path, nil
}

// prepareRemoteProfile builds optional credential-backed self-host admission.
func prepareRemoteProfile(
	localProfilePath string,
	config serverConfig,
) (*serverapp.RemoteProfileConfig, error) {
	if config.remoteProfileKey == "" {
		return nil, nil
	}

	if localProfilePath != "" {
		return nil, fmt.Errorf("local and remote profile admission are mutually exclusive")
	}

	credential, err := serverapp.ReadAdmissionKey(config.remoteProfileKey)
	if err != nil {
		return nil, fmt.Errorf("read remote profile key: %w", err)
	}

	destination, err := profileDestination(config)
	if err != nil {
		return nil, err
	}

	return &serverapp.RemoteProfileConfig{
		Credential:  string(credential),
		PrincipalID: "self-host:remote-user",
		PlayerID:    config.profilePlayer,
		Destination: destination,
		Lifetime:    30 * time.Second,
	}, nil
}

// profileDestination validates the shared local and remote spawn configuration.
func profileDestination(config serverConfig) (playeradapter.Destination, error) {
	destination, err := playeradapter.NewDestination(
		config.profileX,
		config.profileY,
		config.profileWidth,
		config.profileHeight,
		config.profileAct,
		config.profileLevel,
	)
	if err != nil {
		return playeradapter.Destination{}, fmt.Errorf("validate player-profile destination: %w", err)
	}

	return destination, nil
}

// prepareWorkerAdmissions restores Realm ticket and membership authority.
func prepareWorkerAdmissions(
	restoredPlayerIDs []string,
	config serverConfig,
) (*gameserver.TicketAuthority, *realm.WorkerMemberships, error) {
	if !config.workerConfigured() {
		return nil, nil, nil
	}

	secret, err := serverapp.ReadAdmissionKey(config.admissionKey)
	if err != nil {
		return nil, nil, fmt.Errorf("read Realm worker admission key: %w", err)
	}

	tickets, err := gameserver.NewTicketAuthority(secret, config.sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("create Realm worker ticket authority: %w", err)
	}

	memberships := realm.NewWorkerMemberships()
	for _, playerID := range restoredPlayerIDs {
		memberships.Admit(playerID, time.Time{})
	}

	return tickets, memberships, nil
}
