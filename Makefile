.PHONY: test test-lua test-lua-hardening test-lua-format test-lua-syntax test-network-hardening test-network-soak test-network-fuzz architecture test-race fmt vet build-client-backends d2legacy bik-view presentation-coverage skill-behavior-coverage skill-evidence profile profile-acceptance profile-render-backends profile-check realm-up realm-down realm-fresh-install realm-drain-game realm-mailpit-up realm-mailpit-down realm-test-production capture capture-all capture-game-world capture-game-world-movement capture-game-world-panels capture-blood-moor capture-act1-seam capture-monster-lab capture-missile-lab capture-combat-lab capture-warp-lab play-game-world play-monster-lab play-missile-lab play-combat-lab play-warp-lab

test:
	go test ./...

test-lua:
	DARK_MAGIC_LUA_TEST_TIERS=fast,integration go test ./internal/mod/d2legacy -run 'TestLua(Suites|HarnessContract)'

# Repetition and seeded order catch VM leakage and order dependence while
# preserving a reproducible failure command in CI logs.
test-lua-hardening: test-lua-format test-lua-syntax
	DARK_MAGIC_LUA_TEST_TIERS=fast,integration DARK_MAGIC_LUA_TEST_REPEAT=3 DARK_MAGIC_LUA_TEST_ORDER_SEED=8675309 go test -count=1 ./internal/mod/d2legacy -run 'TestLua(Suites|HarnessContract)'

test-lua-format:
	find internal/content/d2legacy/lua/d2legacy \( -name '*_test.lua' -o -path '*/tests/v1.lua' -o -path '*/tests/support/*.lua' \) -print0 | xargs -0 stylua --check

test-lua-syntax:
	find internal/content/d2legacy/lua -name '*.lua' -print0 | xargs -0 -n 100 luac -p

architecture:
	go test ./internal/acceptance -run 'Test(Retired|LegacyRenderer|NoAccidental|DependencyDirection|CommandRemains|GameplayOwnership)'

test-race:
	# Race instrumentation makes the Lua-heavy integration packages and their
	# t.Parallel cases contend for fixed startup deadlines. Keep complete package
	# and test coverage while making the gate deterministic on CI and developer
	# machines with constrained cores or memory.
	go test -race -p 1 -parallel 1 ./...

NETWORK_SOAK_TICKS ?= 80

test-network-hardening:
	DARK_MAGIC_NETWORK_SOAK_TICKS=$(NETWORK_SOAK_TICKS) go test -count=1 ./internal/app/gameserver/... ./internal/app/clientsession ./internal/app/networkclock ./internal/app/networktrust

test-network-soak:
	$(MAKE) --no-print-directory test-network-hardening NETWORK_SOAK_TICKS=1500

test-network-fuzz:
	go test ./internal/app/gameserver/sessionquic -run '^$$' -fuzz FuzzReadFrame -fuzztime 10s
	go test ./internal/app/gameserver/sessionquic -run '^$$' -fuzz FuzzDecodeTransformFrame -fuzztime 10s
	go test ./internal/app/clientsession -run '^$$' -fuzz FuzzDecodeClientView -fuzztime 10s

fmt:
	gofmt -w $$(find cmd internal pkg -name '*.go')

vet:
	go vet ./...

# Raylib remains the default binary. The explicit tags make local and CI
# backend selection visible and catch either adapter drifting out of contract.
build-client-backends:
	mkdir -p .local/bin
	go build -tags raylib -o .local/bin/dark-magic-raylib ./cmd/client
	go build -tags ebitengine -o .local/bin/dark-magic-ebitengine ./cmd/client

realm-up:
	./scripts/realm/up.sh

realm-down:
	./scripts/realm/down.sh

realm-fresh-install:
	./scripts/realm/fresh-install.sh

realm-drain-game:
	@test -n "$(GAME_ID)" || (echo 'GAME_ID is required' >&2; exit 2)
	./scripts/realm/drain-game.sh "$(GAME_ID)"

realm-mailpit-up:
	./scripts/realm/mailpit-up.sh

realm-mailpit-down:
	./scripts/realm/mailpit-down.sh

realm-test-production:
	./scripts/realm/test-production.sh

