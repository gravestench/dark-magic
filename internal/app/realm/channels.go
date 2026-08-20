package realm

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

const (
	ChannelViewVersion    = "RealmChannel/v1"
	maximumChannelBytes   = 64
	maximumChatBytes      = 255
	defaultChannelHistory = 256
)

var (
	ErrChannelInput  = errors.New("realm: invalid channel input")
	ErrChannelMember = errors.New("realm: channel membership required")
)

type CharacterPresence struct {
	CharacterID string             `json:"character_id"`
	Name        string             `json:"name"`
	Class       string             `json:"class"`
	Level       int                `json:"level"`
	Title       string             `json:"title,omitempty"`
	Expansion   bool               `json:"expansion"`
	Hardcore    bool               `json:"hardcore"`
	Appearance  *d2save.Appearance `json:"appearance,omitempty"`
}

type ChannelMember struct {
	MemberID  string            `json:"member_id"`
	Account   string            `json:"account"`
	Character CharacterPresence `json:"character"`
	JoinedAt  time.Time         `json:"joined_at"`
	ActiveAt  time.Time         `json:"active_at"`
}

type ChatEventKind string

const (
	ChatEventMessage ChatEventKind = "message"
	ChatEventJoined  ChatEventKind = "member_joined"
	ChatEventLeft    ChatEventKind = "member_left"
)

