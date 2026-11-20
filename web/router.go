package web

import (
	"net/http"

	"tp1/application"
)

func Router(registro *application.Registro) http.Handler {
	app := &webHandler{registro: registro}
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.entrenamientosPage)
	mux.HandleFunc("/mis-zapatillas", app.zapatillasPage)
	mux.HandleFunc("/zapatillas", app.createZapatilla)
	mux.HandleFunc("/entrenamientos", app.createEntrenamiento)
	mux.HandleFunc("POST /zapatillas/eliminar", app.deleteZapatilla)
	mux.HandleFunc("POST /entrenamientos/eliminar", app.deleteEntrenamiento)
	mux.HandleFunc("DELETE /zapatillas/{id}", app.deleteZapatilla)
	mux.HandleFunc("DELETE /entrenamientos/{id}", app.deleteEntrenamiento)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/mis-entrenamientos", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			methodNotAllowed(w, "GET")
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
	return loggingMiddleware(mux)
}
