#!/bin/bash
set -euo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")"

if [[ ! -f .env ]]; then
    echo "Copiá .env.example a .env antes de ejecutar las pruebas." >&2
    exit 1
fi
set -a
. ./.env
set +a

# Proyecto y puertos independientes: no se reinicia la base de desarrollo.
TEST_PROJECT="progweb-tp5-tests"
export DB_PORT="${TEST_DB_PORT:-55432}"
export TEST_DATABASE_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@127.0.0.1:${DB_PORT}/${POSTGRES_DB}?sslmode=disable"
export DATABASE_URL="$TEST_DATABASE_URL"
export HTTP_ADDR="127.0.0.1:${TEST_HTTP_PORT:-18080}"
export API_URL="http://$HTTP_ADDR"
APP_BIN="./tmp/api-test"
if [[ "$(go env GOOS)" == "windows" ]]; then APP_BIN+=".exe"; fi
API_PID=""

cleanup() {
    local resultado=$?
    trap - EXIT
    if [[ -n "$API_PID" ]] && kill -0 "$API_PID" 2>/dev/null; then
        kill "$API_PID" || true
        wait "$API_PID" 2>/dev/null || true
    fi
    # Solo se elimina el volumen del proyecto de pruebas creado por este script.
    if ! docker compose -p "$TEST_PROJECT" down -v; then resultado=1; fi
    rm -f "$APP_BIN"
    exit "$resultado"
}
trap cleanup EXIT
trap 'exit 130' INT TERM

echo "Generando sqlc, templ y compilando la aplicación..."
sqlc generate
templ generate
mkdir -p tmp
go build -o "$APP_BIN" .
docker compose -p "$TEST_PROJECT" up -d

lista=false
for intento in {1..60}; do
    if docker compose -p "$TEST_PROJECT" exec -T database pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB" >/dev/null 2>&1; then
        lista=true
        break
    fi
    sleep 1
done
if [[ "$lista" != true ]]; then echo "PostgreSQL no inició a tiempo." >&2; exit 1; fi

echo "Ejecutando pruebas de Go..."
go test -p 1 -count=1 -v ./...
"$APP_BIN" >tmp/api-test.log 2>&1 &
API_PID=$!
lista=false
for intento in {1..30}; do
    if curl --fail --silent "$API_URL/" >/dev/null; then lista=true; break; fi
    if ! kill -0 "$API_PID" 2>/dev/null; then break; fi
    sleep 1
done
if [[ "$lista" != true ]]; then cat tmp/api-test.log >&2; exit 1; fi
bash ./requests.bash
echo "Pruebas de Go y HTTP completadas."
