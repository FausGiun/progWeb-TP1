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
	paginas := map[string]string{
		"/":                   "static/index.html",
		"/mis-zapatillas":     "static/zapatillas.html",
		"/mis-entrenamientos": "static/entrenamientos.html",
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		archivo, existe := paginas[r.URL.Path]
		if !existe {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			methodNotAllowed(w, "GET, HEAD")
			return
		}
		http.ServeFile(w, r, archivo)
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