d2legacy:
	go run ./internal/dev/tools/d2legacy_pack -output ./dist/d2legacy.zip

bik-view:
	go run ./internal/dev/tools/bik_view -source "$${MPQ_DIRECTORY}" -asset "$${BIK_ASSET}"

presentation-coverage:
	go run ./internal/dev/tools/presentation_coverage

skill-behavior-coverage:
	go run ./internal/dev/tools/skill_behavior_coverage -mpq-dir "$${MPQ_DIRECTORY}"

SKILL_IDS ?= 0,36,40

skill-evidence:
	@go run ./internal/dev/tools/skill_evidence -mpq-dir "$${MPQ_DIRECTORY}" -skill-ids "$(SKILL_IDS)"

PROFILE_DIR ?= ./profiles/acceptance
PROFILE_BUDGETS ?= ./docs/profile-budgets.json
RENDER_PROFILE_DIR ?= ./profiles/render-backends
RENDER_PROFILE_SCENE ?= game_world
RENDER_PROFILE_SETTLE ?= 300

profile:
	go run -tags ffmpeg ./cmd/client --profile-dir "$(PROFILE_DIR)" --profile-scenes all

# Deterministic real-asset runs for every budgeted scene. game_loading is
# profiled while the capture waits for the game_world it naturally enters.
profile-acceptance:
	go run -tags ffmpeg ./cmd/client --profile-dir "$(PROFILE_DIR)" --profile-scenes loading --capture-dir "$(PROFILE_DIR)/captures/loading" --capture-scenes loading --capture-settle-frames 60 --start-scene loading --fixture-characters 3
	go run -tags ffmpeg ./cmd/client --profile-dir "$(PROFILE_DIR)" --profile-scenes title --capture-dir "$(PROFILE_DIR)/captures/title" --capture-scenes title --capture-settle-frames 60 --start-scene title --fixture-characters 3
	go run -tags ffmpeg ./cmd/client --profile-dir "$(PROFILE_DIR)" --profile-scenes main_menu --capture-dir "$(PROFILE_DIR)/captures/main_menu" --capture-scenes main_menu --capture-settle-frames 60 --start-scene main_menu --fixture-characters 3
	go run -tags ffmpeg ./cmd/client --profile-dir "$(PROFILE_DIR)" --profile-scenes character_select --capture-dir "$(PROFILE_DIR)/captures/character_select" --capture-scenes character_select --capture-settle-frames 60 --start-scene character_select --fixture-characters 3
	go run -tags ffmpeg ./cmd/client --profile-dir "$(PROFILE_DIR)" --profile-scenes character_create --capture-dir "$(PROFILE_DIR)/captures/character_create" --capture-scenes character_create --capture-settle-frames 60 --start-scene character_create --fixture-characters 3
	go run -tags ffmpeg ./cmd/client --profile-dir "$(PROFILE_DIR)" --profile-scenes game_loading --capture-dir "$(PROFILE_DIR)/captures/game_loading" --capture-scenes game_world --capture-settle-frames 120 --start-scene game_loading --fixture-characters 3 --fixture-world-level 2
	go run -tags ffmpeg ./cmd/client --profile-dir "$(PROFILE_DIR)" --profile-scenes game_world --capture-dir "$(PROFILE_DIR)/captures/game_world" --capture-scenes game_world --capture-settle-frames 180 --start-scene game_world --fixture-characters 1 --fixture-world-level 2 --fixture-pointer-move=true

