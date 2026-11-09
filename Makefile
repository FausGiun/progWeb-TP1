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

# Entrega TP5: limpiar, generar, compilar, iniciar, ejecutar y limpiar al salir.
run-app:
	@bash ./run-app.sh

run: run-app

test:
	@bash ./test-app.sh

clean:
	@docker compose -p progweb-tp5-app down -v --remove-orphans
	@rm -f tmp/runner-log tmp/runner-log.exe
