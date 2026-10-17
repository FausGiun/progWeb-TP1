#!/bin/bash
set -e

URL="http://localhost:8080"

echo "Pruebas CRUD Zapatillas"

echo -e "\n-> [POST /zapatillas] Creando zapatilla (Estado esperado: 201)..."
curl -i -s -X POST "$URL/zapatillas" \
  -H "Content-Type: application/json" \
  -d '{"marca_modelo": "Nike Pegasus 40", "kms_acumulados": 15.5}'

echo -e "\n\n-> [POST /zapatillas] Probando validación fallida con nombre vacío (Estado esperado: 400)..."
curl -i -s -X POST "$URL/zapatillas" \
  -H "Content-Type: application/json" \
  -d '{"marca_modelo": "", "kms_acumulados": 0}'

echo -e "\n\n-> [GET /zapatillas] Listando todas las zapatillas (Estado esperado: 200)..."
curl -i -s -X GET "$URL/zapatillas"

echo -e "\n\n-> [GET /zapatillas/1] Obteniendo zapatilla con ID 1 (Estado esperado: 200)..."
curl -i -s -X GET "$URL/zapatillas/1"

echo -e "\n\n-> [PUT /zapatillas/1] Actualizando zapatilla ID 1 (Estado esperado: 200)..."
curl -i -s -X PUT "$URL/zapatillas/1" \
  -H "Content-Type: application/json" \
  -d '{"marca_modelo": "Nike Pegasus 40 Pro", "kms_acumulados": 50.0, "estado": true}'

echo "Pruebas CRUD Entrenamientos"

echo -e "\n-> [POST /entrenamientos] Creando entrenamiento con Zapatilla ID 1 (Estado esperado: 201)..."
curl -i -s -X POST "$URL/entrenamientos" \
  -H "Content-Type: application/json" \
  -d '{"fecha": "2025-05-10", "distancia_km": 10.0, "tiempo_min": 50, "tipo": "Fondo", "zapatilla_id": 1, "lugar": "Parque Independencia"}'

echo -e "\n\n-> [GET /entrenamientos] Listando entrenamientos (Estado esperado: 200)..."
curl -i -s -X GET "$URL/entrenamientos"

echo -e "\n\n-> [GET /entrenamientos/1] Obteniendo entrenamiento ID 1 (Estado esperado: 200)..."
curl -i -s -X GET "$URL/entrenamientos/1"

echo -e "\n\n-> [PUT /entrenamientos/1] Actualizando entrenamiento ID 1 (Estado esperado: 200)..."
curl -i -s -X PUT "$URL/entrenamientos/1" \
  -H "Content-Type: application/json" \
  -d '{"fecha": "2025-05-10", "distancia_km": 12.0, "tiempo_min": 60, "tipo": "Pasadas", "zapatilla_id": 1, "lugar": "Pista Municipal"}'

echo -e "\n\n-> [DELETE /entrenamientos/1] Eliminando entrenamiento ID 1 (Estado esperado: 204)..."
curl -i -s -X DELETE "$URL/entrenamientos/1"

echo -e "\n-> [GET /entrenamientos/1] Verificando eliminación de entrenamiento ID 1 (Estado esperado: 404)..."
curl -i -s -X GET "$URL/entrenamientos/1"

echo -e "\n\n-> [DELETE /zapatillas/1] Eliminando lógicamente zapatilla ID 1 (Estado esperado: 204)..."
curl -i -s -X DELETE "$URL/zapatillas/1"

echo -e "\n Todas las peticiones a la API fueron ejecutadas correctamente."