# Build once, then run the same deterministic workload through each native
# renderer. Audio is muted on both paths so backend-specific sound work cannot
# contaminate the graphics comparison.
profile-render-backends: build-client-backends
	MPQ_DIRECTORY="$${MPQ_DIRECTORY}" .local/bin/dark-magic-raylib --native-audio=false --profile-dir "$(RENDER_PROFILE_DIR)/raylib" --profile-scenes "$(RENDER_PROFILE_SCENE)" --capture-dir "$(RENDER_PROFILE_DIR)/raylib/captures" --capture-scenes "$(RENDER_PROFILE_SCENE)" --capture-settle-frames "$(RENDER_PROFILE_SETTLE)" --start-scene "$(RENDER_PROFILE_SCENE)" --fixture-characters 1 --fixture-world-level 2 --fixture-pointer-move=true
	MPQ_DIRECTORY="$${MPQ_DIRECTORY}" .local/bin/dark-magic-ebitengine --native-audio=false --profile-dir "$(RENDER_PROFILE_DIR)/ebitengine" --profile-scenes "$(RENDER_PROFILE_SCENE)" --capture-dir "$(RENDER_PROFILE_DIR)/ebitengine/captures" --capture-scenes "$(RENDER_PROFILE_SCENE)" --capture-settle-frames "$(RENDER_PROFILE_SETTLE)" --start-scene "$(RENDER_PROFILE_SCENE)" --fixture-characters 1 --fixture-world-level 2 --fixture-pointer-move=true
	go run ./internal/dev/tools/render_backend_compare -profile-dir "$(RENDER_PROFILE_DIR)" -scene "$(RENDER_PROFILE_SCENE)"

profile-check:
	go run ./internal/dev/tools/profile_check -profile-dir "$(PROFILE_DIR)" -budgets "$(PROFILE_BUDGETS)"

CAPTURE_DIR ?= ./captures/frontend
CAPTURE_SCENES ?= loading,title
START_SCENE ?=
FIXTURE_CHARACTERS ?= 0
PRESENTATION_PROFILE ?=
START_OVERLAYS ?=
FIXTURE_WORLD_LEVEL ?= 1
FIXTURE_WORLD_SPAWN ?= entry
FIXTURE_POINTER_MOVE ?= 0
CAPTURE_SETTLE ?= 10

capture:
	go run -tags ffmpeg ./cmd/client --capture-dir "$(CAPTURE_DIR)" --capture-scenes "$(CAPTURE_SCENES)" --capture-settle-frames "$(CAPTURE_SETTLE)" --start-scene "$(START_SCENE)" --start-overlays "$(START_OVERLAYS)" --fixture-characters "$(FIXTURE_CHARACTERS)" --fixture-world-level "$(FIXTURE_WORLD_LEVEL)" --fixture-world-spawn "$(FIXTURE_WORLD_SPAWN)" --fixture-pointer-move="$(FIXTURE_POINTER_MOVE)" --presentation-profile "$(PRESENTATION_PROFILE)"

# These focused entry points always select a deterministic character fixture.
# MPQ_DIRECTORY still points at the user's legally obtained content.
capture-game-world:
	$(MAKE) --no-print-directory capture CAPTURE_DIR="$(CAPTURE_DIR)" CAPTURE_SCENES=game_world START_SCENE=game_world FIXTURE_CHARACTERS=1 CAPTURE_SETTLE=60

capture-game-world-movement:
	$(MAKE) --no-print-directory capture CAPTURE_DIR="$(CAPTURE_DIR)" CAPTURE_SCENES=game_world START_SCENE=game_world FIXTURE_CHARACTERS=1 FIXTURE_POINTER_MOVE=1 CAPTURE_SETTLE=30

capture-blood-moor:
	$(MAKE) --no-print-directory capture CAPTURE_DIR="$(CAPTURE_DIR)" CAPTURE_SCENES=game_world START_SCENE=game_world FIXTURE_CHARACTERS=1 FIXTURE_WORLD_LEVEL=2 CAPTURE_SETTLE=60

# Capture both authoritative arrival points with production MPQ assets. The two
# images make collision, camera, depth, and art continuity reviewable together.
capture-act1-seam:
	$(MAKE) --no-print-directory capture CAPTURE_DIR="$(CAPTURE_DIR)/town" CAPTURE_SCENES=game_world START_SCENE=game_world FIXTURE_CHARACTERS=1 FIXTURE_WORLD_LEVEL=1 FIXTURE_WORLD_SPAWN=seam CAPTURE_SETTLE=60
	$(MAKE) --no-print-directory capture CAPTURE_DIR="$(CAPTURE_DIR)/blood-moor" CAPTURE_SCENES=game_world START_SCENE=game_world FIXTURE_CHARACTERS=1 FIXTURE_WORLD_LEVEL=2 FIXTURE_WORLD_SPAWN=seam CAPTURE_SETTLE=60

