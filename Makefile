.PHONY: test test-race fmt vet shim bik-view

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
