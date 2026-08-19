package realm

import (
	"context"
	"strings"
	"time"
)

// JoinChannel publishes the selected character as the session's lobby
// presence. Requiring a realm-owned selection prevents arbitrary roster data.
func (control *ControlPlane) JoinChannel(
	ctx context.Context,
	token string,
	channel string,
) (view ChannelView, err error) {
	event := AuditEvent{
		Operation: AuditChannelJoin,
		Channel:   strings.TrimSpace(channel),
	}
	defer func() {
		event.Channel = firstNonEmpty(view.Name, event.Channel)
		control.recordAudit(ctx, event, err)
	}()

	principal, err := control.authorize(ctx, token)
	if err != nil {
		return ChannelView{}, err
	}
	event.AccountID = principal.accountID
	event.AccountName = principal.name
	event.SessionID = principal.sessionID

	characterID, err := control.accounts.SelectedCharacter(ctx, token)
	if err != nil {
		return ChannelView{}, err
	}
	record, err := control.characters.Get(ctx, principal.accountID, characterID)
	if err != nil {
		return ChannelView{}, err
	}
	event.CharacterID = record.Character.ID
	event.CharacterName = record.Character.Name

	return control.channels.Join(ctx, principal, channel, presenceFromCharacter(record))
}

// ChannelView returns the caller's current channel projection after
// reauthorizing the session and pruning expired presence.
func (control *ControlPlane) ChannelView(
	ctx context.Context,
	token string,
) (ChannelView, error) {
	principal, err := control.authorize(ctx, token)
	if err != nil {
		return ChannelView{}, err
	}
	return control.channels.View(ctx, principal)
}

// SendChannelMessage attributes a public message to the authenticated session;
// audit data records message size but not chat contents.
func (control *ControlPlane) SendChannelMessage(
	ctx context.Context,
	token string,
	message string,
) (chatEvent ChatEvent, err error) {
	event := AuditEvent{
		Operation:    AuditChannelMessage,
		MessageBytes: len(message),
	}
	defer func() {
		event.Channel = chatEvent.ChannelID
		control.recordAudit(ctx, event, err)
	}()

	principal, err := control.authorize(ctx, token)
	if err != nil {
		return ChatEvent{}, err
	}
	event.AccountID = principal.accountID
	event.AccountName = principal.name
	event.SessionID = principal.sessionID

	return control.channels.Send(ctx, principal, message)
}

// ChannelEvents reads bounded history only for the caller's current channel,
// preserving the channel service's monotonic sequence contract.
func (control *ControlPlane) ChannelEvents(
	ctx context.Context,
	token string,
	after uint64,
	limit int,
) ([]ChatEvent, error) {
	principal, err := control.authorize(ctx, token)
	if err != nil {
		return nil, err
	}
	return control.channels.EventsAfter(ctx, principal, after, limit)
}

// PruneInactivePresence removes channel projections whose renewable heartbeat
// predates the configured cutoff; account sessions themselves remain intact.
func (control *ControlPlane) PruneInactivePresence(ctx context.Context) (int, error) {
	if control == nil || control.channels == nil || control.presenceTimeout <= 0 {
		return 0, ErrChannelInput
	}
	return control.channels.PruneInactive(ctx, time.Now().UTC().Add(-control.presenceTimeout))
}
