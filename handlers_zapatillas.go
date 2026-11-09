package main

import (
	"net/http"
	"strconv"
	"strings"

	db "tp1/db/sqlc"
	"tp1/logic"
	"tp1/views"
)

func (h *webHandler) zapatillasPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, "GET")
		return
	}
	zapatillas, err := h.queries.ListZapatillas(r.Context())
	if err != nil {
		writeError(w, err, "Error obteniendo zapatillas")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := views.ZapatillasPage(zapatillas).Render(r.Context(), w); err != nil {
		writeError(w, err, "Error mostrando zapatillas")
	}
}

func (h *webHandler) createZapatilla(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, "POST")
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Formulario inválido", http.StatusBadRequest)
		return
	}
	marca := strings.TrimSpace(r.FormValue("marca_modelo"))
	kilometros, err := strconv.ParseFloat(r.FormValue("kms_acumulados"), 64)
	if err != nil {
		http.Error(w, "Los kilómetros deben ser un número", http.StatusBadRequest)
		return
	}
	if err := logic.ValidateZapatilla(marca, kilometros); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = h.queries.CreateZapatilla(r.Context(), db.CreateZapatillaParams{
		MarcaModelo: marca, KmsAcumulados: kilometros, Estado: true,
	})
	if err != nil {
		writeError(w, err, "Error guardando zapatilla")
		return
	}
	http.Redirect(w, r, "/mis-zapatillas", http.StatusSeeOther)
}

func (h *webHandler) deleteZapatilla(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, "POST")
		return
	}
	id, err := readFormID(r)
	if err != nil {
		writeError(w, err, "Error leyendo zapatilla")
		return
	}
	err = h.withTransaction(r.Context(), func(q *db.Queries) error {
		if _, err := lockZapatilla(r.Context(), q, id); err != nil {
			return err
		}
		return q.DeleteZapatilla(r.Context(), id)
	})
	if err != nil {
		writeError(w, err, "Error eliminando zapatilla")
		return
	}
	http.Redirect(w, r, "/mis-zapatillas", http.StatusSeeOther)
}
