#!/bin/bash
set -e

APP_BIN="./tmp/api"
API_PID=""

# Limpieza que se ejecuta siempre para salir
cleanup() {
    echo "Limpiando Entorno..."

    if [ -n "$API_PID" ] && kill -0 "$API_PID" 2>/dev/null; then
        echo "-> Deteniendo servidor Go (PID: $API_PID)..."
        kill "$API_PID" || true
        wait "$API_PID" 2>/dev/null || true
    fi
    echo "-> Borrando contenedores y volúmenes de Docker..."
    docker compose down -v
    echo "-> Limpiando binarios..."
    rm -rf tmp
    echo "Limpieza realizada"
}

trap cleanup EXIT

echo "Trabajo Previo: Preparando infraestructura y código"

echo "Borrando contenedores y volúmenes anteriores..."
docker compose down -v

echo "Generando código de base de datos  sqlc..."
sqlc generate

echo "Compilando Go..."
mkdir -p tmp
go build -o "$APP_BIN" .

echo "Levantando contenedor PostgreSQL..."
docker compose up -d

echo "Esperando a que PostgreSQL esté lista..."
until docker compose exec -T database pg_isready -U tp_user -d tp_db > /dev/null 2>&1; do
    sleep 1
done
sleep 2
echo "Base de datos lista"

echo "Levantando API REST"
"$APP_BIN" &
API_PID=$!

echo "Esperando a que el server HTTP escuche en :8080..."
for i in {1..15}; do
    if curl -s http://localhost:8080/zapatillas > /dev/null; then
        break
    fi
    sleep 1
done

echo "Prueba de Endpoints (Ejecutando requests.bash)"

bash ./requests.bash