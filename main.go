package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	db "tp1/db/sqlc"
	"tp1/logic"

	_ "github.com/lib/pq"
)

// Variable global para las consultas generadas por sqlc
var queries *db.Queries

// Obj. de transferencia de datos para datos de entrada JSON (entrenamientos)

type EntrenamientoRequest struct {
	Fecha       string `json:"fecha"` // Formato esperado: "YYYY-MM-DD"
	TiempoMin   int    `json:"tiempo_min"`
	Tipo        string `json:"tipo"`
	ZapatillaID int    `json:"zapatilla_id"`
	Lugar       string `json:"lugar"`
}

// EMiddleware de logging

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

// Func Princ. y config del server

func setupRouter() http.Handler {
	mux := http.NewServeMux()

	// Endpoints API REST para Zapatillas (Estilo Filmina 25)
	mux.HandleFunc("/zapatillas", zapatillasHandler)
	mux.HandleFunc("/zapatillas/", zapatillaByIDHandler)

	// Endpoints API REST para Entrenamientos
	mux.HandleFunc("/entrenamientos", entrenamientosHandler)
	mux.HandleFunc("/entrenamientos/", entrenamientoByIDHandler)

	// Aplicamos el loggingMiddleware a todos los endpoints
	return loggingMiddleware(mux)
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://tp_user:password@localhost:5432/tp_db?sslmode=disable"
	}

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Error abriendo conexión a la BD: %v", err)
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		log.Fatalf("Error conectando a la BD (Ping falló): %v", err)
	}

	queries = db.New(conn)

	handler := setupRouter()

	log.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// Manejador para /zapatillas (GET para listar, POST para crear)
func zapatillasHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/zapatillas" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getZapatillas(w, r)
	case http.MethodPost:
		createZapatilla(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Manejador único para /zapatillas/{id} (GET, PUT, DELETE)
func zapatillaByIDHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) != 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(parts[2])
	if err != nil || id <= 0 {
		http.Error(w, "Invalid zapatilla ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getZapatilla(w, r, int32(id))
	case http.MethodPut:
		updateZapatilla(w, r, int32(id))
	case http.MethodDelete:
		deleteZapatilla(w, r, int32(id))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GET /zapatillas
func getZapatillas(w http.ResponseWriter, r *http.Request) {
	zapatillas, err := queries.ListZapatillas(r.Context())
	if err != nil {
		http.Error(w, "Error obteniendo zapatillas", http.StatusInternalServerError)
		return
	}
	if zapatillas == nil {
		zapatillas = []db.Zapatilla{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(zapatillas)
}

// POST /zapatillas
func createZapatilla(w http.ResponseWriter, r *http.Request) {
	var input db.CreateZapatillaParams
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Validación con el paquete logic (Ejercicio 1 y 2)
	if err := logic.ValidateZapatilla(input.MarcaModelo, input.KmsAcumulados); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Por defecto, al crear una zapatilla está activa
	input.Estado = true

	creada, err := queries.CreateZapatilla(r.Context(), input)
	if err != nil {
		http.Error(w, "Error guardando zapatilla", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(creada)
}

// GET /zapatillas/{id}
func getZapatilla(w http.ResponseWriter, r *http.Request, id int32) {
	zapatilla, err := queries.GetZapatilla(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Zapatilla no encontrada", http.StatusNotFound) // 404
			return
		}
		http.Error(w, "Error buscando zapatilla", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(zapatilla)
}

// PUT /zapatillas/{id}
func updateZapatilla(w http.ResponseWriter, r *http.Request, id int32) {
	var input db.UpdateZapatillaParams
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if err := logic.ValidateZapatilla(input.MarcaModelo, input.KmsAcumulados); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	input.ID = id
	actualizada, err := queries.UpdateZapatilla(r.Context(), input)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Zapatilla no encontrada", http.StatusNotFound) // 404
			return
		}
		http.Error(w, "Error actualizando zapatilla", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(actualizada)
}

// DELETE /zapatillas/{id}
func deleteZapatilla(w http.ResponseWriter, r *http.Request, id int32) {
	// Verificamos primero que exista para poder devolver 404 si no existe
	_, err := queries.GetZapatilla(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Zapatilla no encontrada", http.StatusNotFound)
			return
		}
		http.Error(w, "Error verificando zapatilla", http.StatusInternalServerError)
		return
	}

	if err := queries.DeleteZapatilla(r.Context(), id); err != nil {
		http.Error(w, "Error eliminando zapatilla", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content
}

//Manejadores para entrenamientos

func entrenamientosHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/entrenamientos" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getEntrenamientos(w, r)
	case http.MethodPost:
		createEntrenamiento(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func entrenamientoByIDHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) != 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(parts[2])
	if err != nil || id <= 0 {
		http.Error(w, "Invalid entrenamiento ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getEntrenamiento(w, r, int32(id))
	case http.MethodPut:
		updateEntrenamiento(w, r, int32(id))
	case http.MethodDelete:
		deleteEntrenamiento(w, r, int32(id))
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func parseFecha(fechaStr string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", fechaStr); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, fechaStr)
}

// GET /entrenamientos
func getEntrenamientos(w http.ResponseWriter, r *http.Request) {
	entrenamientos, err := queries.ListEntrenamientos(r.Context())
	if err != nil {
		http.Error(w, "Error obteniendo entrenamientos", http.StatusInternalServerError)
		return
	}
	if entrenamientos == nil {
		entrenamientos = []db.Entrenamiento{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entrenamientos)
}

// POST /entrenamientos
func createEntrenamiento(w http.ResponseWriter, r *http.Request) {
	var req EntrenamientoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	fechaParsed, err := parseFecha(req.Fecha)
	if err != nil {
		http.Error(w, "Formato de fecha inválido (use YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	// Validamos reglas de negocio usando funciones puras del paquete logic
	if err := logic.ValidateEntrenamiento(fechaParsed, req.DistanciaKm, req.TiempoMin, req.Tipo, req.ZapatillaID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Calculamos ritmo y calorías usando nuestras funciones puras
	ritmo := logic.CalcularRitmo(req.DistanciaKm, req.TiempoMin)
	calorias := logic.CalcularCalorias(req.DistanciaKm)

	creado, err := queries.CreateEntrenamiento(r.Context(), db.CreateEntrenamientoParams{
		Fecha:       fechaParsed,
		DistanciaKm: req.DistanciaKm,
		TiempoMin:   int32(req.TiempoMin),
		Tipo:        req.Tipo,
		ZapatillaID: int32(req.ZapatillaID),
		Ritmo:       ritmo,
		Calorias:    calorias,
		Lugar:       sql.NullString{String: req.Lugar, Valid: req.Lugar != ""},
	})
	if err != nil {
		http.Error(w, "Error guardando entrenamiento (verifique que la zapatilla exista)", http.StatusInternalServerError)
		return
	}

	// Regla de negocio: acumular los kilómetros en la zapatilla usada
	_ = queries.UpdateZapatillaKms(r.Context(), db.UpdateZapatillaKmsParams{
		ID:            int32(req.ZapatillaID),
		KmsAcumulados: req.DistanciaKm,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(creado)
}

// GET /entrenamientos/{id}
func getEntrenamiento(w http.ResponseWriter, r *http.Request, id int32) {
	entrenamiento, err := queries.GetEntrenamiento(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Entrenamiento no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "Error obteniendo entrenamiento", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entrenamiento)
}

// PUT /entrenamientos/{id}
func updateEntrenamiento(w http.ResponseWriter, r *http.Request, id int32) {
	var req EntrenamientoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	fechaParsed, err := parseFecha(req.Fecha)
	if err != nil {
		http.Error(w, "Formato de fecha inválido (use YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	if err := logic.ValidateEntrenamiento(fechaParsed, req.DistanciaKm, req.TiempoMin, req.Tipo, req.ZapatillaID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ritmo := logic.CalcularRitmo(req.DistanciaKm, req.TiempoMin)
	calorias := logic.CalcularCalorias(req.DistanciaKm)

	actualizado, err := queries.UpdateEntrenamiento(r.Context(), db.UpdateEntrenamientoParams{
		ID:          id,
		Fecha:       fechaParsed,
		DistanciaKm: req.DistanciaKm,
		TiempoMin:   int32(req.TiempoMin),
		Tipo:        req.Tipo,
		ZapatillaID: int32(req.ZapatillaID),
		Ritmo:       ritmo,
		Calorias:    calorias,
		Lugar:       sql.NullString{String: req.Lugar, Valid: req.Lugar != ""},
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Entrenamiento no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "Error actualizando entrenamiento", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(actualizado)
}

// DELETE /entrenamientos/{id}
func deleteEntrenamiento(w http.ResponseWriter, r *http.Request, id int32) {
	_, err := queries.GetEntrenamiento(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Entrenamiento no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "Error buscando entrenamiento", http.StatusInternalServerError)
		return
	}

	if err := queries.DeleteEntrenamiento(r.Context(), id); err != nil {
		http.Error(w, "Error eliminando entrenamiento", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content
}
