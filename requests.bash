#!/bin/bash
set -euo pipefail

URL="${API_URL:?Ejecutá bash ./test-app.sh para usar la instancia de pruebas}"
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

ruta_borrado() {
    local panel="$1" marca="$2" recurso="$3" fin_fila="$4"
    local html fila patron
    html="$(<"$RESPUESTA")"
    if [[ "$html" != '<div id="'"$panel"'"'* || "$html" != *"$marca"* ]]; then
        echo "Se esperaba el fragmento $panel con el registro recién creado." >&2
        return 1
    fi
    fila="${html#*"$marca"}"
    fila="${fila%%"$fin_fila"*}"
    patron='hx-delete="(/'"$recurso"'/[0-9]+)"'
    if [[ ! "$fila" =~ $patron ]]; then
        echo "Falta la ruta de borrado del registro en $panel." >&2
        return 1
    fi
    printf '%s\n' "${BASH_REMATCH[1]}"
}

borrar() {
    peticion 200 DELETE "$1" -H 'HX-Request: true'
    if [[ -s "$RESPUESTA" ]]; then
        echo "El DELETE de $1 debe devolver un cuerpo vacío." >&2
        exit 1
    fi
}

peticion 200 GET /
peticion 200 GET /mis-zapatillas
peticion 200 GET /static/styles.css
peticion 405 GET /zapatillas
peticion 400 POST /zapatillas -H 'HX-Request: true' \
    --data-urlencode 'marca_modelo=' --data-urlencode 'kms_acumulados=0'

marca="Par prueba HTTP TP6-$$"
peticion 200 POST /zapatillas -H 'HX-Request: true' \
    --data-urlencode "marca_modelo=$marca" --data-urlencode 'kms_acumulados=12.5'
ruta_zapatilla="$(ruta_borrado panel-zapatillas "$marca" zapatillas '</li>')"

lugar="Salida prueba HTTP TP6-$$"
peticion 200 POST /entrenamientos -H 'HX-Request: true' \
    --data-urlencode 'fecha=2025-05-10' --data-urlencode 'distancia_km=10' \
    --data-urlencode 'tiempo_min=50' --data-urlencode 'tipo=Fondo' \
    --data-urlencode "zapatilla_id=${ruta_zapatilla##*/}" --data-urlencode "lugar=$lugar"
ruta_entrenamiento="$(ruta_borrado panel-entrenamientos "$lugar" entrenamientos '</tr>')"

borrar "$ruta_entrenamiento"
borrar "$ruta_zapatilla"
echo 'Altas y bajas HTMX comprobadas por HTTP.'