# Capture the world beneath every spatial overlay arrangement that changes the
# camera anchor. Artifacts stay local because original MPQ assets are required.
capture-game-world-panels:
	$(MAKE) --no-print-directory capture CAPTURE_DIR="$(CAPTURE_DIR)/left" CAPTURE_SCENES=character START_SCENE=game_world START_OVERLAYS=character FIXTURE_CHARACTERS=1 CAPTURE_SETTLE=60
	$(MAKE) --no-print-directory capture CAPTURE_DIR="$(CAPTURE_DIR)/right" CAPTURE_SCENES=inventory START_SCENE=game_world START_OVERLAYS=inventory FIXTURE_CHARACTERS=1 CAPTURE_SETTLE=60
	$(MAKE) --no-print-directory capture CAPTURE_DIR="$(CAPTURE_DIR)/both" CAPTURE_SCENES=character START_SCENE=game_world START_OVERLAYS=inventory,character FIXTURE_CHARACTERS=1 CAPTURE_SETTLE=60
	$(MAKE) --no-print-directory capture CAPTURE_DIR="$(CAPTURE_DIR)/full" CAPTURE_SCENES=help START_SCENE=game_world START_OVERLAYS=help FIXTURE_CHARACTERS=1 CAPTURE_SETTLE=60

play-game-world:
	go run -tags ffmpeg ./cmd/client --start-scene game_world --fixture-characters 1 --presentation-profile "$(PRESENTATION_PROFILE)"

play-monster-lab:
	go run -tags ffmpeg ./cmd/client --start-scene monster_lab

play-missile-lab:
	go run -tags ffmpeg ./cmd/client --start-scene missile_lab

play-combat-lab:
	go run -tags ffmpeg ./cmd/client --start-scene combat_lab --fixture-characters 1 --fixture-world-level 2

play-warp-lab:
	go run -tags ffmpeg ./cmd/client --start-scene warp_lab

capture-monster-lab:
	$(MAKE) --no-print-directory capture CAPTURE_DIR="$(CAPTURE_DIR)" CAPTURE_SCENES=monster_lab START_SCENE=monster_lab CAPTURE_SETTLE=30

capture-missile-lab:
	$(MAKE) --no-print-directory capture CAPTURE_DIR="$(CAPTURE_DIR)" CAPTURE_SCENES=missile_lab START_SCENE=missile_lab CAPTURE_SETTLE=30

capture-combat-lab:
	$(MAKE) --no-print-directory capture CAPTURE_DIR="$(CAPTURE_DIR)" CAPTURE_SCENES=combat_lab START_SCENE=combat_lab FIXTURE_CHARACTERS=1 FIXTURE_WORLD_LEVEL=2 CAPTURE_SETTLE=60

capture-warp-lab:
	$(MAKE) --no-print-directory capture CAPTURE_DIR="$(CAPTURE_DIR)" CAPTURE_SCENES=warp_lab START_SCENE=warp_lab CAPTURE_SETTLE=60

CAPTURE_ALL_DIR ?= ./captures/all-scenes
CAPTURE_ALL_FIXTURE_CHARACTERS ?= 10
CAPTURE_ALL_SCENES := loading title main_menu character_select character_create game_world game_loading tcpip realm_connecting realm_gateway realm_login realm_signup realm_recovery realm_characters realm_create realm_lobby realm_game_create credits cinematics font_lab ui_lab composite_lab monster_lab missile_lab combat_lab dt1_lab ds1_lab mapgen_lab warp_lab inventory character skills automap options pause help quests party stash cube hireling vendor waypoint quick_skills belt messages move_gold npc_interaction npc_dialogue item_tooltip ground_items confirmation_dialog death area_transition player_trade gambling npc_services hireling_hire chat overhead_labels

# Launch each scene in isolation because the capture session records visited
# scenes; it does not manufacture navigation through unrelated UI flows.
capture-all:
	@set -e; for scene in $(CAPTURE_ALL_SCENES); do \
		$(MAKE) --no-print-directory capture \
			CAPTURE_DIR="$(CAPTURE_ALL_DIR)/$$scene" \
			CAPTURE_SCENES="$$scene" \
			START_SCENE="$$scene" \
			FIXTURE_CHARACTERS="$(CAPTURE_ALL_FIXTURE_CHARACTERS)"; \
	done
