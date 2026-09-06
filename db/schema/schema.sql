CREATE TABLE zapatillas (
    id SERIAL PRIMARY KEY,
    marca_modelo VARCHAR(255) NOT NULL,
    kms_acumulados DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    estado BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE entrenamientos (
    id SERIAL PRIMARY KEY,
    fecha DATE NOT NULL,
    distancia_km DOUBLE PRECISION NOT NULL,
    tiempo_min INT NOT NULL,
    tipo VARCHAR(50) NOT NULL,
    zapatilla_id INT NOT NULL REFERENCES zapatillas(id),
    ritmo DOUBLE PRECISION NOT NULL,
    calorias INT NOT NULL DEFAULT 0,
    lugar VARCHAR(100)
);