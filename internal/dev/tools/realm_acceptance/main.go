// Command realm_acceptance exercises the native Realm-to-worker session path.
// It is intentionally a development acceptance tool, not a player client.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gravestench/dark-magic/internal/app/realm"
)

const acceptanceTimeout = 90 * time.Second

// main preserves the command's single-line failure contract for shell-driven acceptance runs.
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "realm acceptance:", err)
		os.Exit(1)
	}
}

// run gives configuration, Realm calls, and worker recovery one shared deadline so a stalled phase cannot hang CI.
func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), acceptanceTimeout)
	defer cancel()

	config, err := loadAcceptanceConfig()
	if err != nil {
		return err
	}

	client, session, err := authenticateRealm(ctx, config)
	if err != nil {
		return err
	}

	// Authentication deliberately precedes mode dispatch so every mode proves the same Realm login path.
	if config.mode == "verify" {
		return verifyPersisted(ctx, client, session, config.characterName)
	}

	if config.mode != "session" {
		return fmt.Errorf("unsupported mode %q", config.mode)
	}

	return runSessionAcceptance(ctx, client, session, config)
}

// authenticateRealm constructs the loopback-only client and installs the bearer token used by all later calls.
func authenticateRealm(
	ctx context.Context,
	config acceptanceConfig,
) (*realm.RealmClient, realm.RealmSession, error) {
	httpClient, err := localHTTPSClient(config.endpoint)
	if err != nil {
		return nil, realm.RealmSession{}, err
	}

	client, err := realm.NewRealmClient(config.endpoint, httpClient)
	if err != nil {
		return nil, realm.RealmSession{}, err
	}

	session, err := client.Authenticate(ctx, config.accountName, config.password)
	if err != nil {
		return nil, realm.RealmSession{}, fmt.Errorf("authenticate: %w", err)
	}

	return client, session, nil
}
