package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	db "tp1/db/sqlc"
	"tp1/logic"
)

type EntrenamientoRequest struct {
	Fecha       string  `json:"fecha"`
	DistanciaKm float64 `json:"distancia_km"`
	TiempoMin   int     `json:"tiempo_min"`
	Tipo        string  `json:"tipo"`
	ZapatillaID int     `json:"zapatilla_id"`
	Lugar       string  `json:"lugar"`
}

func (h *apiHandler) entrenamientosHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getEntrenamientos(w, r)
	case http.MethodPost:
		h.createEntrenamiento(w, r)
	default:
		methodNotAllowed(w, "GET, POST")
	}
}

func (h *apiHandler) entrenamientoByIDHandler(w http.ResponseWriter, r *http.Request) {
	id, err := resourceID(r.URL.Path)
	if err != nil {
		writeAPIError(w, err, "ID de entrenamiento inválido")
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.getEntrenamiento(w, r, id)
	case http.MethodPut:
		h.updateEntrenamiento(w, r, id)
	case http.MethodDelete:
		h.deleteEntrenamiento(w, r, id)
	default:
		methodNotAllowed(w, "GET, PUT, DELETE")
	}
}

func parseFecha(value string) (time.Time, error) {
	if fecha, err := time.Parse("2006-01-02", value); err == nil {
		return fecha, nil
	}
	return time.Parse(time.RFC3339, value)
}

func readEntrenamiento(r *http.Request) (db.CreateEntrenamientoParams, error) {
	var input EntrenamientoRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		return db.CreateEntrenamientoParams{}, requestError{http.StatusBadRequest, "JSON inválido"}
	}
	fecha, err := parseFecha(input.Fecha)
	if err != nil {
		return db.CreateEntrenamientoParams{}, requestError{http.StatusBadRequest, "Fecha inválida (use YYYY-MM-DD)"}
	}
	input.Tipo = strings.TrimSpace(input.Tipo)
	input.Lugar = strings.TrimSpace(input.Lugar)
	if err := logic.ValidateEntrenamiento(fecha, input.DistanciaKm, input.TiempoMin, input.Tipo, input.ZapatillaID); err != nil {
		return db.CreateEntrenamientoParams{}, requestError{http.StatusBadRequest, err.Error()}
	}
	return db.CreateEntrenamientoParams{
		Fecha: fecha, DistanciaKm: input.DistanciaKm, TiempoMin: int32(input.TiempoMin),
		Tipo: input.Tipo, ZapatillaID: int32(input.ZapatillaID),
		Ritmo:    logic.CalcularRitmo(input.DistanciaKm, input.TiempoMin),
		Calorias: logic.CalcularCalorias(input.DistanciaKm),
		Lugar:    sql.NullString{String: input.Lugar, Valid: input.Lugar != ""},
	}, nil
}

func (h *apiHandler) getEntrenamientos(w http.ResponseWriter, r *http.Request) {
	entrenamientos, err := h.queries.ListEntrenamientos(r.Context())
	if err != nil {
		writeAPIError(w, err, "Error obteniendo entrenamientos")
		return
	}
	if entrenamientos == nil {
		entrenamientos = []db.Entrenamiento{}
	}
	// El ritmo se deriva de distancia y tiempo, también en registros del TP3.
	for i := range entrenamientos {
		entrenamientos[i].Ritmo = logic.CalcularRitmo(entrenamientos[i].DistanciaKm, int(entrenamientos[i].TiempoMin))
	}
	writeJSON(w, http.StatusOK, entrenamientos)
}

func (h *apiHandler) getEntrenamiento(w http.ResponseWriter, r *http.Request, id int32) {
	entrenamiento, err := h.queries.GetEntrenamiento(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Entrenamiento no encontrado", http.StatusNotFound)
		return
	}
	if err != nil {
		writeAPIError(w, err, "Error obteniendo entrenamiento")
		return
	}
	entrenamiento.Ritmo = logic.CalcularRitmo(entrenamiento.DistanciaKm, int(entrenamiento.TiempoMin))
	writeJSON(w, http.StatusOK, entrenamiento)
}

