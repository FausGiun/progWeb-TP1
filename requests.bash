#!/bin/bash
set -euo pipefail

URL="${API_URL:?Ejecutá bash ./test-api.sh para usar la instancia de pruebas}"
RESPUESTA="$(mktemp)"
CABECERAS="$(mktemp)"
trap 'rm -f "$RESPUESTA" "$CABECERAS"' EXIT

peticion() {
    local esperado="$1" metodo="$2" ruta="$3"
    shift 3
    local obtenido
    obtenido="$(curl --silent --show-error --output "$RESPUESTA" --dump-header "$CABECERAS" --write-out '%{http_code}' -X "$metodo" "$URL$ruta" "$@")"
    if [[ "$obtenido" != "$esperado" ]]; then
        echo "ERROR: $metodo $ruta devolvió $obtenido; esperado $esperado." >&2
        cat "$RESPUESTA" >&2
        exit 1
    fi
    echo "OK: $metodo $ruta -> $obtenido"
}

peticion 200 GET /
peticion 200 GET /mis-zapatillas
peticion 303 GET /mis-entrenamientos
peticion 200 GET /static/styles.css
peticion 405 GET /zapatillas
peticion 400 POST /zapatillas --data-urlencode 'marca_modelo=' --data-urlencode 'kms_acumulados=0'
peticion 303 POST /zapatillas --data-urlencode 'marca_modelo=TP5 prueba HTTP' --data-urlencode 'kms_acumulados=0'
if [[ "$(<"$CABECERAS")" != *'Location: /mis-zapatillas'* ]]; then
    echo "El POST debe redirigir al listado de zapatillas." >&2
    exit 1
fi
peticion 200 GET /mis-zapatillas
if [[ "$(<"$RESPUESTA")" != *'TP5 prueba HTTP'* ]]; then
    echo "El registro nuevo debe estar en el HTML devuelto por Go." >&2
    exit 1
fi
echo "Formularios, redirecciones y HTML del TP5 comprobados."
