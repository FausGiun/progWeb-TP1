package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	db "tp1/db/sqlc"

	_ "github.com/lib/pq"
)

func setupRouter(conn *sql.DB) http.Handler {
	app := &webHandler{conn: conn, queries: db.New(conn)}
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.entrenamientosPage)
	mux.HandleFunc("/mis-zapatillas", app.zapatillasPage)
	mux.HandleFunc("/zapatillas", app.createZapatilla)
	mux.HandleFunc("/entrenamientos", app.createEntrenamiento)
	mux.HandleFunc("/zapatillas/eliminar", app.deleteZapatilla)
	mux.HandleFunc("/entrenamientos/eliminar", app.deleteEntrenamiento)
	// El CSS sigue siendo un archivo estático; templ genera las páginas.
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

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("Falta DATABASE_URL. Configurá .env y ejecutá make run (ver README).")
	}
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Error abriendo la base de datos: %v", err)
	}
	defer conn.Close()
	if err := conn.Ping(); err != nil {
		log.Fatalf("Error conectando a la base de datos: %v", err)
	}
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("Runner Log - Tracker disponible en %s", addr)
	log.Fatal(http.ListenAndServe(addr, setupRouter(conn)))
}
