package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	db "tp1/db/sqlc"
)

type apiHandler struct {
	conn    *sql.DB
	queries *db.Queries
}

type requestError struct {
	status  int
	message string
}

func (e requestError) Error() string { return e.message }

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error escribiendo JSON: %v", err)
	}
}

func writeAPIError(w http.ResponseWriter, err error, message string) {
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

func resourceID(path string) (int32, error) {
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	if len(parts) != 3 {
		return 0, requestError{http.StatusBadRequest, "URL inválida"}
	}
	id, err := strconv.ParseInt(parts[2], 10, 32)
	if err != nil || id <= 0 {
		return 0, requestError{http.StatusBadRequest, "ID inválido"}
	}
	return int32(id), nil
}

// Todas las escrituras de una operación usan la misma transacción de sqlc.
func (h *apiHandler) withTransaction(ctx context.Context, action func(*db.Queries) error) error {
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

// El orden de los bloqueos es estable al mover un entrenamiento entre dos pares.
func lockZapatillas(ctx context.Context, q *db.Queries, ids ...int32) (map[int32]db.Zapatilla, error) {
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	result := make(map[int32]db.Zapatilla)
	for _, id := range ids {
		if _, exists := result[id]; exists {
			continue
		}
		zapatilla, err := q.GetZapatillaForUpdate(ctx, id)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, requestError{http.StatusNotFound, "Zapatilla no encontrada"}
		}
		if err != nil {
			return nil, err
		}
		result[id] = zapatilla
	}
	return result, nil
}
