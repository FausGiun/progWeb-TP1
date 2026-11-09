package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	db "tp1/db/sqlc"
	"tp1/logic"
)

func (h *apiHandler) zapatillasHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getZapatillas(w, r)
	case http.MethodPost:
		h.createZapatilla(w, r)
	default:
		methodNotAllowed(w, "GET, POST")
	}
}

func (h *apiHandler) zapatillaByIDHandler(w http.ResponseWriter, r *http.Request) {
	id, err := resourceID(r.URL.Path)
	if err != nil {
		writeAPIError(w, err, "ID de zapatilla inválido")
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.getZapatilla(w, r, id)
	case http.MethodPut:
		h.updateZapatilla(w, r, id)
	case http.MethodDelete:
		h.deleteZapatilla(w, r, id)
	default:
		methodNotAllowed(w, "GET, PUT, DELETE")
	}
}

func (h *apiHandler) getZapatillas(w http.ResponseWriter, r *http.Request) {
	zapatillas, err := h.queries.ListZapatillas(r.Context())
	if err != nil {
		writeAPIError(w, err, "Error obteniendo zapatillas")
		return
	}
	if zapatillas == nil {
		zapatillas = []db.Zapatilla{}
	}
	writeJSON(w, http.StatusOK, zapatillas)
}

func (h *apiHandler) createZapatilla(w http.ResponseWriter, r *http.Request) {
	var input db.CreateZapatillaParams
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	input.MarcaModelo = strings.TrimSpace(input.MarcaModelo)
	if err := logic.ValidateZapatilla(input.MarcaModelo, input.KmsAcumulados); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	input.Estado = true
	creada, err := h.queries.CreateZapatilla(r.Context(), input)
	if err != nil {
		writeAPIError(w, err, "Error guardando zapatilla")
		return
	}
	writeJSON(w, http.StatusCreated, creada)
}

func (h *apiHandler) getZapatilla(w http.ResponseWriter, r *http.Request, id int32) {
	zapatilla, err := h.queries.GetZapatilla(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Zapatilla no encontrada", http.StatusNotFound)
		return
	}
	if err != nil {
		writeAPIError(w, err, "Error buscando zapatilla")
		return
	}
	writeJSON(w, http.StatusOK, zapatilla)
}

func (h *apiHandler) updateZapatilla(w http.ResponseWriter, r *http.Request, id int32) {
	var input db.UpdateZapatillaParams
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	input.MarcaModelo = strings.TrimSpace(input.MarcaModelo)
	if err := logic.ValidateZapatilla(input.MarcaModelo, input.KmsAcumulados); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	input.ID = id
	var actualizada db.Zapatilla
	err := h.withTransaction(r.Context(), func(q *db.Queries) error {
		if _, err := lockZapatillas(r.Context(), q, id); err != nil {
			return err
		}
		registrados, err := q.GetKilometrosRegistrados(r.Context(), id)
		if err != nil {
			return err
		}
		if input.KmsAcumulados+0.000001 < registrados {
			return requestError{http.StatusBadRequest, "Los kilómetros no pueden ser menores a los entrenamientos registrados"}
		}
		actualizada, err = q.UpdateZapatilla(r.Context(), input)
		return err
	})
	if err != nil {
		writeAPIError(w, err, "Error actualizando zapatilla")
		return
	}
	writeJSON(w, http.StatusOK, actualizada)
}

func (h *apiHandler) deleteZapatilla(w http.ResponseWriter, r *http.Request, id int32) {
	err := h.withTransaction(r.Context(), func(q *db.Queries) error {
		if _, err := lockZapatillas(r.Context(), q, id); err != nil {
			return err
		}
		return q.DeleteZapatilla(r.Context(), id)
	})
	if err != nil {
		writeAPIError(w, err, "Error eliminando zapatilla")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
