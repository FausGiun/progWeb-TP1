-- name: CreateZapatilla :one
INSERT INTO zapatillas (marca_modelo, kms_acumulados, estado) 
VALUES ($1, $2, $3) 
RETURNING *;

-- name: ListZapatillas :many
SELECT * FROM zapatillas ORDER BY id;

-- name: UpdateZapatillaKms :exec
UPDATE zapatillas 
SET kms_acumulados = kms_acumulados + $2 
WHERE id = $1;

-- name: CreateEntrenamiento :one
INSERT INTO entrenamientos (fecha, distancia_km, tiempo_min, tipo, zapatilla_id, ritmo, calorias, lugar) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
RETURNING *;

-- name: GetEntrenamiento :one
SELECT * FROM entrenamientos WHERE id = $1;

-- name: ListEntrenamientos :many
SELECT * FROM entrenamientos ORDER BY fecha DESC;

-- name: DeleteEntrenamiento :exec
DELETE FROM entrenamientos WHERE id = $1;