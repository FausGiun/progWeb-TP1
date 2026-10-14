#!/bin/bash
set -e

APP_BIN="./tmp/api"
API_PID=""

# Función de limpieza (3. Tareas posteriores) que se ejecuta siempre al salir
cleanup() {
    echo ""
    echo "=========================================================="
    echo "🧹 3) TAREAS POSTERIORES: Limpiando entorno"
    echo "=========================================================="
    if [ -n "$API_PID" ] && kill -0 "$API_PID" 2>/dev/null; then
        echo "-> Deteniendo servidor Go (PID: $API_PID)..."
        kill "$API_PID" || true
        wait "$API_PID" 2>/dev/null || true
    fi
    echo "-> Borrando contenedores y volúmenes de Docker..."
    docker compose down -v
    echo "-> Limpiando binarios compilados..."
    rm -rf tmp
    echo "✅ Limpieza completada."
}

trap cleanup EXIT

echo "=========================================================="
echo "⚙️  1) TAREAS PREVIAS: Preparando infraestructura y código"
echo "=========================================================="
echo "-> Borrando contenedores y volúmenes previos..."
docker compose down -v

echo "-> Generando código de base de datos con sqlc..."
sqlc generate

echo "-> Compilando la aplicación Go..."
mkdir -p tmp
go build -o "$APP_BIN" .

echo "-> Levantando contenedor de PostgreSQL..."
docker compose up -d

echo "-> Esperando a que la base de datos PostgreSQL esté lista..."
until docker compose exec -T database pg_isready -U tp_user -d tp_db > /dev/null 2>&1; do
    sleep 1
done
sleep 2
echo "-> ¡Base de datos lista!"

echo "-> Levantando la aplicación API REST en segundo plano..."
"$APP_BIN" &
API_PID=$!

echo "-> Esperando a que el servidor HTTP escuche en :8080..."
for i in {1..15}; do
    if curl -s http://localhost:8080/zapatillas > /dev/null; then
        break
    fi
    sleep 1
done

echo ""
echo "=========================================================="
echo "🧪 2) PRUEBA DE ENDPOINTS (Ejecutando requests.bash)"
echo "=========================================================="
bash ./requests.bash