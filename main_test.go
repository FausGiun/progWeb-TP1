package main

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	db "tp1/db/sqlc"

	_ "github.com/lib/pq"
)

func TestAPIREST(t *testing.T) {
	conn, err := sql.Open("postgres", "postgres://tp_user:password@localhost:5432/tp_db?sslmode=disable")
	if err != nil {
		t.Fatalf("Error conectando a la BD de prueba: %v", err)
	}
	defer conn.Close()

	queries = db.New(conn)
	router := setupRouter()

	// 1. Probar GET /zapatillas (200 OK)
	t.Run("GET /zapatillas", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/zapatillas", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Se esperaba 200 OK, se obtuvo %v", rr.Code)
		}
	})

	// 2. Probar POST /zapatillas válido (201 Created)
	t.Run("POST /zapatillas valido", func(t *testing.T) {
		body := []byte(`{"marca_modelo": "Adidas Boston 12", "kms_acumulados": 10}`)
		req, _ := http.NewRequest(http.MethodPost, "/zapatillas", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("Se esperaba 201 Created, se obtuvo %v", rr.Code)
		}
	})

	// 3. Probar POST /zapatillas inválido (400 Bad Request)
	t.Run("POST /zapatillas invalido", func(t *testing.T) {
		body := []byte(`{"marca_modelo": "", "kms_acumulados": -5}`)
		req, _ := http.NewRequest(http.MethodPost, "/zapatillas", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Se esperaba 400 Bad Request, se obtuvo %v", rr.Code)
		}
	})
}
