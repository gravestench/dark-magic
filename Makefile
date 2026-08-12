.PHONY: test architecture test-race fmt vet d2legacy bik-view presentation-coverage profile profile-acceptance profile-check capture capture-all capture-game-world capture-game-world-movement capture-game-world-panels capture-blood-moor capture-act1-seam capture-monster-lab capture-missile-lab capture-combat-lab capture-warp-lab play-game-world play-monster-lab play-missile-lab play-combat-lab play-warp-lab

test:
	go test ./...

architecture:
	go test ./internal/acceptance -run 'Test(Retired|LegacyRenderer|NoAccidental|DependencyDirection|CommandRemains|GameplayOwnership)'

test-race:
	go test -race ./...

fmt:
	gofmt -w $$(find cmd internal pkg -name '*.go')

vet:
	go vet ./...

d2legacy:
	go run ./internal/dev/tools/d2legacy_pack -output ./dist/d2legacy.zip

bik-view:
	go run ./internal/dev/tools/bik_view -source "$${MPQ_DIRECTORY}" -asset "$${BIK_ASSET}"

presentation-coverage:
	go run ./internal/dev/tools/presentation_coverage

PROFILE_DIR ?= ./profiles/acceptance
PROFILE_BUDGETS ?= ./docs/profile-budgets.json

profile:
	go run -tags ffmpeg ./cmd/darkmagic --profile-dir "$(PROFILE_DIR)" --profile-scenes all

# Deterministic real-asset runs for every budgeted scene. game_loading is
# profiled while the capture waits for the game_world it naturally enters.
profile-acceptance:
	go run -tags ffmpeg ./cmd/darkmagic --profile-dir "$(PROFILE_DIR)" --profile-scenes loading --capture-dir "$(PROFILE_DIR)/captures/loading" --capture-scenes loading --capture-settle-frames 60 --start-scene loading --fixture-characters 3
	go run -tags ffmpeg ./cmd/darkmagic --profile-dir "$(PROFILE_DIR)" --profile-scenes title --capture-dir "$(PROFILE_DIR)/captures/title" --capture-scenes title --capture-settle-frames 60 --start-scene title --fixture-characters 3
	go run -tags ffmpeg ./cmd/darkmagic --profile-dir "$(PROFILE_DIR)" --profile-scenes main_menu --capture-dir "$(PROFILE_DIR)/captures/main_menu" --capture-scenes main_menu --capture-settle-frames 60 --start-scene main_menu --fixture-characters 3
	go run -tags ffmpeg ./cmd/darkmagic --profile-dir "$(PROFILE_DIR)" --profile-scenes character_select --capture-dir "$(PROFILE_DIR)/captures/character_select" --capture-scenes character_select --capture-settle-frames 60 --start-scene character_select --fixture-characters 3
	go run -tags ffmpeg ./cmd/darkmagic --profile-dir "$(PROFILE_DIR)" --profile-scenes character_create --capture-dir "$(PROFILE_DIR)/captures/character_create" --capture-scenes character_create --capture-settle-frames 60 --start-scene character_create --fixture-characters 3
	go run -tags ffmpeg ./cmd/darkmagic --profile-dir "$(PROFILE_DIR)" --profile-scenes game_loading --capture-dir "$(PROFILE_DIR)/captures/game_loading" --capture-scenes game_world --capture-settle-frames 120 --start-scene game_loading --fixture-characters 3 --fixture-world-level 2
	go run -tags ffmpeg ./cmd/darkmagic --profile-dir "$(PROFILE_DIR)" --profile-scenes game_world --capture-dir "$(PROFILE_DIR)/captures/game_world" --capture-scenes game_world --capture-settle-frames 180 --start-scene game_world --fixture-characters 1 --fixture-world-level 2 --fixture-pointer-move=true

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
	go run -tags ffmpeg ./cmd/darkmagic --capture-dir "$(CAPTURE_DIR)" --capture-scenes "$(CAPTURE_SCENES)" --capture-settle-frames "$(CAPTURE_SETTLE)" --start-scene "$(START_SCENE)" --start-overlays "$(START_OVERLAYS)" --fixture-characters "$(FIXTURE_CHARACTERS)" --fixture-world-level "$(FIXTURE_WORLD_LEVEL)" --fixture-world-spawn "$(FIXTURE_WORLD_SPAWN)" --fixture-pointer-move="$(FIXTURE_POINTER_MOVE)" --presentation-profile "$(PRESENTATION_PROFILE)"

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
	go run -tags ffmpeg ./cmd/darkmagic --start-scene game_world --fixture-characters 1 --presentation-profile "$(PRESENTATION_PROFILE)"

play-monster-lab:
	go run -tags ffmpeg ./cmd/darkmagic --start-scene monster_lab

play-missile-lab:
	go run -tags ffmpeg ./cmd/darkmagic --start-scene missile_lab

play-combat-lab:
	go run -tags ffmpeg ./cmd/darkmagic --start-scene combat_lab --fixture-characters 1 --fixture-world-level 2

play-warp-lab:
	go run -tags ffmpeg ./cmd/darkmagic --start-scene warp_lab

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
CAPTURE_ALL_SCENES := loading title main_menu character_select character_create game_world game_loading tcpip credits cinematics font_lab ui_lab composite_lab monster_lab missile_lab combat_lab dt1_lab ds1_lab mapgen_lab warp_lab inventory character skills automap options pause help quests party stash cube hireling vendor waypoint quick_skills belt messages move_gold npc_interaction npc_dialogue item_tooltip ground_items confirmation_dialog death area_transition player_trade gambling npc_services hireling_hire chat overhead_labels

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
