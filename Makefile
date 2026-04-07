BINARY  := kt
MAIN    := ./cmd/kt
DIST    := dist

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -s -w"

PLATFORMS := \
	darwin/amd64 \
	darwin/arm64 \
	linux/amd64 \
	linux/arm64 \
	windows/amd64

.PHONY: build install clean release $(PLATFORMS)

## build: compile for the current platform → dist/kt
build:
	@mkdir -p $(DIST)
	go build $(LDFLAGS) -o $(DIST)/$(BINARY) $(MAIN)

## install: install to $GOPATH/bin (or $GOBIN)
install:
	go install $(LDFLAGS) $(MAIN)

## release: cross-compile for all platforms → dist/
release: $(PLATFORMS)

$(PLATFORMS):
	$(eval OS   := $(word 1, $(subst /, ,$@)))
	$(eval ARCH := $(word 2, $(subst /, ,$@)))
	$(eval EXT  := $(if $(filter windows,$(OS)),.exe,))
	@mkdir -p $(DIST)
	GOOS=$(OS) GOARCH=$(ARCH) go build $(LDFLAGS) \
		-o $(DIST)/$(BINARY)-$(OS)-$(ARCH)$(EXT) $(MAIN)
	@echo "  built $(DIST)/$(BINARY)-$(OS)-$(ARCH)$(EXT)"

## clean: remove dist/
clean:
	rm -rf $(DIST)

## help: list available targets
help:
	@grep -E '^## ' Makefile | sed 's/^## /  /'
