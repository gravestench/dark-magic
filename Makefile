.PHONY: test test-race fmt vet

test:
	go test ./...

test-race:
	go test -race ./pkg/assetinspect ./pkg/scene ./pkg/services/fileLoader ./pkg/services/luaManager ./pkg/services/luaModLoader ./pkg/services/tweens ./pkg/cache

fmt:
	gofmt -w $$(find cmd internal pkg -name '*.go')

vet:
	go vet ./...
