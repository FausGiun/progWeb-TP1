package web

import (
	"net/http"
	"strconv"

	"tp1/views"
)

func (h *webHandler) zapatillasPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, "GET")
		return
	}
	h.renderZapatillas(w, r, false)
}

func (h *webHandler) renderZapatillas(w http.ResponseWriter, r *http.Request, soloLista bool) {
	zapatillas, err := h.registro.ListarZapatillas(r.Context())
	if err != nil {
		writeError(w, err, "Error obteniendo zapatillas")
		return
	}
	if soloLista {
		renderHTML(w, r, views.ZapatillasList(zapatillas))
		return
	}
	renderHTML(w, r, views.ZapatillasPage(zapatillas))
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
	marca := r.FormValue("marca_modelo")
	kilometros, err := strconv.ParseFloat(r.FormValue("kms_acumulados"), 64)
	if err != nil {
		http.Error(w, "Los kilómetros deben ser un número", http.StatusBadRequest)
		return
	}
	if err := h.registro.AgregarZapatilla(r.Context(), marca, kilometros); err != nil {
		writeError(w, err, "Error guardando zapatilla")
		return
	}
	if r.Header.Get("HX-Request") == "true" {
		h.renderZapatillas(w, r, true)
		return
	}
	http.Redirect(w, r, "/mis-zapatillas", http.StatusSeeOther)
}

func (h *webHandler) deleteZapatilla(w http.ResponseWriter, r *http.Request) {
	id, err := readDeleteID(r)
	if err != nil {
		writeError(w, err, "Error leyendo zapatilla")
		return
	}
	if err := h.registro.RetirarZapatilla(r.Context(), id); err != nil {
		writeError(w, err, "Error eliminando zapatilla")
		return
	}
	if r.Method == http.MethodDelete {
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/mis-zapatillas", http.StatusSeeOther)
}