func (h *apiHandler) createEntrenamiento(w http.ResponseWriter, r *http.Request) {
	input, err := readEntrenamiento(r)
	if err != nil {
		writeAPIError(w, err, "Error leyendo entrenamiento")
		return
	}
	var creado db.Entrenamiento
	err = h.withTransaction(r.Context(), func(q *db.Queries) error {
		zapatillas, err := lockZapatillas(r.Context(), q, input.ZapatillaID)
		if err != nil {
			return err
		}
		if !zapatillas[input.ZapatillaID].Estado {
			return requestError{http.StatusBadRequest, "La zapatilla seleccionada ya no está activa"}
		}
		creado, err = q.CreateEntrenamiento(r.Context(), input)
		if err != nil {
			return err
		}
		return q.UpdateZapatillaKms(r.Context(), db.UpdateZapatillaKmsParams{
			ID: input.ZapatillaID, KmsAcumulados: input.DistanciaKm,
		})
	})
	if err != nil {
		writeAPIError(w, err, "Error guardando entrenamiento")
		return
	}
	writeJSON(w, http.StatusCreated, creado)
}

func (h *apiHandler) updateEntrenamiento(w http.ResponseWriter, r *http.Request, id int32) {
	input, err := readEntrenamiento(r)
	if err != nil {
		writeAPIError(w, err, "Error leyendo entrenamiento")
		return
	}
	var actualizado db.Entrenamiento
	err = h.withTransaction(r.Context(), func(q *db.Queries) error {
		anterior, err := q.GetEntrenamientoForUpdate(r.Context(), id)
		if errors.Is(err, sql.ErrNoRows) {
			return requestError{http.StatusNotFound, "Entrenamiento no encontrado"}
		}
		if err != nil {
			return err
		}
		zapatillas, err := lockZapatillas(r.Context(), q, anterior.ZapatillaID, input.ZapatillaID)
		if err != nil {
			return err
		}
		if input.ZapatillaID != anterior.ZapatillaID && !zapatillas[input.ZapatillaID].Estado {
			return requestError{http.StatusBadRequest, "La zapatilla seleccionada ya no está activa"}
		}
		actualizado, err = q.UpdateEntrenamiento(r.Context(), db.UpdateEntrenamientoParams{
			ID: id, Fecha: input.Fecha, DistanciaKm: input.DistanciaKm, TiempoMin: input.TiempoMin,
			Tipo: input.Tipo, ZapatillaID: input.ZapatillaID, Ritmo: input.Ritmo,
			Calorias: input.Calorias, Lugar: input.Lugar,
		})
		if err != nil {
			return err
		}
		if err := q.UpdateZapatillaKms(r.Context(), db.UpdateZapatillaKmsParams{
			ID: anterior.ZapatillaID, KmsAcumulados: -anterior.DistanciaKm,
		}); err != nil {
			return err
		}
		return q.UpdateZapatillaKms(r.Context(), db.UpdateZapatillaKmsParams{
			ID: input.ZapatillaID, KmsAcumulados: input.DistanciaKm,
		})
	})
	if err != nil {
		writeAPIError(w, err, "Error actualizando entrenamiento")
		return
	}
	writeJSON(w, http.StatusOK, actualizado)
}

func (h *apiHandler) deleteEntrenamiento(w http.ResponseWriter, r *http.Request, id int32) {
	err := h.withTransaction(r.Context(), func(q *db.Queries) error {
		anterior, err := q.GetEntrenamientoForUpdate(r.Context(), id)
		if errors.Is(err, sql.ErrNoRows) {
			return requestError{http.StatusNotFound, "Entrenamiento no encontrado"}
		}
		if err != nil {
			return err
		}
		if _, err := lockZapatillas(r.Context(), q, anterior.ZapatillaID); err != nil {
			return err
		}
		if err := q.DeleteEntrenamiento(r.Context(), id); err != nil {
			return err
		}
		return q.UpdateZapatillaKms(r.Context(), db.UpdateZapatillaKmsParams{
			ID: anterior.ZapatillaID, KmsAcumulados: -anterior.DistanciaKm,
		})
	})
	if err != nil {
		writeAPIError(w, err, "Error eliminando entrenamiento")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
