package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

// Estructuras

type Entrenamiento struct {
	ID          int
	Fecha       time.Time
	DistanciaKm float64
	TiempoMin   int
	Tipo        string
	ZapatillaID int
}

type Zapatilla struct {
	ID            int
	MarcaModelo   string
	KmsAcumulados float64
	Estado        string
}

// Memoria

var (
	entrenamientos []Entrenamiento
	proximoIDEnt   = 1

	// Inicializamos con una zapatilla hardcodeada para probar la funcionalidad
	zapatillas = []Zapatilla{
		{ID: 1, MarcaModelo: "Nike Pegasus 40", KmsAcumulados: 0, Estado: "Activa"},
	}
	proximoIDZapa = 2
)

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
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "historial.html", entrenamientos)
}

// 3. Formulario de entrenamiento
func registrarEjHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/registrarEj.html")
	if err != nil {
		http.Error(w, "Error", http.StatusInternalServerError)
		return
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

	nuevoEnt := Entrenamiento{
		ID:          proximoIDEnt,
		Fecha:       fechaParsed,
		DistanciaKm: distancia,
		TiempoMin:   tiempo,
		Tipo:        r.FormValue("tipo"),
		ZapatillaID: zapatillaID,
	}

	entrenamientos = append(entrenamientos, nuevoEnt)
	proximoIDEnt++

	for i := range zapatillas {
		if zapatillas[i].ID == zapatillaID {
			zapatillas[i].KmsAcumulados += distancia
			break
		}
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
		nuevaZapa := Zapatilla{
			ID:            proximoIDZapa,
			MarcaModelo:   r.FormValue("marca_modelo"),
			KmsAcumulados: 0,
			Estado:        "Activa",
		}
		zapatillas = append(zapatillas, nuevaZapa)
		proximoIDZapa++
	}
	http.Redirect(w, r, "/nuevo-entrenamiento", http.StatusSeeOther)
}

func main() {
	//http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/historial", historialHandler)
	http.HandleFunc("/nuevo-entrenamiento", registrarEjHandler)
	http.HandleFunc("/guardar-entrenamiento", guardarEntrenamientoHandler)
	http.HandleFunc("/nueva-zapatilla", nuevaZapatillaHandler)
	http.HandleFunc("/guardar-zapatilla", guardarZapatillaHandler)

	port := ":8080"
	fmt.Printf("Servidor corriendo en http://localhost%s\n", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Printf("Error al iniciar el servidor: %s\n", err)
	}
}
