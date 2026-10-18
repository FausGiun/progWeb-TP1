package db

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

const defaultTestDatabaseSource = "postgres://tp_user:password@localhost:5432/tp_db?sslmode=disable"

var testDatabase *sql.DB

func TestMain(m *testing.M) {
	databaseSource := os.Getenv("DB_SOURCE")
	if databaseSource == "" {
		databaseSource = defaultTestDatabaseSource
	}
	database, err := sql.Open("postgres", databaseSource)
	if err != nil {
		os.Exit(1)
	}
	testDatabase = database
	if err := testDatabase.Ping(); err != nil {
		testDatabase.Close()
		os.Exit(1)
	}
	exitCode := m.Run()
	testDatabase.Close()
	os.Exit(exitCode)
}

func cleanDatabase(t *testing.T) {
	t.Helper()
	if _, err := testDatabase.Exec("TRUNCATE TABLE entrenamientos, zapatillas RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("no se pudo limpiar la base de datos: %v", err)
	}
}

func TestZapatillaCRUD(t *testing.T) {

	queries := New(testDatabase)
	ctx := context.Background()

	// 1. Crear (Genera la variable 'created')
	created, err := queries.CreateZapatilla(ctx, CreateZapatillaParams{
		MarcaModelo:   "Nike Pegasus 40",
		KmsAcumulados: 0,
		Estado:        true,
	})
	if err != nil {
		t.Fatalf("CreateZapatilla() error = %v", err)
	}

	// 2. Leer y verificar longitud
	listed, err := queries.ListZapatillas(ctx)
	if err != nil {
		t.Fatalf("ListZapatillas() falló con error: %v", err)
	}
	if len(listed) == 0 {
		t.Fatalf("ListZapatillas() devolvió 0 elementos, se esperaba al menos 1")
	}

	// 3. Actualizar (usa 'created.ID' para resolver el error de variable no usada)
	err = queries.UpdateZapatillaKms(ctx, UpdateZapatillaKmsParams{
		ID:            created.ID,
		KmsAcumulados: 12.5,
	})
	if err != nil {
		t.Fatalf("UpdateZapatillaKms() error = %v", err)
	}

	// 4. Borrar (usa 'created.ID')
	err = queries.DeleteZapatilla(ctx, created.ID)
	if err != nil {
		t.Fatalf("DeleteZapatilla() error = %v", err)
	}
}

func TestEntrenamientoCRUD(t *testing.T) {
	cleanDatabase(t)
	queries := New(testDatabase)
	ctx := context.Background()

	shoe, _ := queries.CreateZapatilla(ctx, CreateZapatillaParams{
		MarcaModelo: "Asics Novablast 3",
		Estado:      true,
	})

	// Test de Creación de Entrenamiento (con Lugar)
	created, err := queries.CreateEntrenamiento(ctx, CreateEntrenamientoParams{
		Fecha:       time.Now(),
		DistanciaKm: 10.0,
		TiempoMin:   50,
		Tipo:        "Fondo",
		ZapatillaID: shoe.ID,
		Ritmo:       5.0,
		Calorias:    600,
		Lugar:       sql.NullString{String: "Parque", Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateEntrenamiento() error = %v", err)
	}

	// Test de Lectura
	found, err := queries.GetEntrenamiento(ctx, created.ID)
	if err != nil || found.ID != created.ID {
		t.Fatalf("GetEntrenamiento() falló al recuperar el registro")
	}

	// Test de Eliminación
	if err := queries.DeleteEntrenamiento(ctx, created.ID); err != nil {
		t.Fatalf("DeleteEntrenamiento() error = %v", err)
	}

	//Para verificar que elimino
	_, err = queries.GetEntrenamiento(ctx, created.ID)
	if err == nil {
		t.Fatal("GetEntrenamiento() esperaba un error después de eliminar, pero devolvió el registro")
	}
}
