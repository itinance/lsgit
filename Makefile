.PHONY: build install clean release snapshot test

BINARY   := lsgit
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS  := -s -w -X main.version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

install: build
	install -m 755 $(BINARY) /usr/local/bin/$(BINARY)

uninstall:
	rm -f /usr/local/bin/$(BINARY)

clean:
	rm -f $(BINARY)
	rm -rf dist/

test:
	go test ./...

# Dry-run release (builds all targets locally, no publish)
snapshot:
	goreleaser release --snapshot --clean

# Full release — requires GITHUB_TOKEN and HOMEBREW_TAP_TOKEN env vars
release:
	goreleaser release --clean
