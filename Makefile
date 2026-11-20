SHELL := /bin/bash
APP_NAME := runner-log

.PHONY: all generate build run-app run test clean

all: build

generate:
	@sqlc generate
	@templ generate

build: generate
	@mkdir -p tmp
	@go build -o tmp/$(APP_NAME) .

run-app:
	@bash ./run-app.sh

run: run-app

test:
	@bash ./test-app.sh

clean:
	@docker compose -p progweb-tp6-app down -v --remove-orphans
	@rm -f tmp/runner-log tmp/runner-log.exe
