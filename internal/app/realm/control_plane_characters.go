package realm

import (
	"context"
	"strings"

	"github.com/google/uuid"
	d2save "github.com/gravestench/dark-magic/internal/mod/d2legacy/adapter/save"
)

const maximumRealmCharacters = 18

// CreateCharacterRequest contains only player-controlled character choices;
// the realm supplies ownership, identity, revision, and compatibility.
type CreateCharacterRequest struct {
	Name      string `json:"name"`
	Class     string `json:"class"`
	Expansion bool   `json:"expansion"`
	Hardcore  bool   `json:"hardcore"`
}

// ListCharacters returns only records owned by the authenticated account.
func (control *ControlPlane) ListCharacters(
	ctx context.Context,
	token string,
) ([]CharacterRecord, error) {
	principal, err := control.authorize(ctx, token)
	if err != nil {
		return nil, err
	}
	return control.characters.List(ctx, principal.accountID)
}

// CreateCharacter accepts only player choices. Identity, level, stats,
// appearance, revision, ownership, and compatibility are realm/d2legacy owned.
func (control *ControlPlane) CreateCharacter(
	ctx context.Context,
	token string,
	request CreateCharacterRequest,
) (record CharacterRecord, err error) {
	request.Expansion = true
	event := AuditEvent{
		Operation:     AuditCharacterCreate,
		CharacterName: strings.TrimSpace(request.Name),
	}
	defer func() {
		event.AccountID = firstNonEmpty(event.AccountID, record.AccountID)
		event.CharacterID = record.Character.ID
		event.CharacterName = firstNonEmpty(record.Character.Name, event.CharacterName)
		control.recordAudit(ctx, event, err)
	}()

	principal, err := control.authorize(ctx, token)
	if err != nil {
		return CharacterRecord{}, err
	}
	event.AccountID = principal.accountID
	event.AccountName = principal.name
	event.SessionID = principal.sessionID

	existing, err := control.characters.List(ctx, principal.accountID)
	if err != nil {
		return CharacterRecord{}, err
	}
	if len(existing) >= maximumRealmCharacters {
		return CharacterRecord{}, ErrCharacterLimit
	}

	wanted := strings.ToLower(strings.TrimSpace(request.Name))
	for _, record := range existing {
		if strings.ToLower(record.Character.Name) == wanted {
			return CharacterRecord{}, ErrCharacterExists
		}
	}

	character, err := d2save.NewCharacter(d2save.CharacterRequest{
		ID:        uuid.New().String(),
		Name:      request.Name,
		Class:     request.Class,
		Expansion: request.Expansion,
		Hardcore:  request.Hardcore,
	})
	if err != nil {
		return CharacterRecord{}, ErrCharacterInput
	}

	record = CharacterRecord{
		AccountID:     principal.accountID,
		Revision:      1,
		Character:     character,
		Compatibility: control.characterCompatibility,
	}
	if err := control.characters.Create(ctx, record); err != nil {
		return CharacterRecord{}, err
	}
	if err := control.accounts.SelectCharacter(ctx, token, character.ID); err != nil {
		return CharacterRecord{}, err
	}
	return cloneCharacterRecord(record), nil
}

// DeleteCharacter authorizes ownership before removing an idle character; the
// repository remains responsible for rejecting leased characters.
func (control *ControlPlane) DeleteCharacter(
	ctx context.Context,
	token string,
	characterID string,
) (err error) {
	event := AuditEvent{
		Operation:   AuditCharacterDelete,
		CharacterID: strings.TrimSpace(characterID),
	}
	defer func() { control.recordAudit(ctx, event, err) }()

	principal, err := control.authorize(ctx, token)
	if err != nil {
		return err
	}
	event.AccountID = principal.accountID
	event.AccountName = principal.name
	event.SessionID = principal.sessionID

	return control.characters.Delete(ctx, principal.accountID, strings.TrimSpace(characterID))
}

// SelectCharacter verifies account ownership before storing the selection on
// the authenticated session, preventing clients from claiming foreign records.
func (control *ControlPlane) SelectCharacter(
	ctx context.Context,
	token string,
	characterID string,
) (record CharacterRecord, err error) {
	event := AuditEvent{
		Operation:   AuditCharacterSelect,
		CharacterID: strings.TrimSpace(characterID),
	}
	defer func() {
		event.CharacterName = record.Character.Name
		control.recordAudit(ctx, event, err)
	}()

	principal, err := control.authorize(ctx, token)
	if err != nil {
		return CharacterRecord{}, err
	}
	event.AccountID = principal.accountID
	event.AccountName = principal.name
	event.SessionID = principal.sessionID

	record, err = control.characters.Get(ctx, principal.accountID, strings.TrimSpace(characterID))
	if err != nil {
		return CharacterRecord{}, err
	}
	if err := control.accounts.SelectCharacter(ctx, token, record.Character.ID); err != nil {
		return CharacterRecord{}, err
	}
	return record, nil
}

// SelectedCharacter resolves the session selection through the authenticated
// account so callers receive only an owned character record.
func (control *ControlPlane) SelectedCharacter(
	ctx context.Context,
	token string,
) (CharacterRecord, error) {
	principal, err := control.authorize(ctx, token)
	if err != nil {
		return CharacterRecord{}, err
	}
	characterID, err := control.accounts.SelectedCharacter(ctx, token)
	if err != nil {
		return CharacterRecord{}, err
	}
	return control.characters.Get(ctx, principal.accountID, characterID)
}

// presenceFromCharacter projects only public character fields and defensively
// copies appearance data before it enters shared channel state.
func presenceFromCharacter(record CharacterRecord) CharacterPresence {
	character := record.Character
	return CharacterPresence{
		CharacterID: character.ID,
		Name:        character.Name,
		Class:       character.Class,
		Level:       character.Level,
		Expansion:   character.Expansion,
		Hardcore:    character.Hardcore,
		Appearance: clonePresence(CharacterPresence{
			Appearance: character.Appearance,
		}).Appearance,
	}
}
