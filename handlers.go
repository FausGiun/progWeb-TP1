package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	db "tp1/db/sqlc"
)

type webHandler struct {
	conn    *sql.DB
	queries *db.Queries
}

type requestError struct {
	status  int
	message string
}

func (e requestError) Error() string { return e.message }

func writeError(w http.ResponseWriter, err error, message string) {
	var inputError requestError
	if errors.As(err, &inputError) {
		http.Error(w, inputError.message, inputError.status)
		return
	}
	log.Printf("%s: %v", message, err)
	http.Error(w, message, http.StatusInternalServerError)
}

func methodNotAllowed(w http.ResponseWriter, methods string) {
	w.Header().Set("Allow", methods)
	http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
}

func readFormID(r *http.Request) (int32, error) {
	if err := r.ParseForm(); err != nil {
		return 0, requestError{http.StatusBadRequest, "Formulario inválido"}
	}
	id, err := strconv.ParseInt(r.FormValue("id"), 10, 32)
	if err != nil || id <= 0 {
		return 0, requestError{http.StatusBadRequest, "ID inválido"}
	}
	return int32(id), nil
}

// Conserva la operación de entrenamiento y kilometraje en una sola transacción.
func (h *webHandler) withTransaction(ctx context.Context, action func(*db.Queries) error) error {
	tx, err := h.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := action(h.queries.WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit()
}

func lockZapatilla(ctx context.Context, q *db.Queries, id int32) (db.Zapatilla, error) {
	zapatilla, err := q.GetZapatillaForUpdate(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return db.Zapatilla{}, requestError{http.StatusNotFound, "Zapatilla no encontrada"}
	}
	return zapatilla, err
}
