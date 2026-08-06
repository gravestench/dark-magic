.PHONY: test test-race fmt vet

test:
	go test ./...

test-race:
	go test -race ./...

fmt:
	gofmt -w $$(find cmd internal pkg -name '*.go')

vet:
	go vet ./...
