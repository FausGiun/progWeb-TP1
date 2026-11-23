package web

import (
	"net/http"
	"strconv"
	"time"

	"tp1/application"
	"tp1/views"
)

func (h *webHandler) entrenamientosPage(w http.ResponseWriter, r *http.Request) {
	h.renderEntrenamientos(w, r, false)
}

func (h *webHandler) renderEntrenamientos(w http.ResponseWriter, r *http.Request, soloLista bool) {
	entrenamientos, err := h.registro.ListarEntrenamientos(r.Context())
	if err != nil {
		writeError(w, err, "Error obteniendo entrenamientos")
		return
	}
	zapatillas, err := h.registro.ListarZapatillas(r.Context())
	if err != nil {
		writeError(w, err, "Error obteniendo zapatillas")
		return
	}
	if soloLista {
		renderHTML(w, r, views.EntrenamientosList(entrenamientos, zapatillas))
		return
	}
	renderHTML(w, r, views.IndexPage(entrenamientos, zapatillas, time.Now().Format("2006-01-02")))
}

func readEntrenamiento(r *http.Request) (application.NuevoEntrenamiento, error) {
	if err := r.ParseForm(); err != nil {
		return application.NuevoEntrenamiento{}, requestError{http.StatusBadRequest, "Formulario inválido"}
	}
	fecha, err := time.Parse("2006-01-02", r.FormValue("fecha"))
	if err != nil {
		return application.NuevoEntrenamiento{}, requestError{http.StatusBadRequest, "Fecha inválida (use YYYY-MM-DD)"}
	}
	distancia, errDistancia := strconv.ParseFloat(r.FormValue("distancia_km"), 64)
	tiempo, errTiempo := strconv.Atoi(r.FormValue("tiempo_min"))
	zapatillaID, errZapatilla := strconv.Atoi(r.FormValue("zapatilla_id"))
	if errDistancia != nil || errTiempo != nil || errZapatilla != nil {
		return application.NuevoEntrenamiento{}, requestError{http.StatusBadRequest, "Distancia, tiempo y zapatilla deben contener números válidos"}
	}
	return application.NuevoEntrenamiento{
		Fecha: fecha, DistanciaKm: distancia, TiempoMin: tiempo,
		Tipo: r.FormValue("tipo"), ZapatillaID: zapatillaID, Lugar: r.FormValue("lugar"),
	}, nil
}

func (h *webHandler) createEntrenamiento(w http.ResponseWriter, r *http.Request) {
	input, err := readEntrenamiento(r)
	if err != nil {
		writeError(w, err, "Error leyendo entrenamiento")
		return
	}
	if err := h.registro.RegistrarEntrenamiento(r.Context(), input); err != nil {
		writeError(w, err, "Error guardando entrenamiento")
		return
	}
	if r.Header.Get("HX-Request") == "true" {
		h.renderEntrenamientos(w, r, true)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *webHandler) deleteEntrenamiento(w http.ResponseWriter, r *http.Request) {
	id, err := readDeleteID(r)
	if err != nil {
		writeError(w, err, "Error leyendo entrenamiento")
		return
	}
	if err := h.registro.EliminarEntrenamiento(r.Context(), id); err != nil {
		writeError(w, err, "Error eliminando entrenamiento")
		return
	}
	if r.Method == http.MethodDelete {
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
