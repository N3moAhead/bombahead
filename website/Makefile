CSS_INPUT = ./static/css/app.css
CSS_OUTPUT = ./static/css/dist.css

.PHONY: fmt vet templ tailwind build run dev docs

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

templ:
	go tool templ generate

tailwind:
	npx @tailwindcss/cli -i $(CSS_INPUT) -o $(CSS_OUTPUT) --minify

docs:
	@echo "--- Generating Docs ---"
	go run cmd/generate_docs/main.go

build: templ tailwind docs vet
	go build -o bombahead ./cmd/website/main.go

run: build
	./bombahead

image-website:
	podman build -f Dockerfile.Website -t ghcr.io/n3moahead/bombahead/website

image-website-public: image-website
	podman push ghcr.io/n3moahead/bombahead/website:latest

image-matchmaker:
	podman build -f Dockerfile.Matchmaker -t ghcr.io/n3moahead/bombahead/matchmaker

image-matchmaker-public: image-matchmaker
	podman push ghcr.io/n3moahead/bombahead/matchmaker:latest

lint:
	@echo "==> Running linter..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@golangci-lint run ./...

dev:
	@if ! command -v air > /dev/null 2>&1; then \
		echo "air is not installed. Installing..."; \
		go install github.com/air-verse/air@latest; \
	fi
	air
