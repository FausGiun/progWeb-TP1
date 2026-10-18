APP_NAME := api

.PHONY: all generate build db-up run test test-api clean

all: build

# Genera el código seguro de acceso a datos con sqlc
generate:
	@sqlc generate

# Construye el binario de la aplicación en tmp/
build: generate
	@mkdir -p tmp
	@go build -o tmp/$(APP_NAME) .

# Levanta PostgreSQL y espera a que acepte conexiones
db-up:
	@docker compose up -d
	@echo "Esperando a que PostgreSQL esté listo..."
	@until docker compose exec -T database pg_isready -U tp_user -d tp_db > /dev/null 2>&1; do sleep 1; done

# Levanta la BD, corre la API y apaga el contenedor automáticamente al presionar Ctrl+C
run: build db-up
	@trap 'echo "\n-> Deteniendo contenedor de base de datos..."; docker compose down' INT TERM EXIT; ./tmp/$(APP_NAME)

# Ejecuta las pruebas unitarias y de integración limpiando el entorno antes y después
test: generate
	@docker compose down -v
	@docker compose up -d
	@until docker compose exec -T database pg_isready -U tp_user -d tp_db > /dev/null 2>&1; do sleep 1; done
	@sleep 2
	@go test -p 1 -count=1 -v ./...
	@docker compose down -v

# Ejecuta la prueba integral de la API REST 
test-api:
	@./test-api.sh

# Limpia contenedores, volúmenes y artefactos de compilación
clean:
	@docker compose down -v
	@rm -rf tmp main tp1 api