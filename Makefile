BINARY    := codeye
MODULE    := github.com/codeye/codeye
CMD       := ./cmd/codeye

VERSION   := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT    := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE      := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS   := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.buildDate=$(DATE)


# Resolve Go binary from common install locations (user GOPATH, /usr/local/go, system PATH)
GO_CANDIDATES := $(HOME)/go/bin/go /usr/local/go/bin/go /usr/bin/go go
GO            := $(or $(shell for g in $(GO_CANDIDATES); do command -v $$g 2>/dev/null && break; done), go)
GOFLAGS       := CGO_ENABLED=0

HUGO          := hugo
WEB_DIR       := web
WEB_CONFIG    := $(WEB_DIR)/config.toml

# ─── Build ────────────────────────────────────────────────────────────────────

.PHONY: build
build:
	$(GOFLAGS) $(GO) build -ldflags "$(LDFLAGS)" -o $(BINARY) $(CMD)

.PHONY: install
install:
	$(GOFLAGS) $(GO) install -ldflags "$(LDFLAGS)" $(CMD)

# Cross-compile to a specific target
# Usage: make build-linux-arm64
build-%:
	$(eval GOOS   := $(word 1, $(subst -, ,$*)))
	$(eval GOARCH := $(word 2, $(subst -, ,$*)))
	$(eval EXT    := $(if $(filter windows,$(GOOS)),.exe,))
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 \
		$(GO) build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-$(GOOS)-$(GOARCH)$(EXT) $(CMD)

.PHONY: cross
cross:
	@mkdir -p dist
	$(MAKE) build-linux-amd64
	$(MAKE) build-linux-arm64
	$(MAKE) build-linux-386
	$(MAKE) build-darwin-amd64
	$(MAKE) build-darwin-arm64
	$(MAKE) build-windows-amd64
	$(MAKE) build-freebsd-amd64

# ─── Test ─────────────────────────────────────────────────────────────────────

.PHONY: test
test:
	$(GOFLAGS) $(GO) test -count=1 -timeout=120s ./...

.PHONY: test-v
test-v:
	$(GOFLAGS) $(GO) test -v -count=1 -timeout=120s ./...

.PHONY: bench
bench:
	$(GOFLAGS) $(GO) test -run='^$$' -bench=. -benchmem -benchtime=3s ./...

.PHONY: cover
cover:
	$(GOFLAGS) $(GO) test -count=1 -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "coverage report → coverage.html"

# ─── Lint / Format ────────────────────────────────────────────────────────────

.PHONY: vet
vet:
	$(GOFLAGS) $(GO) vet ./...

.PHONY: fmt
fmt:
	$(GO) fmt ./...

.PHONY: lint
lint: vet fmt
	@which golangci-lint >/dev/null 2>&1 || (echo "golangci-lint not found — skipping" && exit 0)
	golangci-lint run ./...

# ─── Release ─────────────────────────────────────────────────────────────────

.PHONY: snapshot
snapshot:
	@which goreleaser >/dev/null 2>&1 || (echo "goreleaser not found; run: go install github.com/goreleaser/goreleaser/v2@latest" && exit 1)
	goreleaser release --snapshot --clean

.PHONY: release
release:
	@which goreleaser >/dev/null 2>&1 || (echo "goreleaser not found; run: go install github.com/goreleaser/goreleaser/v2@latest" && exit 1)
	goreleaser release --clean

# ─── Misc ────────────────────────────────────────────────────────────────────

.PHONY: clean
clean:
	rm -f $(BINARY) coverage.out coverage.html
	rm -rf dist/

# ─── Website ─────────────────────────────────────────────────────────────────

.PHONY: web
web:
	cd $(WEB_DIR) && $(HUGO) --config config.toml

.PHONY: web-serve
web-serve:
	cd $(WEB_DIR) && $(HUGO) server --config config.toml --port 1313 --bind 127.0.0.1

.PHONY: web-clean
web-clean:
	rm -rf $(WEB_DIR)/public $(WEB_DIR)/resources


.PHONY: tidy
tidy:
	$(GO) mod tidy

.PHONY: doctor
doctor: build
	./$(BINARY) doctor

.PHONY: help
help:
	@echo "Targets:"
	@echo "  build          build for current platform"
	@echo "  install        install to GOPATH/bin"
	@echo "  cross          cross-compile for all platforms"
	@echo "  test           run all tests"
	@echo "  test-v         run all tests (verbose)"
	@echo "  bench          run benchmarks"
	@echo "  cover          generate coverage report"
	@echo "  vet            run go vet"
	@echo "  fmt            run go fmt"
	@echo "  lint           run golangci-lint"
	@echo "  snapshot       goreleaser dry-run"
	@echo "  release        goreleaser production release"
	@echo "  clean          remove build artifacts"
	@echo "  tidy           go mod tidy"
	@echo "  doctor         build and run codeye doctor"
