package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	db "tp1/db/sqlc"

	_ "github.com/lib/pq"
)

func TestAPIREST_CRUDCompleto(t *testing.T) {
	dsn := "postgres://tp_user:password@localhost:5432/tp_db?sslmode=disable"
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Error abriendo conexión de prueba: %v", err)
	}
	defer conn.Close()

	// Limpiamos las tablas antes de iniciar para garantizar un estado inicial limpio
	if _, err := conn.Exec("TRUNCATE entrenamientos, zapatillas RESTART IDENTITY CASCADE;"); err != nil {
		t.Fatalf("Error limpiando tablas antes del test: %v", err)
	}

	queries = db.New(conn)
	router := setupRouter()

	var zapatillaID int32
	var entrenamientoID int32

	// 1. CICLO CRUD DE ZAPATILLAS
	t.Run("1. POST /zapatillas valido (201 Created)", func(t *testing.T) {
		body := []byte(`{"marca_modelo": "Adidas Boston 12", "kms_acumulados": 10.5}`)
		req := httptest.NewRequest(http.MethodPost, "/zapatillas", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("Se esperaba 201 Created, se obtuvo %d: %s", rr.Code, rr.Body.String())
		}

		var creada db.Zapatilla
		if err := json.NewDecoder(rr.Body).Decode(&creada); err != nil {
			t.Fatalf("Error decodificando respuesta JSON: %v", err)
		}
		if creada.MarcaModelo != "Adidas Boston 12" || !creada.Estado {
			t.Errorf("Datos incorrectos en zapatilla creada: %+v", creada)
		}
		zapatillaID = creada.ID
	})

	t.Run("2. POST /zapatillas invalido (400 Bad Request)", func(t *testing.T) {
		body := []byte(`{"marca_modelo": "", "kms_acumulados": -5}`)
		req := httptest.NewRequest(http.MethodPost, "/zapatillas", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Se esperaba 400 Bad Request, se obtuvo %d", rr.Code)
		}
	})

	t.Run("3. GET /zapatillas listar (200 OK)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/zapatillas", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Se esperaba 200 OK, se obtuvo %d", rr.Code)
		}

		var lista []db.Zapatilla
		if err := json.NewDecoder(rr.Body).Decode(&lista); err != nil {
			t.Fatalf("Error decodificando lista JSON: %v", err)
		}
		if len(lista) != 1 {
			t.Errorf("Se esperaba 1 zapatilla en la lista, se obtuvieron %d", len(lista))
		}
	})

	t.Run("4. GET /zapatillas/{id} obtener por ID (200 OK)", func(t *testing.T) {
		url := fmt.Sprintf("/zapatillas/%d", zapatillaID)
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Se esperaba 200 OK, se obtuvo %d", rr.Code)
		}
	})

	t.Run("5. PUT /zapatillas/{id} actualizar (200 OK)", func(t *testing.T) {
		url := fmt.Sprintf("/zapatillas/%d", zapatillaID)
		body := []byte(`{"marca_modelo": "Adidas Boston 12 Pro", "kms_acumulados": 25.0, "estado": true}`)
		req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Se esperaba 200 OK, se obtuvo %d: %s", rr.Code, rr.Body.String())
		}

		var actualizada db.Zapatilla
		_ = json.NewDecoder(rr.Body).Decode(&actualizada)
		if actualizada.MarcaModelo != "Adidas Boston 12 Pro" || actualizada.KmsAcumulados != 25.0 {
			t.Errorf("No se actualizaron los datos correctamente: %+v", actualizada)
		}
	})
	// 2. CICLO CRUD DE ENTRENAMIENTOS
	t.Run("6. POST /entrenamientos valido (201 Created)", func(t *testing.T) {
		payload := fmt.Sprintf(`{"fecha": "2025-05-10", "distancia_km": 10.0, "tiempo_min": 50, "tipo": "Fondo", "zapatilla_id": %d, "lugar": "Parque"}`, zapatillaID)
		req := httptest.NewRequest(http.MethodPost, "/entrenamientos", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("Se esperaba 201 Created, se obtuvo %d: %s", rr.Code, rr.Body.String())
		}

		var creado db.Entrenamiento
		if err := json.NewDecoder(rr.Body).Decode(&creado); err != nil {
			t.Fatalf("Error decodificando entrenamiento: %v", err)
		}
		if creado.Ritmo != 12.0 || creado.Calorias != 650 {
			t.Errorf("Cálculos de lógica pura incorrectos: ritmo=%v, calorias=%v", creado.Ritmo, creado.Calorias)
		}
		entrenamientoID = creado.ID
	})

	t.Run("7. PUT /entrenamientos/{id} actualizar (200 OK)", func(t *testing.T) {
		url := fmt.Sprintf("/entrenamientos/%d", entrenamientoID)
		payload := fmt.Sprintf(`{"fecha": "2025-05-10", "distancia_km": 12.0, "tiempo_min": 60, "tipo": "Pasadas", "zapatilla_id": %d, "lugar": "Pista"}`, zapatillaID)
		req := httptest.NewRequest(http.MethodPut, url, bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("Se esperaba 200 OK, se obtuvo %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("8. DELETE /entrenamientos/{id} eliminar (204 No Content)", func(t *testing.T) {
		url := fmt.Sprintf("/entrenamientos/%d", entrenamientoID)
		req := httptest.NewRequest(http.MethodDelete, url, nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusNoContent {
			t.Fatalf("Se esperaba 204 No Content, se obtuvo %d", rr.Code)
		}
	})

	t.Run("9. GET /entrenamientos/{id} eliminado (404 Not Found)", func(t *testing.T) {
		url := fmt.Sprintf("/entrenamientos/%d", entrenamientoID)
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("Se esperaba 404 Not Found, se obtuvo %d", rr.Code)
		}
	})

	t.Run("10. DELETE /zapatillas/{id} y verificar 405 Method Not Allowed", func(t *testing.T) {
		url := fmt.Sprintf("/zapatillas/%d", zapatillaID)
		reqDel := httptest.NewRequest(http.MethodDelete, url, nil)
		rrDel := httptest.NewRecorder()
		router.ServeHTTP(rrDel, reqDel)

		if rrDel.Code != http.StatusNoContent {
			t.Errorf("Se esperaba 204 No Content al borrar zapatilla, se obtuvo %d", rrDel.Code)
		}

		// Probar método no soportado (PATCH)-405 Method Not Allowed
		reqPatch := httptest.NewRequest(http.MethodPatch, "/zapatillas", nil)
		rrPatch := httptest.NewRecorder()
		router.ServeHTTP(rrPatch, reqPatch)

		if rrPatch.Code != http.StatusMethodNotAllowed {
			t.Errorf("Se esperaba 405 Method Not Allowed, se obtuvo %d", rrPatch.Code)
		}
	})
}
