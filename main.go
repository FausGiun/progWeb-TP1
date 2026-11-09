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
	api := &apiHandler{conn: conn, queries: db.New(conn)}
	mux := http.NewServeMux()
	mux.HandleFunc("/zapatillas", api.zapatillasHandler)
	mux.HandleFunc("/zapatillas/", api.zapatillaByIDHandler)
	mux.HandleFunc("/entrenamientos", api.entrenamientosHandler)
	mux.HandleFunc("/entrenamientos/", api.entrenamientoByIDHandler)
	// La interfaz y la API comparten origen; fetch usa rutas relativas.
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			methodNotAllowed(w, "GET, HEAD")
			return
		}
		http.ServeFile(w, r, "static/index.html")
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
	log.Printf("Running Tracker disponible en %s", addr)
	log.Fatal(http.ListenAndServe(addr, setupRouter(conn)))
}
