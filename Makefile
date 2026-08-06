.PHONY: test architecture test-race fmt vet shim bik-view profile profile-check capture

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

capture:
	go run -tags ffmpeg ./cmd/darkmagic --capture-dir "$(CAPTURE_DIR)" --capture-scenes "$(CAPTURE_SCENES)" --start-scene "$(START_SCENE)" --fixture-characters "$(FIXTURE_CHARACTERS)"
