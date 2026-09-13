package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"
	db "tp1/db/sqlc"

	_ "github.com/lib/pq"
)

// Estructuras

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

var dbQueries *db.Queries

// handlers
// 1. Inicio
func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "static/index.html")
}

// 2. Historial de Entrenamientos
func historialHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/historial.html")
	if err != nil {
		http.Error(w, "Error cargando template", http.StatusInternalServerError)
		return
	}

	// Obtenemos los registros reales de la base de datos
	entrenamientosDB, _ := dbQueries.ListEntrenamientos(context.Background())
	var entrenamientos []Entrenamiento

	for _, e := range entrenamientosDB {
		entrenamientos = append(entrenamientos, Entrenamiento{
			ID:          int(e.ID),
			Fecha:       e.Fecha,
			DistanciaKm: e.DistanciaKm,
			TiempoMin:   int(e.TiempoMin),
			Tipo:        e.Tipo,
			Ritmo:       e.Ritmo,
			Calorias:    int(e.Calorias),
			Lugar:       e.Lugar.String, // Convertimos el NullString a string normal
		})
	}

	tmpl.ExecuteTemplate(w, "historial.html", entrenamientos)
}

// 3. Formulario de entrenamiento
func registrarEjHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/registrarEj.html")
	if err != nil {
		http.Error(w, "Error cargando template", http.StatusInternalServerError)
		return
	}

	zapatillasDB, _ := dbQueries.ListZapatillas(context.Background())
	var zapatillas []Zapatilla

	for _, z := range zapatillasDB {
		zapatillas = append(zapatillas, Zapatilla{
			ID:            int(z.ID),
			MarcaModelo:   z.MarcaModelo,
			KmsAcumulados: z.KmsAcumulados,
			Estado:        z.Estado,
		})
	}

	tmpl.ExecuteTemplate(w, "registrarEj.html", zapatillas)
}

// 4- Guardar entrenamiento
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

	lugarStr := r.FormValue("lugar")
	lugarNull := sql.NullString{String: lugarStr, Valid: lugarStr != ""}

	_, err := dbQueries.CreateEntrenamiento(context.Background(), db.CreateEntrenamientoParams{
		Fecha:       fechaParsed,
		DistanciaKm: distancia,
		TiempoMin:   int32(tiempo),
		Tipo:        r.FormValue("tipo"),
		ZapatillaID: int32(zapatillaID),
		Ritmo:       distancia / (float64(tiempo) / 60),
		Calorias:    int32(calorias),
		Lugar:       lugarNull,
	})

	if err == nil {
		dbQueries.UpdateZapatillaKms(context.Background(), db.UpdateZapatillaKmsParams{
			ID:            int32(zapatillaID),
			KmsAcumulados: distancia,
		})
	}

	http.Redirect(w, r, "/historial", http.StatusSeeOther)
}

// 5- Formulario y guardado de zapatilla nueva
func nuevaZapatillaHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/nuevaZapatilla.html")
}

func guardarZapatillaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		dbQueries.CreateZapatilla(context.Background(), db.CreateZapatillaParams{
			MarcaModelo:   r.FormValue("marca_modelo"),
			KmsAcumulados: 0,
			Estado:        true,
		})
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
	dbQueries = db.New(conn)
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
