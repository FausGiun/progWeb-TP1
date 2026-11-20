package postgres

import (
	"context"
	"database/sql"
	"errors"

	"tp1/application"
	db "tp1/db/sqlc"
	"tp1/logic"
)

type Repositorio struct {
	conn    *sql.DB
	queries *db.Queries
}

func NuevoRepositorio(conn *sql.DB) *Repositorio {
	return &Repositorio{conn: conn, queries: db.New(conn)}
}

func (r *Repositorio) ListarZapatillas(ctx context.Context) ([]logic.Zapatilla, error) {
	filas, err := r.queries.ListZapatillas(ctx)
	if err != nil {
		return nil, err
	}
	zapatillas := make([]logic.Zapatilla, len(filas))
	for i, fila := range filas {
		zapatillas[i] = zapatillaDeFila(fila)
	}
	return zapatillas, nil
}

func (r *Repositorio) ListarEntrenamientos(ctx context.Context) ([]logic.Entrenamiento, error) {
	filas, err := r.queries.ListEntrenamientos(ctx)
	if err != nil {
		return nil, err
	}
	entrenamientos := make([]logic.Entrenamiento, len(filas))
	for i, fila := range filas {
		entrenamientos[i] = entrenamientoDeFila(fila)
	}
	return entrenamientos, nil
}

func (r *Repositorio) GuardarZapatilla(ctx context.Context, zapatilla logic.Zapatilla) error {
	_, err := r.queries.CreateZapatilla(ctx, db.CreateZapatillaParams{
		MarcaModelo: zapatilla.MarcaModelo, KmsAcumulados: zapatilla.KmsAcumulados, Estado: zapatilla.Estado,
	})
	return err
}

func (r *Repositorio) EnTransaccion(ctx context.Context, operacion func(application.Cambios) error) error {
	tx, err := r.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := operacion(cambiosPostgres{queries: r.queries.WithTx(tx)}); err != nil {
		return err
	}
	return tx.Commit()
}

type cambiosPostgres struct {
	queries *db.Queries
}

func (c cambiosPostgres) BuscarZapatillaParaActualizar(ctx context.Context, id int32) (logic.Zapatilla, error) {
	fila, err := c.queries.GetZapatillaForUpdate(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return logic.Zapatilla{}, application.ErrZapatillaNoEncontrada
	}
	return zapatillaDeFila(fila), err
}

func (c cambiosPostgres) BuscarEntrenamientoParaActualizar(ctx context.Context, id int32) (logic.Entrenamiento, error) {
	fila, err := c.queries.GetEntrenamientoForUpdate(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return logic.Entrenamiento{}, application.ErrEntrenamientoNoEncontrado
	}
	return entrenamientoDeFila(fila), err
}

func (c cambiosPostgres) GuardarEntrenamiento(ctx context.Context, entrenamiento logic.Entrenamiento) error {
	var lugar sql.NullString
	if entrenamiento.Lugar != nil {
		lugar = sql.NullString{String: *entrenamiento.Lugar, Valid: true}
	}
	_, err := c.queries.CreateEntrenamiento(ctx, db.CreateEntrenamientoParams{
		Fecha: entrenamiento.Fecha, DistanciaKm: entrenamiento.DistanciaKm, TiempoMin: entrenamiento.TiempoMin,
		Tipo: entrenamiento.Tipo, ZapatillaID: entrenamiento.ZapatillaID,
		Ritmo: entrenamiento.Ritmo, Calorias: entrenamiento.Calorias, Lugar: lugar,
	})
	return err
}

func (c cambiosPostgres) CambiarKilometros(ctx context.Context, id int32, distancia float64) error {
	return c.queries.UpdateZapatillaKms(ctx, db.UpdateZapatillaKmsParams{ID: id, KmsAcumulados: distancia})
}

func (c cambiosPostgres) RetirarZapatilla(ctx context.Context, id int32) error {
	return c.queries.DeleteZapatilla(ctx, id)
}

func (c cambiosPostgres) BorrarEntrenamiento(ctx context.Context, id int32) error {
	return c.queries.DeleteEntrenamiento(ctx, id)
}

func zapatillaDeFila(fila db.Zapatilla) logic.Zapatilla {
	return logic.Zapatilla{
		ID: fila.ID, MarcaModelo: fila.MarcaModelo, KmsAcumulados: fila.KmsAcumulados, Estado: fila.Estado,
	}
}

func entrenamientoDeFila(fila db.Entrenamiento) logic.Entrenamiento {
	entrenamiento := logic.Entrenamiento{
		ID: fila.ID, Fecha: fila.Fecha, DistanciaKm: fila.DistanciaKm, TiempoMin: fila.TiempoMin,
		Tipo: fila.Tipo, ZapatillaID: fila.ZapatillaID, Ritmo: fila.Ritmo, Calorias: fila.Calorias,
	}
	if fila.Lugar.Valid {
		entrenamiento.Lugar = &fila.Lugar.String
	}
	return entrenamiento
}
