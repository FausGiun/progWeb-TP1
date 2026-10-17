-- Consultas para Zapatillas

-- name: CreateZapatilla :one
INSERT INTO zapatillas (marca_modelo, kms_acumulados, estado)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListZapatillas :many
SELECT * FROM zapatillas
WHERE estado = true
ORDER BY id ASC;

-- name: GetZapatilla :one
SELECT * FROM zapatillas
WHERE id = $1 LIMIT 1;

-- name: UpdateZapatilla :one
UPDATE zapatillas
SET marca_modelo = $2, kms_acumulados = $3, estado = $4
WHERE id = $1
RETURNING *;

-- name: UpdateZapatillaKms :exec
UPDATE zapatillas
SET kms_acumulados = kms_acumulados + $2
WHERE id = $1;

-- name: DeleteZapatilla :exec
UPDATE zapatillas
SET estado = false
WHERE id = $1;

-- CONSULTAS PARA ENTRENAMIENTOS

-- name: CreateEntrenamiento :one
INSERT INTO entrenamientos (
    fecha, distancia_km, tiempo_min, tipo, zapatilla_id, ritmo, calorias, lugar
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: ListEntrenamientos :many
SELECT * FROM entrenamientos
ORDER BY fecha DESC, id DESC;

-- name: GetEntrenamiento :one
SELECT * FROM entrenamientos
WHERE id = $1 LIMIT 1;

-- name: UpdateEntrenamiento :one
UPDATE entrenamientos
SET fecha = $2, distancia_km = $3, tiempo_min = $4, tipo = $5,
    zapatilla_id = $6, ritmo = $7, calorias = $8, lugar = $9
WHERE id = $1
RETURNING *;

-- name: DeleteEntrenamiento :exec
DELETE FROM entrenamientos
WHERE id = $1;