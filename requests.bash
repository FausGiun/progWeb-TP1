#!/bin/bash
set -euo pipefail

URL="${API_URL:-http://localhost:8080}"
RESPUESTA="$(mktemp)"
trap 'rm -f "$RESPUESTA"' EXIT

peticion() {
    local esperado="$1" metodo="$2" ruta="$3"
    shift 3
    local obtenido
    obtenido="$(curl --silent --show-error --output "$RESPUESTA" --write-out '%{http_code}' -X "$metodo" "$URL$ruta" "$@")"
    if [[ "$obtenido" != "$esperado" ]]; then
        echo "ERROR: $metodo $ruta devolvió $obtenido; esperado $esperado." >&2
        cat "$RESPUESTA" >&2
        exit 1
    fi
    echo "OK: $metodo $ruta -> $obtenido"
}

id_creado() {
    local id
    id="$(sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p' "$RESPUESTA")"
    if [[ -z "$id" ]]; then echo "La respuesta no contiene el ID creado." >&2; exit 1; fi
    printf '%s' "$id"
}

peticion 200 GET /
peticion 200 GET /static/app.js
peticion 200 GET /static/styles.css
peticion 201 POST /zapatillas -H 'Content-Type: application/json' -d '{"marca_modelo":"Nike Pegasus 40","kms_acumulados":15.5}'
ZAPATILLA_ID="$(id_creado)"
peticion 400 POST /zapatillas -H 'Content-Type: application/json' -d '{"marca_modelo":"","kms_acumulados":0}'
peticion 200 GET /zapatillas
peticion 200 GET "/zapatillas/$ZAPATILLA_ID"
peticion 200 PUT "/zapatillas/$ZAPATILLA_ID" -H 'Content-Type: application/json' -d '{"marca_modelo":"Nike Pegasus 40 Pro","kms_acumulados":50,"estado":true}'
peticion 201 POST /entrenamientos -H 'Content-Type: application/json' -d "{\"fecha\":\"2025-05-10\",\"distancia_km\":10,\"tiempo_min\":50,\"tipo\":\"Fondo\",\"zapatilla_id\":$ZAPATILLA_ID,\"lugar\":\"Parque Independencia\"}"
ENTRENAMIENTO_ID="$(id_creado)"
peticion 200 GET /entrenamientos
peticion 200 GET "/entrenamientos/$ENTRENAMIENTO_ID"
peticion 200 PUT "/entrenamientos/$ENTRENAMIENTO_ID" -H 'Content-Type: application/json' -d "{\"fecha\":\"2025-05-10\",\"distancia_km\":12,\"tiempo_min\":60,\"tipo\":\"Pasadas\",\"zapatilla_id\":$ZAPATILLA_ID,\"lugar\":\"Pista\"}"
peticion 204 DELETE "/entrenamientos/$ENTRENAMIENTO_ID"
peticion 404 GET "/entrenamientos/$ENTRENAMIENTO_ID"
peticion 204 DELETE "/zapatillas/$ZAPATILLA_ID"
peticion 405 PATCH /zapatillas
echo "Todas las respuestas HTTP coinciden con los estados esperados."
