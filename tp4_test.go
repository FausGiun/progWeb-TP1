package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	db "tp1/db/sqlc"
)

func TestInterfazYArchivosEstaticos(t *testing.T) {
	router := setupRouter(nil)
	for _, ruta := range []string{"/", "/static/app.js", "/static/styles.css"} {
		respuesta := httptest.NewRecorder()
		router.ServeHTTP(respuesta, httptest.NewRequest(http.MethodGet, ruta, nil))
		if respuesta.Code != http.StatusOK {
			t.Errorf("GET %s = %d; esperado 200", ruta, respuesta.Code)
		}
	}
	respuesta := httptest.NewRecorder()
	router.ServeHTTP(respuesta, httptest.NewRequest(http.MethodGet, "/no-existe", nil))
	if respuesta.Code != http.StatusNotFound {
		t.Errorf("Ruta inexistente = %d; esperado 404", respuesta.Code)
	}
}

func TestKilometrajeYTransacciones(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Fatal("Falta TEST_DATABASE_URL. Ejecutá make test.")
	}
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, err := conn.Exec("TRUNCATE entrenamientos, zapatillas RESTART IDENTITY CASCADE"); err != nil {
		t.Fatal(err)
	}
	q := db.New(conn)
	router := setupRouter(conn)
	ctx := context.Background()
	peticion := func(metodo, ruta, cuerpo string, esperado int) *httptest.ResponseRecorder {
		t.Helper()
		respuesta := httptest.NewRecorder()
		req := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(respuesta, req)
		if respuesta.Code != esperado {
			t.Fatalf("%s %s = %d, esperado %d: %s", metodo, ruta, respuesta.Code, esperado, respuesta.Body.String())
		}
		return respuesta
	}
	comprobarKms := func(id int32, esperado float64) {
		t.Helper()
		zapatilla, err := q.GetZapatilla(ctx, id)
		if err != nil || math.Abs(zapatilla.KmsAcumulados-esperado) > 0.000001 {
			t.Fatalf("Zapatilla %d: km=%v, esperado %v, error=%v", id, zapatilla.KmsAcumulados, esperado, err)
		}
	}
	primera, err := q.CreateZapatilla(ctx, db.CreateZapatillaParams{MarcaModelo: "Primera", KmsAcumulados: 20, Estado: true})
	if err != nil {
		t.Fatal(err)
	}
	segunda, err := q.CreateZapatilla(ctx, db.CreateZapatillaParams{MarcaModelo: "Segunda", KmsAcumulados: 3, Estado: true})
	if err != nil {
		t.Fatal(err)
	}
	cuerpo := func(km float64, zapatillaID int32) string {
		return fmt.Sprintf(`{"fecha":"2025-05-10","distancia_km":%v,"tiempo_min":50,"tipo":"Fondo","zapatilla_id":%d}`, km, zapatillaID)
	}
	respuesta := peticion("POST", "/entrenamientos", cuerpo(10, primera.ID), 201)
	var entrenamiento db.Entrenamiento
	if err := json.Unmarshal(respuesta.Body.Bytes(), &entrenamiento); err != nil {
		t.Fatal(err)
	}
	comprobarKms(primera.ID, 30)
	ruta := fmt.Sprintf("/entrenamientos/%d", entrenamiento.ID)
	peticion("PUT", ruta, cuerpo(12, primera.ID), 200)
	comprobarKms(primera.ID, 32)
	peticion("PUT", ruta, cuerpo(7, segunda.ID), 200)
	comprobarKms(primera.ID, 20)
	comprobarKms(segunda.ID, 10)

	// Una zapatilla inexistente no modifica ni el entrenamiento ni los kilómetros.
	peticion("PUT", ruta, cuerpo(50, 999999), 404)
	comprobarKms(segunda.ID, 10)
	actual, err := q.GetEntrenamiento(ctx, entrenamiento.ID)
	if err != nil || actual.ZapatillaID != segunda.ID || actual.DistanciaKm != 7 {
		t.Fatalf("Se modificó un entrenamiento rechazado: %+v, %v", actual, err)
	}
	peticion("PUT", fmt.Sprintf("/zapatillas/%d", segunda.ID), `{"marca_modelo":"Segunda","kms_acumulados":1,"estado":true}`, 400)
	comprobarKms(segunda.ID, 10)
	peticion("DELETE", ruta, "", 204)
	comprobarKms(segunda.ID, 3)
	peticion("DELETE", ruta, "", 404)
	comprobarKms(segunda.ID, 3)
	peticion("DELETE", fmt.Sprintf("/zapatillas/%d", segunda.ID), "", 204)
	peticion("POST", "/entrenamientos", cuerpo(5, segunda.ID), 400)
	comprobarKms(segunda.ID, 3)

	// Una falla al actualizar kilómetros debe revertir también el INSERT previo.
	if _, err := conn.Exec(`ALTER TABLE zapatillas ADD CONSTRAINT tp4_test_maxkms CHECK (kms_acumulados <= 25)`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := conn.Exec("ALTER TABLE zapatillas DROP CONSTRAINT tp4_test_maxkms"); err != nil {
			t.Errorf("No se pudo limpiar la restricción de prueba: %v", err)
		}
	}()
	peticion("POST", "/entrenamientos", cuerpo(10, primera.ID), 500)
	comprobarKms(primera.ID, 20)
	registros, err := q.ListEntrenamientos(ctx)
	if err != nil || len(registros) != 0 {
		t.Fatalf("Se guardó parte de una operación fallida: %d registros, error=%v", len(registros), err)
	}
}
