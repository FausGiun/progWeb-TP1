package main

import (
	"context"
	"database/sql"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	db "tp1/db/sqlc"
)

func testDatabase(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Fatal("Falta TEST_DATABASE_URL. Ejecutá make test.")
	}
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	if _, err := conn.Exec("TRUNCATE entrenamientos, zapatillas RESTART IDENTITY CASCADE"); err != nil {
		t.Fatal(err)
	}
	return conn
}

func webRequest(t *testing.T, router http.Handler, method, path string, form url.Values, status int) *httptest.ResponseRecorder {
	t.Helper()
	return checkWebRequest(t, router, method, path, form, status, false)
}

func checkWebRequest(t *testing.T, router http.Handler, method, path string, form url.Values, status int, htmx bool) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if htmx {
		req.Header.Set("HX-Request", "true")
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	if response.Code != status {
		t.Fatalf("%s %s: estado %d, esperado %d: %s", method, path, response.Code, status, response.Body.String())
	}
	return response
}

func TestRutas(t *testing.T) {
	router := setupRouter(nil)
	webRequest(t, router, "GET", "/static/styles.css", nil, 200)
	webRequest(t, router, "GET", "/no-existe", nil, 404)
	webRequest(t, router, "POST", "/", nil, 405)
	response := webRequest(t, router, "GET", "/zapatillas", nil, 405)
	if response.Header().Get("Allow") != "POST" {
		t.Fatal("El alta de zapatillas solo debe aceptar POST")
	}
	response = webRequest(t, router, "GET", "/mis-entrenamientos", nil, 303)
	if response.Header().Get("Location") != "/" {
		t.Fatal("La ruta anterior de entrenamientos debe llevar al inicio")
	}
}

