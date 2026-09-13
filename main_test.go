package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	db "tp1/db/sqlc"

	_ "github.com/lib/pq"
)

func TestRutasGET(t *testing.T) {
	// 1. Conectamos a la BD de prueba que levantó el Makefile
	conn, err := sql.Open("postgres", "postgres://tp_user:password@localhost:5432/tp_db?sslmode=disable")
	if err != nil {
		t.Fatalf("Error conectando a la BD de prueba: %v", err)
	}
	defer conn.Close()
	dbQueries = db.New(conn)

	rutas := []struct {
		nombre  string
		url     string
		handler http.HandlerFunc //del main.go
	}{
		{"Inicio", "/", rootHandler},
		{"Nueva Zapatilla", "/nueva-zapatilla", nuevaZapatillaHandler},
		{"Historial", "/historial", historialHandler},
		{"Nuevo Entrenamiento", "/nuevo-entrenamiento", registrarEjHandler},
	}

	for _, ruta := range rutas {
		t.Run(ruta.nombre, func(t *testing.T) {
			req, err := http.NewRequest("GET", ruta.url, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()

			// Ejecutamos el handler
			ruta.handler.ServeHTTP(rr, req)
			if status := rr.Code; status != http.StatusOK {
				t.Errorf("La ruta %s devolvió el estado %v; se esperaba %v", ruta.url, status, http.StatusOK)
			}
		})
	}
}
