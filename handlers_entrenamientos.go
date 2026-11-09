package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	db "tp1/db/sqlc"
	"tp1/logic"
	"tp1/views"
)

func (h *webHandler) entrenamientosPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		methodNotAllowed(w, "GET")
		return
	}
	entrenamientos, err := h.queries.ListEntrenamientos(r.Context())
	if err != nil {
		writeError(w, err, "Error obteniendo entrenamientos")
		return
	}
	zapatillas, err := h.queries.ListZapatillas(r.Context())
	if err != nil {
		writeError(w, err, "Error obteniendo zapatillas")
		return
	}
	for i := range entrenamientos {
		entrenamientos[i].Ritmo = logic.CalcularRitmo(entrenamientos[i].DistanciaKm, int(entrenamientos[i].TiempoMin))
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.IndexPage(entrenamientos, zapatillas, time.Now().Format("2006-01-02")).Render(r.Context(), w); err != nil {
		writeError(w, err, "Error mostrando entrenamientos")
	}
}

func readEntrenamiento(r *http.Request) (db.CreateEntrenamientoParams, error) {
	if err := r.ParseForm(); err != nil {
		return db.CreateEntrenamientoParams{}, requestError{http.StatusBadRequest, "Formulario inválido"}
	}
	fecha, err := time.Parse("2006-01-02", r.FormValue("fecha"))
	if err != nil {
		return db.CreateEntrenamientoParams{}, requestError{http.StatusBadRequest, "Fecha inválida (use YYYY-MM-DD)"}
	}
	distancia, errDistancia := strconv.ParseFloat(r.FormValue("distancia_km"), 64)
	tiempo, errTiempo := strconv.Atoi(r.FormValue("tiempo_min"))
	zapatillaID, errZapatilla := strconv.Atoi(r.FormValue("zapatilla_id"))
	if errDistancia != nil || errTiempo != nil || errZapatilla != nil {
		return db.CreateEntrenamientoParams{}, requestError{http.StatusBadRequest, "Distancia, tiempo y zapatilla deben contener números válidos"}
	}
	tipo := strings.TrimSpace(r.FormValue("tipo"))
	lugar := strings.TrimSpace(r.FormValue("lugar"))
	if err := logic.ValidateEntrenamiento(fecha, distancia, tiempo, tipo, zapatillaID); err != nil {
		return db.CreateEntrenamientoParams{}, requestError{http.StatusBadRequest, err.Error()}
	}
	return db.CreateEntrenamientoParams{
		Fecha: fecha, DistanciaKm: distancia, TiempoMin: int32(tiempo),
		Tipo: tipo, ZapatillaID: int32(zapatillaID),
		Ritmo: logic.CalcularRitmo(distancia, tiempo), Calorias: logic.CalcularCalorias(distancia),
		Lugar: sql.NullString{String: lugar, Valid: lugar != ""},
	}, nil
}

func (h *webHandler) createEntrenamiento(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, "POST")
		return
	}
	input, err := readEntrenamiento(r)
	if err != nil {
		writeError(w, err, "Error leyendo entrenamiento")
		return
	}
	err = h.withTransaction(r.Context(), func(q *db.Queries) error {
		zapatilla, err := lockZapatilla(r.Context(), q, input.ZapatillaID)
		if err != nil {
			return err
		}
		if !zapatilla.Estado {
			return requestError{http.StatusBadRequest, "La zapatilla seleccionada ya no está activa"}
		}
		if _, err := q.CreateEntrenamiento(r.Context(), input); err != nil {
			return err
		}
		return q.UpdateZapatillaKms(r.Context(), db.UpdateZapatillaKmsParams{
			ID: input.ZapatillaID, KmsAcumulados: input.DistanciaKm,
		})
	})
	if err != nil {
		writeError(w, err, "Error guardando entrenamiento")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *webHandler) deleteEntrenamiento(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, "POST")
		return
	}
	id, err := readFormID(r)
	if err != nil {
		writeError(w, err, "Error leyendo entrenamiento")
		return
	}
	err = h.withTransaction(r.Context(), func(q *db.Queries) error {
		anterior, err := q.GetEntrenamientoForUpdate(r.Context(), id)
		if errors.Is(err, sql.ErrNoRows) {
			return requestError{http.StatusNotFound, "Entrenamiento no encontrado"}
		}
		if err != nil {
			return err
		}
		if _, err := lockZapatilla(r.Context(), q, anterior.ZapatillaID); err != nil {
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
		writeError(w, err, "Error eliminando entrenamiento")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
