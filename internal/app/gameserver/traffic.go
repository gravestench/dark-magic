package gameserver

import (
	"bytes"
	"errors"
	"time"

	gamesession "github.com/gravestench/dark-magic/internal/game/session"
	"github.com/gravestench/dark-magic/internal/game/simulation"
)

// Submit binds client-controlled intent to authenticated player authority and tolerates only exact retransmission.
func (endpoint *Endpoint) Submit(credential SessionCredential, intent CommandIntent) error {
	member, err := endpoint.consume(credential, false)
	if err != nil {
		return err
	}

	target := intent.TargetTick
	if target == 0 {
		target = intent.ObservedServerTick + 2
	}

	command := simulation.Command{
		Tick:      target,
		Player:    member.principal.PlayerID,
		Authority: simulation.AuthorityPlayer,
		Sequence:  intent.Sequence,
		Kind:      intent.Kind,
		Payload:   intent.Payload,
	}

	_, err = endpoint.host.Session.SubmitNetwork(command)
	if exactCommandRetransmission(endpoint.host.Session, command, err) {
		return nil
	}

	return err
}

// exactCommandRetransmission makes retry idempotent without allowing one sequence to rewrite accepted input.
func exactCommandRetransmission(session *gamesession.Session, command simulation.Command, submitErr error) bool {
	if !errors.Is(submitErr, gamesession.ErrCommandSequence) {
		return false
	}

	accepted, found := session.AcceptedNetworkCommand(command.Player, command.Sequence)

	return found && accepted.Tick == command.Tick && accepted.Kind == command.Kind &&
		bytes.Equal(accepted.Payload, command.Payload)
}

// Refresh returns a canonical correction while charging the membership's client-paced refresh budget.
func (endpoint *Endpoint) Refresh(credential SessionCredential) (Snapshot, error) {
	member, err := endpoint.consume(credential, true)
	if err != nil {
		return Snapshot{}, err
	}

	return endpoint.snapshot(member.principal.PlayerID)
}

// Observe returns a correction for a server-paced watch without double-charging the unary refresh bucket.
func (endpoint *Endpoint) Observe(credential SessionCredential) (Snapshot, error) {
	member, err := endpoint.connection(credential)
	if err != nil {
		return Snapshot{}, err
	}

	return endpoint.snapshot(member.principal.PlayerID)
}

// BeginWatch reserves the membership's single server-paced correction stream.
func (endpoint *Endpoint) BeginWatch(credential SessionCredential) error {
	endpoint.mu.Lock()
	defer endpoint.mu.Unlock()

	key := string(credential)
	if member, found := endpoint.connections[key]; !found || credential == "" || !member.connected {
		return ErrAuthentication
	}

	if endpoint.watches[key] {
		return ErrRateLimit
	}

	endpoint.watches[key] = true

	return nil
}

// EndWatch releases stream ownership and is intentionally idempotent for transport cleanup paths.
func (endpoint *Endpoint) EndWatch(credential SessionCredential) {
	endpoint.mu.Lock()
	defer endpoint.mu.Unlock()

	delete(endpoint.watches, string(credential))
}

// consume authenticates a connected membership and atomically debits the selected request budget.
func (endpoint *Endpoint) consume(credential SessionCredential, refresh bool) (connection, error) {
	endpoint.mu.Lock()
	defer endpoint.mu.Unlock()

	member, found := endpoint.connections[string(credential)]
	if !found || credential == "" || !member.connected {
		return connection{}, ErrAuthentication
	}

	bucket := &member.commands
	if refresh {
		bucket = &member.refreshes
	}

	if !bucket.take(endpoint.now()) {
		return connection{}, ErrRateLimit
	}

	endpoint.connections[string(credential)] = member

	return member, nil
}

// newTokenBucket starts each membership at full burst capacity instead of imposing an admission delay.
func newTokenBucket(capacity, rate float64, now time.Time) tokenBucket {
	return tokenBucket{tokens: capacity, capacity: capacity, rate: rate, updated: now}
}

// take refills against monotonic endpoint time and refuses backward clocks from manufacturing capacity.
func (bucket *tokenBucket) take(now time.Time) bool {
	if now.Before(bucket.updated) {
		now = bucket.updated
	}

	bucket.tokens = min(bucket.capacity, bucket.tokens+now.Sub(bucket.updated).Seconds()*bucket.rate)

	bucket.updated = now
	if bucket.tokens < 1 {
		return false
	}

	bucket.tokens--

	return true
}
