.PHONY: test architecture test-race fmt vet shim bik-view presentation-coverage profile profile-check capture capture-all

test:
	go test ./...

architecture:
	go test ./internal/acceptance -run 'Test(Retired|NoAccidental|DependencyDirection|CommandRemains)'

test-race:
	go test -race ./...

fmt:
	gofmt -w $$(find cmd internal pkg -name '*.go')

vet:
	go vet ./...

shim:
	go run ./internal/dev/tools/shim_pack -output ./dist/darkmagic.zip

bik-view:
	go run ./internal/dev/tools/bik_view -source "$${MPQ_DIRECTORY}" -asset "$${BIK_ASSET}"

presentation-coverage:
	go run ./internal/dev/tools/presentation_coverage

PROFILE_DIR ?= ./profiles/acceptance
PROFILE_BUDGETS ?= ./docs/profile-budgets.json

profile:
	go run -tags ffmpeg ./cmd/darkmagic --profile-dir "$(PROFILE_DIR)" --profile-scenes all

profile-check:
	go run ./internal/dev/tools/profile_check -profile-dir "$(PROFILE_DIR)" -budgets "$(PROFILE_BUDGETS)"

CAPTURE_DIR ?= ./captures/frontend
CAPTURE_SCENES ?= loading,title
START_SCENE ?=
FIXTURE_CHARACTERS ?= 0
PRESENTATION_PROFILE ?=

capture:
	go run -tags ffmpeg ./cmd/darkmagic --capture-dir "$(CAPTURE_DIR)" --capture-scenes "$(CAPTURE_SCENES)" --start-scene "$(START_SCENE)" --fixture-characters "$(FIXTURE_CHARACTERS)" --presentation-profile "$(PRESENTATION_PROFILE)"

CAPTURE_ALL_DIR ?= ./captures/all-scenes
CAPTURE_ALL_FIXTURE_CHARACTERS ?= 10
CAPTURE_ALL_SCENES := loading title main_menu character_select character_create game_world game_loading tcpip credits cinematics font_lab ui_lab inventory character skills automap options pause help quests party stash cube hireling vendor waypoint quick_skills belt messages move_gold npc_interaction npc_dialogue item_tooltip ground_items confirmation_dialog death area_transition player_trade gambling npc_services hireling_hire chat overhead_labels

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
