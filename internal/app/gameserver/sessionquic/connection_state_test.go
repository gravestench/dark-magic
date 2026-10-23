package sessionquic

import (
	"testing"

	"github.com/gravestench/dark-magic/internal/app/gameserver"
)

func TestReconnectReplayTransfersSocketCleanup(t *testing.T) {
	for _, ordering := range []string{"close_before_replay", "close_after_replay", "replacement_dispatches_first"} {
		t.Run(ordering, func(t *testing.T) {
			fixture := newQUICProtocolFixture(t)
			original, replacement := newConnectionMemberships(), newConnectionMemberships()

			joined := fixture.server.dispatchConnection(fixture.ctx, request{
				Operation: operationJoin,
				Join: &gameserver.JoinRequest{Version: gameserver.SessionProtocolVersion,
					Credential: "realm-ticket", Identity: fixture.identity},
			}, original)
			if joined.Error != "" || joined.Join == nil {
				t.Fatalf("join = %#v", joined)
			}

			reconnect := request{Operation: operationReconnect, Reconnect: &gameserver.ReconnectRequest{
				Credential: joined.Join.Credential, Identity: fixture.identity,
				Nonce: "0123456789abcdef0123456789abcdef",
			}}

			first := original
			if ordering == "replacement_dispatches_first" {
				first = replacement
			}

			rotated := fixture.server.dispatchConnection(fixture.ctx, reconnect, first)
			if rotated.Error != "" || rotated.Join == nil {
				t.Fatalf("rotation = %#v", rotated)
			}

			if ordering == "close_before_replay" {
				fixture.server.disconnectMemberships(original)
			}

			if ordering == "replacement_dispatches_first" {
				late := fixture.server.dispatchConnection(fixture.ctx, reconnect, original)
				if late.Error != "" && late.Error != gameserver.ErrAuthentication.Error() {
					t.Fatalf("late original = %#v", late)
				}

				fixture.server.disconnectMemberships(original)

				if _, err := fixture.server.endpoint.Refresh(rotated.Join.Credential); err != nil {
					t.Fatalf("late original reclaimed replacement cleanup: %v", err)
				}
			}

			wrongNonce := *reconnect.Reconnect

			wrongNonce.Nonce = "fedcba9876543210fedcba9876543210"
			if result := fixture.server.dispatchConnection(fixture.ctx, request{
				Operation: operationReconnect, Reconnect: &wrongNonce,
			}, replacement); result.Error == "" {
				t.Fatal("wrong nonce resumed membership")
			}

			wrongRuntime := *reconnect.Reconnect

			wrongRuntime.Identity.Recipe.Packages.Base.Digest =
				"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
			if result := fixture.server.dispatchConnection(fixture.ctx, request{
				Operation: operationReconnect, Reconnect: &wrongRuntime,
			}, replacement); result.Error == "" {
				t.Fatal("wrong runtime resumed membership")
			}

			replayed := fixture.server.dispatchConnection(fixture.ctx, reconnect, replacement)
			if replayed.Error != "" || replayed.Join == nil || replayed.Join.Credential != rotated.Join.Credential {
				t.Fatalf("replay = %#v", replayed)
			}

			fixture.server.disconnectMemberships(original)

			if _, err := fixture.server.endpoint.Refresh(replayed.Join.Credential); err != nil {
				t.Fatalf("replacement lost authenticated membership: %v", err)
			}

			if _, err := fixture.server.endpoint.Refresh(joined.Join.Credential); err != gameserver.ErrAuthentication {
				t.Fatalf("old credential remains valid: %v", err)
			}

			if result := fixture.server.dispatchConnection(fixture.ctx, reconnect, original); result.Error == "" {
				t.Fatal("closed socket reclaimed cleanup")
			}

			fixture.server.disconnectMemberships(replacement)

			if _, err := fixture.server.endpoint.Refresh(replayed.Join.Credential); err != gameserver.ErrAuthentication {
				t.Fatalf("replacement closure did not suspend membership: %v", err)
			}

			if err := fixture.server.endpoint.Leave(replayed.Join.Credential); err != nil {
				t.Fatal(err)
			}

			result := fixture.server.dispatchConnection(fixture.ctx, reconnect, newConnectionMemberships())
			if result.Error == "" {
				t.Fatal("replay resurrected revoked membership")
			}
		})
	}
}
