#!/bin/bash
set -euo pipefail
cd -- "$(dirname -- "${BASH_SOURCE[0]}")"

for comando in go sqlc templ docker; do
    if ! command -v "$comando" >/dev/null 2>&1; then
        echo "Falta $comando. Revisá los requisitos del README.md." >&2
        exit 1
    fi
done
docker compose version >/dev/null

if [[ ! -f .env ]]; then
    cp .env.example .env
    echo "Se creó .env a partir de .env.example."
fi
set -a
. ./.env
set +a
: "${DATABASE_URL:?Configurar DATABASE_URL en .env}"

APP_PROJECT="progweb-tp6-app"
APP_BIN="./tmp/runner-log"
if [[ "$(go env GOOS)" == "windows" ]]; then APP_BIN+=".exe"; fi
APP_PID=""

cleanup() {
    local resultado=$?
    trap - EXIT
    if [[ -n "$APP_PID" ]] && kill -0 "$APP_PID" 2>/dev/null; then
        kill "$APP_PID" || true
        wait "$APP_PID" 2>/dev/null || true
    fi
    if ! docker compose -p "$APP_PROJECT" down -v --remove-orphans; then resultado=1; fi
    exit "$resultado"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

echo "Limpiando contenedores y volúmenes anteriores del TP6..."
docker compose -p "$APP_PROJECT" down -v --remove-orphans

echo "Generando sqlc, templ y compilando la aplicación..."
sqlc generate
templ generate
mkdir -p tmp
go build -o "$APP_BIN" .

echo "Iniciando PostgreSQL y esperando a que esté listo..."
docker compose -p "$APP_PROJECT" up -d --wait --wait-timeout 60

echo "Ejecutando Runner Log - Tracker. Ctrl+C detiene y limpia la entrega."
"$APP_BIN" &
APP_PID=$!
wait "$APP_PID"