type ChatEvent struct {
	Sequence  uint64         `json:"sequence"`
	Kind      ChatEventKind  `json:"kind"`
	ChannelID string         `json:"channel_id"`
	Sender    *ChannelMember `json:"sender,omitempty"`
	Text      string         `json:"text,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

type ChannelView struct {
	Version   string          `json:"version"`
	Revision  uint64          `json:"revision"`
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Members   []ChannelMember `json:"members"`
	LastEvent uint64          `json:"last_event"`
}

type channelState struct {
	id       string
	name     string
	revision uint64
	members  map[string]ChannelMember
	events   []ChatEvent
	sequence uint64
}

// Channels owns realm-scoped channel presence and public messages. Returned
// members/events are defensive projections suitable for a lobby capability.
type Channels struct {
	mu           sync.RWMutex
	now          func() time.Time
	historyLimit int
	channels     map[string]*channelState
	bySession    map[string]string
	byCharacter  map[string]string
}

// NewChannels constructs the channels boundary and validates dependencies before callers can publish or mutate shared
// state.
func NewChannels(historyLimit int) *Channels {
	if historyLimit <= 0 {
		historyLimit = defaultChannelHistory
	}

	return &Channels{now: time.Now, historyLimit: historyLimit, channels: make(map[string]*channelState),
		bySession: make(map[string]string), byCharacter: make(map[string]string)}
}

// Join coordinates join through the owning channels synchronization boundary so shared state is published only after a
// complete transition.
func (channels *Channels) Join(
	ctx context.Context,
	principal AuthenticatedPrincipal,
	name string,
	character CharacterPresence,
) (ChannelView, error) {
	if err := contextErr(ctx); err != nil {
		return ChannelView{}, err
	}

	if channels == nil || !principal.valid() || validatePresence(character) != nil {
		return ChannelView{}, ErrChannelInput
	}

	display, id, err := normalizeChannelName(name)
	if err != nil {
		return ChannelView{}, err
	}

	channels.mu.Lock()
	defer channels.mu.Unlock()

	if sessionID := channels.byCharacter[character.CharacterID]; sessionID != "" && sessionID != principal.sessionID {
		return ChannelView{}, ErrCharacterOnline
	}

	if previousID := channels.bySession[principal.sessionID]; previousID != "" && previousID != id {
		channels.leaveLocked(previousID, principal.sessionID)
	}

	channel := channels.channels[id]
	if channel == nil {
		channel = &channelState{id: id, name: display, members: make(map[string]ChannelMember)}
		channels.channels[id] = channel
	}

	member, exists := channel.members[principal.sessionID]
	if exists && member.Character.CharacterID != character.CharacterID {
		delete(channels.byCharacter, member.Character.CharacterID)
	}

	if !exists {
		now := channels.now().UTC()
		member = ChannelMember{MemberID: uuid.New().String(), Account: principal.name, JoinedAt: now, ActiveAt: now}
	} else {
		member.ActiveAt = channels.now().UTC()
	}

	member.Character = clonePresence(character)
	channel.members[principal.sessionID] = member
	channels.bySession[principal.sessionID] = id
	channels.byCharacter[character.CharacterID] = principal.sessionID

	channel.revision++
	if !exists {
		channels.appendEventLocked(channel, ChatEvent{Kind: ChatEventJoined, Sender: cloneMemberPointer(member)})
	}

	return channelView(channel), nil
}

// Leave coordinates leave through the owning channels synchronization boundary so shared state is published only after
// a complete transition.
func (channels *Channels) Leave(ctx context.Context, principal AuthenticatedPrincipal) error {
	if err := contextErr(ctx); err != nil {
		return err
	}

	if channels == nil || !principal.valid() {
		return ErrChannelMember
	}

	channels.mu.Lock()
	defer channels.mu.Unlock()

	id := channels.bySession[principal.sessionID]
	if id == "" || !channels.leaveLocked(id, principal.sessionID) {
		return ErrChannelMember
	}

	return nil
}

// View coordinates view through the owning channels synchronization boundary so shared state is published only after a
// complete transition.
func (channels *Channels) View(ctx context.Context, principal AuthenticatedPrincipal) (ChannelView, error) {
	if err := contextErr(ctx); err != nil {
		return ChannelView{}, err
	}

	if channels == nil || !principal.valid() {
		return ChannelView{}, ErrChannelMember
	}

	channels.mu.Lock()
	defer channels.mu.Unlock()

	channel := channels.channels[channels.bySession[principal.sessionID]]
	if channel == nil {
		return ChannelView{}, ErrChannelMember
	}

	member, found := channel.members[principal.sessionID]
	if !found {
		return ChannelView{}, ErrChannelMember
	}

	member.ActiveAt = channels.now().UTC()
	channel.members[principal.sessionID] = member

	return channelView(channel), nil
}

// Send coordinates send through the owning channels synchronization boundary so shared state is published only after a
// complete transition.
func (channels *Channels) Send(ctx context.Context, principal AuthenticatedPrincipal, text string) (ChatEvent, error) {
	text = strings.TrimSpace(text)

	if err := contextErr(ctx); err != nil {
		return ChatEvent{}, err
	}

	if channels == nil || !principal.valid() || text == "" || len(text) > maximumChatBytes || !utf8.ValidString(text) {
		return ChatEvent{}, ErrChannelInput
	}

	channels.mu.Lock()
	defer channels.mu.Unlock()

	channel := channels.channels[channels.bySession[principal.sessionID]]
	if channel == nil {
		return ChatEvent{}, ErrChannelMember
	}

	member, found := channel.members[principal.sessionID]
	if !found {
		return ChatEvent{}, ErrChannelMember
	}

	member.ActiveAt = channels.now().UTC()
	channel.members[principal.sessionID] = member
	event := ChatEvent{Kind: ChatEventMessage, Sender: cloneMemberPointer(member), Text: text}
	channels.appendEventLocked(channel, event)

	return cloneChatEvent(channel.events[len(channel.events)-1]), nil
}

// EventsAfter coordinates events after through the owning channels synchronization boundary so shared state is
// published only after a complete transition.
func (channels *Channels) EventsAfter(
	ctx context.Context,
	principal AuthenticatedPrincipal,
	sequence uint64,
	limit int,
) ([]ChatEvent, error) {
	if err := contextErr(ctx); err != nil {
		return nil, err
	}

	if channels == nil || !principal.valid() || limit < 0 {
		return nil, ErrChannelInput
	}

	channels.mu.Lock()
	defer channels.mu.Unlock()

	channel := channels.channels[channels.bySession[principal.sessionID]]
	if channel == nil {
		return nil, ErrChannelMember
	}

	member, found := channel.members[principal.sessionID]
	if !found {
		return nil, ErrChannelMember
	}

	member.ActiveAt = channels.now().UTC()
	channel.members[principal.sessionID] = member

	if limit == 0 || limit > channels.historyLimit {
		limit = channels.historyLimit
	}

	result := make([]ChatEvent, 0, limit)

	for _, event := range channel.events {
		if event.Sequence > sequence {
			result = append(result, cloneChatEvent(event))
			if len(result) == limit {
				break
			}
		}
	}

	return result, nil
}

// PruneInactive removes channel presence whose client has stopped renewing it.
// It does not invalidate the account session: a resumed client can rejoin the
// channel without re-entering credentials while the login itself remains valid.
func (channels *Channels) PruneInactive(ctx context.Context, cutoff time.Time) (int, error) {
	if err := contextErr(ctx); err != nil {
		return 0, err
	}

	if channels == nil || cutoff.IsZero() {
		return 0, ErrChannelInput
	}

	channels.mu.Lock()
	defer channels.mu.Unlock()

	pruned := 0

	for sessionID, channelID := range channels.bySession {
		channel := channels.channels[channelID]

		var member ChannelMember

		found := false
		if channel != nil {
			member, found = channel.members[sessionID]
		}

		if found && member.ActiveAt.After(cutoff) {
			continue
		}

		if channels.leaveLocked(channelID, sessionID) {
			pruned++
		}
	}

	return pruned, nil
}

// leaveLocked contains leave locked within the channels boundary so callers do not duplicate its domain-specific
// policy.
func (channels *Channels) leaveLocked(channelID, sessionID string) bool {
	channel := channels.channels[channelID]
	if channel == nil {
		delete(channels.bySession, sessionID)
		return false
	}

	member, found := channel.members[sessionID]
	if !found {
		delete(channels.bySession, sessionID)
		return false
	}

	delete(channel.members, sessionID)
	delete(channels.bySession, sessionID)

	if channels.byCharacter[member.Character.CharacterID] == sessionID {
		delete(channels.byCharacter, member.Character.CharacterID)
	}

	channel.revision++
	channels.appendEventLocked(channel, ChatEvent{Kind: ChatEventLeft, Sender: cloneMemberPointer(member)})

	return true
}

// appendEventLocked contains append event locked within the channels boundary so callers do not duplicate its
// domain-specific policy.
func (channels *Channels) appendEventLocked(channel *channelState, event ChatEvent) {
	channel.sequence++
	event.Sequence, event.ChannelID, event.CreatedAt = channel.sequence, channel.id, channels.now().UTC()

	channel.events = append(channel.events, cloneChatEvent(event))
	if excess := len(channel.events) - channels.historyLimit; excess > 0 {
		copy(channel.events, channel.events[excess:])
		channel.events = channel.events[:channels.historyLimit]
	}
}

// channelView contains channel view within the channels boundary so callers do not duplicate its domain-specific
// policy.
func channelView(channel *channelState) ChannelView {
	view := ChannelView{
		Version:   ChannelViewVersion,
		Revision:  channel.revision,
		ID:        channel.id,
		Name:      channel.name,
		LastEvent: channel.sequence,
		Members:   make([]ChannelMember, 0, len(channel.members)),
	}
	for _, member := range channel.members {
		view.Members = append(view.Members, cloneMember(member))
	}

	sort.Slice(view.Members, func(i, j int) bool {
		if view.Members[i].JoinedAt.Equal(view.Members[j].JoinedAt) {
			return view.Members[i].MemberID < view.Members[j].MemberID
		}

		return view.Members[i].JoinedAt.Before(view.Members[j].JoinedAt)
	})

	return view
}

// normalizeChannelName checks the channels invariant before state changes, keeping invalid values off shared paths.
func normalizeChannelName(name string) (string, string, error) {
	display := strings.Join(strings.Fields(name), " ")
	if display == "" || len(display) > maximumChannelBytes || !utf8.ValidString(display) {
		return "", "", ErrChannelInput
	}

	for _, value := range display {
		if value < 0x20 || value == 0x7f {
			return "", "", ErrChannelInput
		}
	}

	return display, strings.ToLower(display), nil
}

// validatePresence checks the channels invariant before state changes, keeping invalid values off shared paths.
func validatePresence(presence CharacterPresence) error {
	if strings.TrimSpace(presence.CharacterID) == "" || strings.TrimSpace(presence.Name) == "" ||
		strings.TrimSpace(presence.Class) == "" ||
		presence.Level < 1 {
		return ErrChannelInput
	}

	return nil
}

// clonePresence returns an independent channels value so callers cannot mutate repository-owned state through a
// returned record.
func clonePresence(presence CharacterPresence) CharacterPresence {
	if presence.Appearance == nil {
		return presence
	}

	appearance := *presence.Appearance

	appearance.Components = make(map[string]string, len(presence.Appearance.Components))
	for key, value := range presence.Appearance.Components {
		appearance.Components[key] = value
	}

	presence.Appearance = &appearance

	return presence
}

// cloneMember returns an independent channels value so callers cannot mutate repository-owned state through a returned
// record.
func cloneMember(member ChannelMember) ChannelMember {
	member.Character = clonePresence(member.Character)
	return member
}

// cloneMemberPointer returns an independent channels value so callers cannot mutate repository-owned state through a
// returned record.
func cloneMemberPointer(member ChannelMember) *ChannelMember {
	cloned := cloneMember(member)
	return &cloned
}

// cloneChatEvent returns an independent channels value so callers cannot mutate repository-owned state through a
// returned record.
func cloneChatEvent(event ChatEvent) ChatEvent {
	if event.Sender != nil {
		event.Sender = cloneMemberPointer(*event.Sender)
	}

	return event
}
