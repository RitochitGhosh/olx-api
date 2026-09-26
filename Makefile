# Mark command targets as .PHONY so Make always executes them.
# Without this, if a file/directory named "build" exists, Make may consider
# the target already up to date and skip the build command.
.PHONY: build run test clean

build:
	@go build -o bin/api ./cmd/api

run: build
	@./bin/api
