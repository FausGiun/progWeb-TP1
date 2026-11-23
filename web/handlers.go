package web

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"tp1/application"

	"github.com/a-h/templ"
)

type webHandler struct {
	registro *application.Registro
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
	var validation application.ErrorValidacion
	if errors.As(err, &validation) {
		http.Error(w, validation.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, application.ErrZapatillaNoEncontrada) || errors.Is(err, application.ErrEntrenamientoNoEncontrado) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	log.Printf("%s: %v", message, err)
	http.Error(w, message, http.StatusInternalServerError)
}

func renderHTML(w http.ResponseWriter, r *http.Request, component templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := component.Render(r.Context(), w); err != nil {
		writeError(w, err, "Error mostrando la vista")
	}
}

func readDeleteID(r *http.Request) (int32, error) {
	if r.Method == http.MethodDelete {
		return parseID(r.PathValue("id"))
	}
	if err := r.ParseForm(); err != nil {
		return 0, requestError{http.StatusBadRequest, "Formulario inválido"}
	}
	return parseID(r.FormValue("id"))
}

func parseID(value string) (int32, error) {
	id, err := strconv.ParseInt(value, 10, 32)
	if err != nil || id <= 0 {
		return 0, requestError{http.StatusBadRequest, "ID inválido"}
	}
	return int32(id), nil
}
