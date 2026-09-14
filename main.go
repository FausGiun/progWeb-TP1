package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"
	db "tp1/db/sqlc"

	_ "github.com/lib/pq"
)

// 1. ESTRUCTURAS y DTO

type Entrenamiento struct {
	ID          int
	Fecha       time.Time
	DistanciaKm float64
	TiempoMin   int
	Tipo        string
	ZapatillaID int
	Ritmo       float64
	Calorias    int
	Lugar       string
}
type Zapatilla struct {
	ID            int
	MarcaModelo   string
	KmsAcumulados float64
	Estado        bool
}

type EntrenamientoDTO struct {
	Fecha       time.Time
	DistanciaKm float64
	TiempoMin   int
	Tipo        string
	ZapatillaID int
	Calorias    int
	Lugar       string
}

// 2. Negocio
type ServicioRunning struct {
	repo *db.Queries
}

func (s *ServicioRunning) ObtenerHistorial(ctx context.Context) ([]Entrenamiento, error) {
	entrenamientosDB, err := s.repo.ListEntrenamientos(ctx)
	if err != nil {
		return nil, err
	}

	// de sql a los struct definidos
	var resultados []Entrenamiento
	for _, e := range entrenamientosDB {
		resultados = append(resultados, Entrenamiento{
			ID:          int(e.ID),
			Fecha:       e.Fecha,
			DistanciaKm: e.DistanciaKm,
			TiempoMin:   int(e.TiempoMin),
			Tipo:        e.Tipo,
			Ritmo:       e.Ritmo,
			Calorias:    int(e.Calorias),
			Lugar:       e.Lugar.String,
		})
	}
	return resultados, nil
}

func (s *ServicioRunning) ObtenerZapatillas(ctx context.Context) ([]Zapatilla, error) {
	zapatillasDB, err := s.repo.ListZapatillas(ctx)
	if err != nil {
		return nil, err
	}

	var resultados []Zapatilla
	for _, zapa := range zapatillasDB {
		resultados = append(resultados, Zapatilla{
			ID:            int(zapa.ID),
			MarcaModelo:   zapa.MarcaModelo,
			KmsAcumulados: zapa.KmsAcumulados,
			Estado:        zapa.Estado,
		})
	}
	return resultados, nil
}

func (s *ServicioRunning) RegistrarNuevaZapatilla(ctx context.Context, marca string) error {
	_, err := s.repo.CreateZapatilla(ctx, db.CreateZapatillaParams{
		MarcaModelo:   marca,
		KmsAcumulados: 0,
		Estado:        true,
	})
	return err
}

func (s *ServicioRunning) RegistrarNuevaSalida(ctx context.Context, dto EntrenamientoDTO) error {
	if dto.Fecha.After(time.Now()) {
		return errors.New("la fecha del entrenamiento no puede ser en el futuro")
	}
	ritmo := dto.DistanciaKm / (float64(dto.TiempoMin) / 60)
	lugarNull := sql.NullString{String: dto.Lugar, Valid: dto.Lugar != ""}

	_, err := s.repo.CreateEntrenamiento(ctx, db.CreateEntrenamientoParams{
		Fecha:       dto.Fecha,
		DistanciaKm: dto.DistanciaKm,
		TiempoMin:   int32(dto.TiempoMin),
		Tipo:        dto.Tipo,
		ZapatillaID: int32(dto.ZapatillaID),
		Ritmo:       ritmo,
		Calorias:    int32(dto.Calorias),
		Lugar:       lugarNull,
	})

	if err != nil {
		return err
	}
	return s.repo.UpdateZapatillaKms(ctx, db.UpdateZapatillaKmsParams{
		ID:            int32(dto.ZapatillaID),
		KmsAcumulados: dto.DistanciaKm,
	})
}

// 3. Handlers

var servicioRunning *ServicioRunning

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "static/index.html")
}

func historialHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/historial.html")
	if err != nil {
		http.Error(w, "Error cargando template", http.StatusInternalServerError)
		return
	}

	entrenamientos, err := servicioRunning.ObtenerHistorial(r.Context())
	if err != nil {
		http.Error(w, "Error obteniendo el historial", http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "historial.html", entrenamientos)
}

func registrarEjHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/registrarEj.html")
	if err != nil {
		http.Error(w, "Error cargando template", http.StatusInternalServerError)
		return
	}

	zapatillas, err := servicioRunning.ObtenerZapatillas(r.Context())
	if err != nil {
		http.Error(w, "Error obteniendo zapatillas", http.StatusInternalServerError)
		return
	}

	tmpl.ExecuteTemplate(w, "registrarEj.html", zapatillas)
}

func guardarEntrenamientoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	r.ParseForm()

	fechaParsed, _ := time.Parse("2006-01-02", r.FormValue("fecha"))
	distancia, _ := strconv.ParseFloat(r.FormValue("distancia_km"), 64)
	tiempo, _ := strconv.Atoi(r.FormValue("tiempo_min"))
	zapatillaID, _ := strconv.Atoi(r.FormValue("zapatilla_id"))

	caloriasStr := r.FormValue("calorias")
	calorias := 0
	if caloriasStr != "" {
		calorias, _ = strconv.Atoi(caloriasStr)
	}

	dto := EntrenamientoDTO{
		Fecha:       fechaParsed,
		DistanciaKm: distancia,
		TiempoMin:   tiempo,
		Tipo:        r.FormValue("tipo"),
		ZapatillaID: zapatillaID,
		Calorias:    calorias,
		Lugar:       r.FormValue("lugar"),
	}

	err := servicioRunning.RegistrarNuevaSalida(r.Context(), dto)
	if err != nil {
		if err.Error() == "la fecha del entrenamiento no puede ser en el futuro" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "Error interno al guardar el entrenamiento", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/historial", http.StatusSeeOther)
}

func nuevaZapatillaHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/nuevaZapatilla.html")
}

func guardarZapatillaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		marca := r.FormValue("marca_modelo")

		err := servicioRunning.RegistrarNuevaZapatilla(r.Context(), marca)
		if err != nil {
			http.Error(w, "Error guardando la zapatilla", http.StatusInternalServerError)
			return
		}
	}
	http.Redirect(w, r, "/nuevo-entrenamiento", http.StatusSeeOther)
}

func main() {
	conn, err := sql.Open("postgres", "postgres://tp_user:password@localhost:5432/tp_db?sslmode=disable")
	if err != nil {
		fmt.Printf("Error conectando a la base de datos: %v\n", err)
		return
	}
	defer conn.Close()

	repo := db.New(conn)
	servicioRunning = &ServicioRunning{repo: repo}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/historial", historialHandler)
	http.HandleFunc("/nuevo-entrenamiento", registrarEjHandler)
	http.HandleFunc("/guardar-entrenamiento", guardarEntrenamientoHandler)
	http.HandleFunc("/nueva-zapatilla", nuevaZapatillaHandler)
	http.HandleFunc("/guardar-zapatilla", guardarZapatillaHandler)

	port := ":8080"
	fmt.Printf("Servidor corriendo en http://localhost%s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Error al iniciar el servidor: %s\n", err)
	}
}