func TestFormulariosYRenderizado(t *testing.T) {
	conn := testDatabase(t)
	q := db.New(conn)
	ctx := context.Background()
	router := setupRouter(conn)
	response := webRequest(t, router, "GET", "/", nil, 200)
	if !strings.Contains(response.Body.String(), "disabled") || strings.Contains(response.Body.String(), "app.js") {
		t.Fatal("Sin zapatillas, el formulario debe estar deshabilitado y funcionar sin app.js")
	}
	webRequest(t, router, "POST", "/zapatillas", url.Values{"marca_modelo": {""}, "kms_acumulados": {"0"}}, 400)
	for _, invalid := range []string{"abc", "-1", "NaN", "+Inf"} {
		webRequest(t, router, "POST", "/zapatillas", url.Values{"marca_modelo": {"Prueba"}, "kms_acumulados": {invalid}}, 400)
	}
	response = webRequest(t, router, "POST", "/zapatillas", url.Values{
		"marca_modelo": {"Asics <b>prueba</b>"}, "kms_acumulados": {"12.5"},
	}, 303)
	if response.Header().Get("Location") != "/mis-zapatillas" {
		t.Fatal("El alta debe redirigir a la página de zapatillas")
	}
	zapatillas, err := q.ListZapatillas(ctx)
	if err != nil || len(zapatillas) != 1 {
		t.Fatalf("Alta de zapatilla: %v, %v", zapatillas, err)
	}
	zapatillaID := strconv.Itoa(int(zapatillas[0].ID))
	response = webRequest(t, router, "GET", "/mis-zapatillas", nil, 200)
	if !strings.Contains(response.Body.String(), "Asics &lt;b&gt;prueba&lt;/b&gt;") || strings.Contains(response.Body.String(), `id="form-entrenamiento"`) {
		t.Fatal("La página debe mostrar los datos escapados y solo el formulario de zapatillas")
	}
	response = webRequest(t, router, "GET", "/", nil, 200)
	if strings.Contains(response.Body.String(), "disabled") || strings.Contains(response.Body.String(), `id="form-zapatilla"`) {
		t.Fatal("El entrenamiento debe habilitarse al existir zapatillas y tener su propia página")
	}
	form := url.Values{
		"fecha": {"2025-05-10"}, "distancia_km": {"10"}, "tiempo_min": {"50"},
		"tipo": {"Fondo"}, "zapatilla_id": {zapatillaID}, "lugar": {"Parque <script>prueba</script>"},
	}
	response = webRequest(t, router, "POST", "/entrenamientos", form, 303)
	if response.Header().Get("Location") != "/" {
		t.Fatal("El alta debe aplicar POST/Redirect/GET hacia el inicio")
	}
	entrenamientos, err := q.ListEntrenamientos(ctx)
	if err != nil || len(entrenamientos) != 1 {
		t.Fatalf("Alta de entrenamiento: %v, %v", entrenamientos, err)
	}
	if entrenamientos[0].Ritmo != 5 || entrenamientos[0].Calorias != 650 {
		t.Fatal("El servidor debe calcular ritmo y calorías")
	}
	zapatilla, err := q.GetZapatilla(ctx, zapatillas[0].ID)
	if err != nil || zapatilla.KmsAcumulados != 22.5 {
		t.Fatalf("Kilometraje luego del alta: %v, %v", zapatilla.KmsAcumulados, err)
	}
	response = webRequest(t, router, "GET", "/", nil, 200)
	html := response.Body.String()
	if !strings.Contains(response.Header().Get("Content-Type"), "text/html") || !strings.Contains(html, "5.00 min/km") || !strings.Contains(html, "Parque &lt;script&gt;prueba&lt;/script&gt;") || strings.Contains(html, "<script>prueba</script>") {
		t.Fatal("GET / debe contener el listado HTML completo y escapar los datos")
	}
	webRequest(t, router, "GET", "/", nil, 200)
	registros, err := q.ListEntrenamientos(ctx)
	if err != nil || len(registros) != 1 {
		t.Fatal("Actualizar la página no debe repetir el POST")
	}
	entrenamientoID := strconv.Itoa(int(entrenamientos[0].ID))
	response = webRequest(t, router, "POST", "/entrenamientos/eliminar", url.Values{"id": {entrenamientoID}}, 303)
	if response.Header().Get("Location") != "/" {
		t.Fatal("El borrado debe redirigir al listado")
	}
	zapatilla, err = q.GetZapatilla(ctx, zapatillas[0].ID)
	if err != nil || math.Abs(zapatilla.KmsAcumulados-12.5) > 0.000001 {
		t.Fatal("Borrar una salida debe descontar su distancia")
	}
	webRequest(t, router, "POST", "/entrenamientos/eliminar", url.Values{"id": {entrenamientoID}}, 404)
	webRequest(t, router, "POST", "/zapatillas/eliminar", url.Values{"id": {"abc"}}, 400)
	webRequest(t, router, "POST", "/zapatillas/eliminar", url.Values{"id": {zapatillaID}}, 303)
	webRequest(t, router, "POST", "/entrenamientos", form, 400)
	registros, err = q.ListEntrenamientos(ctx)
	if err != nil || len(registros) != 0 {
		t.Fatal("La zapatilla retirada no debe admitir nuevas salidas")
	}
}

