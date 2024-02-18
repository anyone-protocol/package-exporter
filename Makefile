.PHONY: help
help: # Display this help
	@awk 'BEGIN{FS=":.*#";printf "Usage:\n  make <target>\n\nTargets:\n"}/^[a-zA-Z_-]+:.*?#/{printf"  %-10s %s\n",$$1,$$2}' $(MAKEFILE_LIST)

.PHONY: build
build: # Build downloads-exporter to bin/ directory
	go build -o bin/downloads-exporter cmd/exporter/main.go

.PHONY: run
run: build # Runs downloads-exporter after build
	./bin/downloads-exporter