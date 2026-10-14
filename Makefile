APP_NAME := api

.PHONY: all generate build run db-up db-wait test test-api clean

all: build

# Genera el código seguro de acceso a datos con sqlc
generate:
	@sqlc generate

# Compila el binario en la carpeta temporal tmp/
build: generate
	@mkdir -p tmp
	@go build -o tmp/$(APP_NAME) .

# Levanta la base de datos y espera a que acepte conexiones
db-up:
	@docker compose up -d
	@echo "Esperando a que PostgreSQL esté listo..."
	@until docker compose exec -T database pg_isready -U tp_user -d tp_db > /dev/null 2>&1; do sleep 1; done

# Automatiza todo el arranque de la aplicación (Recomendación del profesor)
run: build db-up
	@./tmp/$(APP_NAME)

# Ejecuta los tests unitarios de Go (paquete logic, db/sqlc y handlers)
test: generate
	@docker compose down -v
	@docker compose up -d
	@until docker compose exec -T database pg_isready -U tp_user -d tp_db > /dev/null 2>&1; do sleep 1; done
	@sleep 2
	@go test -p 1 -v ./...
	@docker compose down -v

# Ejecuta el ciclo completo exigido en el TP3 (Tareas previas -> cURL a cada endpoint -> Limpieza)
test-api:
	@./test-api.sh

# Limpia contenedores, volúmenes y artefactos compilados
clean:
	@docker compose down -v
	@rm -rf tmp main tp1