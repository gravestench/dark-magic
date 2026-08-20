package clientapp

import (
	"context"
	"fmt"

	"github.com/gravestench/dark-magic/internal/app/realm"
)

// Login establishes a new Realm session from player-supplied credentials.
// Account authorization never acts as an implicit game login, and passwords
// are never copied into Realm status.
func (controller *realmController) Login(name, password string) error {
	return controller.start("authenticating_account", func(ctx context.Context, client realmAPI) error {
		session, err := client.Authenticate(ctx, name, password)
		if err != nil {
			return fmt.Errorf("authenticate Realm account: %w", err)
		}

		controller.update(func(state *realmClientState) {
			state.Phase = "loading_characters"
		})

		characters, err := client.ListCharacters(ctx)
		if err != nil {
			return fmt.Errorf("list Realm characters: %w", err)
		}

		controller.update(func(state *realmClientState) {
			state.Account = session.Account
			state.Characters = characters
			state.Phase = "characters"
		})

		return nil
	})
}

// Signup creates an account but never converts successful registration into an authenticated
// session. Verification and explicit login remain distinct security and user-consent steps.
func (controller *realmController) Signup(name, email, password string) error {
	return controller.start("creating_account", func(ctx context.Context, client realmAPI) error {
		account, err := client.Signup(ctx, name, email, password)
		if err != nil {
			return err
		}

		controller.update(func(state *realmClientState) {
			state.Account = account
			state.Phase = "verification_required"
		})

		return nil
	})
}

// RecoverPassword initiates server-side recovery without placing email, tokens, or password
// material in presentation state; the frontend receives only a completion phase.
func (controller *realmController) RecoverPassword(email string) error {
	return controller.start("requesting_password_recovery", func(ctx context.Context, client realmAPI) error {
		if err := client.BeginPasswordRecovery(ctx, email); err != nil {
			return err
		}

		controller.update(func(state *realmClientState) {
			state.Phase = "recovery_sent"
		})

		return nil
	})
}

// Logout asks Realm to remove presence and invalidate the bearer session before clearing local
// account data. The configured endpoint and gateway survive so the login screen can reconnect.
func (controller *realmController) Logout() error {
	return controller.start("logging_out", func(ctx context.Context, client realmAPI) error {
		if err := client.Logout(ctx); err != nil {
			return err
		}

		controller.update(func(state *realmClientState) {
			gateway := state.Gateway
			endpoint := state.Endpoint
			*state = realmClientState{
				Phase:    "login",
				Gateway:  gateway,
				Endpoint: endpoint,
			}
		})

		return nil
	})
}

// CreateCharacter reloads the authoritative directory after creation instead of appending the
// response locally. This preserves server ordering and any normalized fields added during creation.
func (controller *realmController) CreateCharacter(name, class string, expansion, hardcore bool) error {
	request := realm.CreateCharacterRequest{
		Name:      name,
		Class:     class,
		Expansion: expansion,
		Hardcore:  hardcore,
	}

	return controller.start("creating_character", func(ctx context.Context, client realmAPI) error {
		record, err := client.CreateCharacter(ctx, request)
		if err != nil {
			return err
		}

		characters, err := client.ListCharacters(ctx)
		if err != nil {
			return err
		}

		controller.update(func(state *realmClientState) {
			state.Characters = characters
			state.Selected = record
			state.Phase = "characters"
		})

		return nil
	})
}

// DeleteCharacter clears selection and replaces the directory from authority, preventing a deleted
// character from remaining eligible for game admission in stale UI state.
func (controller *realmController) DeleteCharacter(id string) error {
	return controller.start("deleting_character", func(ctx context.Context, client realmAPI) error {
		if err := client.DeleteCharacter(ctx, id); err != nil {
			return err
		}

		characters, err := client.ListCharacters(ctx)
		if err != nil {
			return err
		}

		controller.update(func(state *realmClientState) {
			state.Characters = characters
			state.Selected = realm.CharacterSummary{}
			state.Phase = "characters"
		})

		return nil
	})
}

// SelectCharacter stores the server-returned summary rather than trusting the requested ID. Later
// game admission therefore uses canonical character identity and revision data.
func (controller *realmController) SelectCharacter(id string) error {
	return controller.start("selecting_character", func(ctx context.Context, client realmAPI) error {
		record, err := client.SelectCharacter(ctx, id)
		if err != nil {
			return err
		}

		controller.update(func(state *realmClientState) {
			state.Selected = record
			state.Phase = "character_selected"
		})

		return nil
	})
}
