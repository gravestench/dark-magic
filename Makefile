.PHONY: test test-race fmt vet shim bik-view profile profile-check

test:
	go test ./...

test-race:
	go test -race ./...

fmt:
	gofmt -w $$(find cmd internal pkg -name '*.go')

vet:
	go vet ./...

shim:
	go run ./internal/tools/shim_pack -output ./dist/darkmagic.zip

bik-view:
	go run ./internal/tools/bik_view -source "$${MPQ_DIRECTORY}" -asset "$${BIK_ASSET}"

PROFILE_DIR ?= ./profiles/acceptance
PROFILE_BUDGETS ?= ./docs/profile-budgets.json

profile:
	go run -tags ffmpeg ./cmd/darkmagic --profile-dir "$(PROFILE_DIR)" --profile-scenes all

profile-check:
	go run ./internal/tools/profile_check -profile-dir "$(PROFILE_DIR)" -budgets "$(PROFILE_BUDGETS)"
