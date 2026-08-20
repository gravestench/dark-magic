package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const defaultAcceptanceMode = "session"

// acceptanceConfig keeps credentials and scenario selection together without exposing them beyond this command.
type acceptanceConfig struct {
	endpoint      string
	accountName   string
	password      string
	characterName string
	mode          string
}

// loadAcceptanceConfig reads credentials in the command's established order so validation failures remain stable.
func loadAcceptanceConfig() (acceptanceConfig, error) {
	config := acceptanceConfig{
		endpoint:    strings.TrimSpace(os.Getenv("DARK_MAGIC_REALM_ACCEPTANCE_URL")),
		accountName: strings.TrimSpace(os.Getenv("DARK_MAGIC_REALM_ACCEPTANCE_ACCOUNT")),
	}

	password, err := acceptancePassword()
	if err != nil {
		return acceptanceConfig{}, err
	}

	config.password = password

	config.characterName = strings.TrimSpace(os.Getenv("DARK_MAGIC_REALM_ACCEPTANCE_CHARACTER"))

	config.mode = strings.TrimSpace(os.Getenv("DARK_MAGIC_REALM_ACCEPTANCE_MODE"))
	if config.mode == "" {
		config.mode = defaultAcceptanceMode
	}

	if err := config.validate(); err != nil {
		return acceptanceConfig{}, err
	}

	return config, nil
}

// validate rejects incomplete invocations before they can create or lease any Realm resources.
func (config acceptanceConfig) validate() error {
	if config.endpoint == "" || config.accountName == "" || config.password == "" || config.characterName == "" {
		return errors.New("URL, account, password, and character environment variables are required")
	}

	return nil
}

// acceptancePassword gives the file credential precedence so callers can avoid placing secrets in the environment.
func acceptancePassword() (string, error) {
	path := strings.TrimSpace(os.Getenv("DARK_MAGIC_REALM_ACCEPTANCE_PASSWORD_FILE"))
	if path != "" {
		value, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("read password file: %w", err)
		}

		password := strings.TrimRight(string(value), "\r\n")
		if password == "" {
			return "", errors.New("password file is empty")
		}

		return password, nil
	}

	password := os.Getenv("DARK_MAGIC_REALM_ACCEPTANCE_PASSWORD")
	if password == "" {
		return "", errors.New("password or password file is required")
	}

	return password, nil
}

// localHTTPSClient permits disposable certificate trust only when the endpoint is an explicit loopback IP.
func localHTTPSClient(endpoint string) (*http.Client, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Scheme != "https" {
		return nil, errors.New("acceptance requires an HTTPS Realm URL")
	}

	// Requiring a parsed IP prevents a hostname such as localhost from being redirected outside the machine.
	ip := net.ParseIP(parsed.Hostname())
	if ip == nil || !ip.IsLoopback() {
		return nil, errors.New("development acceptance permits certificate TOFU only on an explicit loopback IP")
	}

	// This disposable loopback test starts with an empty trust store. Worker identity remains strictly pinned from the
	// authenticated Realm handoff rather than inheriting this Realm-only exception.
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
	}

	return &http.Client{Transport: transport, Timeout: 15 * time.Second}, nil
}
