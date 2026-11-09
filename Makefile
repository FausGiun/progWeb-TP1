SHELL := /bin/bash
APP_NAME := api

ifneq ($(wildcard .env),)
include .env
export
endif

.PHONY: all check-env generate build db-up run test test-api clean

all: build

check-env:
	@test -n "$$DATABASE_URL" || { echo "Copiá .env.example a .env y configurá la base de datos."; exit 1; }

generate:
	@sqlc generate

build: generate
	@mkdir -p tmp
	@go build -o tmp/$(APP_NAME) .

db-up: check-env
	@docker compose up -d
	@for intento in {1..60}; do \
		if docker compose exec -T database pg_isready -U "$$POSTGRES_USER" -d "$$POSTGRES_DB" >/dev/null 2>&1; then exit 0; fi; \
		sleep 1; \
	done; echo "PostgreSQL no inició a tiempo."; exit 1

# Ctrl+C detiene el servidor y los contenedores, conservando el volumen de datos.
run: check-env build db-up
	@trap 'docker compose down' EXIT; trap 'exit 130' INT TERM; ./tmp/$(APP_NAME)

# El script ejecuta Go y las peticiones HTTP en una base de pruebas separada.
test:
	@./test-api.sh

test-api: test

clean:
	@docker compose down
	@rm -rf tmp
