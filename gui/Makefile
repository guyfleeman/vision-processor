BINARY = ssl-vision-processor-gui

# ./... would also descend into frontend/node_modules, where npm packages ship
# stray Go source (flatted/golang, for one). Enumerate our own trees instead.
GO_PACKAGES = ./cmd/... ./internal/... ./frontend

FRONTEND_SOURCES = $(shell find frontend -type f \
	-not -path "frontend/node_modules/*" \
	-not -path "frontend/dist/*")

.PHONY: all frontend install test run proto clean

all: install

# Sentinel file, so npm reruns only when a frontend source is newer than the
# last build. //go:embed resolves at compile time, which is why every Go target
# below depends on this.
.frontend: $(FRONTEND_SOURCES)
	cd frontend && npm ci && npm run build
	touch .frontend

frontend: .frontend

install: frontend
	go install -v $(GO_PACKAGES)

test: frontend
	cd frontend && npm run check && npm run lint && npm run format:check
	go test $(GO_PACKAGES)

run: frontend
	go run ./cmd/$(BINARY)

# Regenerates Go and TypeScript protobuf bindings. Needs network access for
# buf's remote plugins, so it never runs as part of a build.
proto:
	buf generate

clean:
	rm -rf frontend/dist .frontend
	mkdir -p frontend/dist
	touch frontend/dist/.gitkeep
