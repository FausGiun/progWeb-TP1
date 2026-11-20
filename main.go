package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"tp1/application"
	"tp1/db/postgres"
	"tp1/web"

	_ "github.com/lib/pq"
)

func setupRouter(conn *sql.DB) http.Handler {
	repo := postgres.NuevoRepositorio(conn)
	registro := application.NuevoRegistro(repo)
	return web.Router(registro)
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