func TestAltasYBajasHTMX(t *testing.T) {
	conn := testDatabase(t)
	router := setupRouter(conn)
	q := db.New(conn)
	ctx := context.Background()

	checkFragment := func(response *httptest.ResponseRecorder, panel, record string) {
		t.Helper()
		html := response.Body.String()
		if !strings.HasPrefix(html, `<div id="`+panel+`"`) || !strings.Contains(html, record) ||
			strings.Contains(html, "<!DOCTYPE") || response.Header().Get("Location") != "" {
			t.Fatalf("Se esperaba solo el listado %s, sin redirección: %s", panel, html)
		}
	}

	response := checkWebRequest(t, router, http.MethodPost, "/zapatillas", url.Values{
		"marca_modelo": {"Par HTMX"}, "kms_acumulados": {"12.5"},
	}, 200, true)
	checkFragment(response, "panel-zapatillas", "Par HTMX")
	zapatillas, err := q.ListZapatillas(ctx)
	if err != nil || len(zapatillas) != 1 {
		t.Fatalf("No se guardó la zapatilla: %v, %v", zapatillas, err)
	}
	zapatillaID := strconv.Itoa(int(zapatillas[0].ID))

	response = checkWebRequest(t, router, http.MethodPost, "/entrenamientos", url.Values{
		"fecha": {"2025-05-10"}, "distancia_km": {"10"}, "tiempo_min": {"50"},
		"tipo": {"Fondo"}, "zapatilla_id": {zapatillaID},
	}, 200, true)
	checkFragment(response, "panel-entrenamientos", "5.00 min/km")
	entrenamientos, err := q.ListEntrenamientos(ctx)
	if err != nil || len(entrenamientos) != 1 {
		t.Fatalf("No se guardó el entrenamiento: %v, %v", entrenamientos, err)
	}
	entrenamientoURL := "/entrenamientos/" + strconv.Itoa(int(entrenamientos[0].ID))
	if !strings.Contains(response.Body.String(), `hx-delete="`+entrenamientoURL+`"`) {
		t.Fatal("El botón de borrado debe usar el ID del entrenamiento en la URL")
	}
	response = checkWebRequest(t, router, http.MethodDelete, entrenamientoURL, nil, 200, true)
	if response.Body.Len() != 0 {
		t.Fatal("El DELETE debe devolver un cuerpo vacío para quitar la fila")
	}
	zapatilla, err := q.GetZapatilla(ctx, zapatillas[0].ID)
	if err != nil || zapatilla.KmsAcumulados != 12.5 {
		t.Fatalf("El borrado no devolvió el kilometraje inicial: %v, %v", zapatilla, err)
	}
	checkWebRequest(t, router, http.MethodDelete, entrenamientoURL, nil, 404, true)
	checkWebRequest(t, router, http.MethodDelete, "/zapatillas/no-es-un-id", nil, 400, true)
	response = checkWebRequest(t, router, http.MethodDelete, "/zapatillas/"+zapatillaID, nil, 200, true)
	zapatilla, err = q.GetZapatilla(ctx, zapatillas[0].ID)
	if response.Body.Len() != 0 || err != nil || zapatilla.Estado {
		t.Fatalf("Se esperaba retirar el par y devolver un cuerpo vacío: %v, %v", zapatilla, err)
	}
}

func TestEntrenamientoAtomico(t *testing.T) {
	conn := testDatabase(t)
	q := db.New(conn)
	ctx := context.Background()
	zapatilla, err := q.CreateZapatilla(ctx, db.CreateZapatillaParams{MarcaModelo: "Prueba", KmsAcumulados: 20, Estado: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec("ALTER TABLE zapatillas ADD CONSTRAINT test_limite_kilometraje CHECK (kms_acumulados <= 25)"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := conn.Exec("ALTER TABLE zapatillas DROP CONSTRAINT test_limite_kilometraje"); err != nil {
			t.Error(err)
		}
	})
	webRequest(t, setupRouter(conn), "POST", "/entrenamientos", url.Values{
		"fecha": {"2025-05-10"}, "distancia_km": {"10"}, "tiempo_min": {"50"},
		"tipo": {"Fondo"}, "zapatilla_id": {strconv.Itoa(int(zapatilla.ID))},
	}, 500)
	actual, err := q.GetZapatilla(ctx, zapatilla.ID)
	if err != nil || actual.KmsAcumulados != 20 {
		t.Fatal("Una operación fallida debe conservar el kilometraje anterior")
	}
	registros, err := q.ListEntrenamientos(ctx)
	if err != nil || len(registros) != 0 {
		t.Fatal("Debe revertirse también el entrenamiento si falla la actualización de kilómetros")
	}
}
