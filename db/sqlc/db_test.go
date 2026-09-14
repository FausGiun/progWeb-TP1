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
	cleanDatabase(t)
	queries := New(testDatabase)
	ctx := context.Background()

	// Test de Creación
	created, err := queries.CreateZapatilla(ctx, CreateZapatillaParams{
		MarcaModelo:   "Nike Pegasus 40",
		KmsAcumulados: 0,
		Estado:        true,
	})
	if err != nil {
		t.Fatalf("CreateZapatilla() error = %v", err)
	}

	// Test de Lectura
	listed, err := queries.ListZapatillas(ctx)
	if err != nil || len(listed) != 1 {
		t.Fatalf("ListZapatillas() error o cantidad incorrecta")
	}

	// Test de Actualización de Kilómetros
	if err := queries.UpdateZapatillaKms(ctx, UpdateZapatillaKmsParams{
		ID:            created.ID,
		KmsAcumulados: 12.5,
	}); err != nil {
		t.Fatalf("UpdateZapatillaKms() error = %v", err)
	}

	err = queries.DeleteZapatilla(ctx, created.ID)
	if err != nil {
		t.Fatalf("DeleteZapatilla() falló con el error = %v", err)
	}

	zapatillaBorrada, err := queries.GetZapatilla(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetZapatilla() error al buscar zapatilla borrada = %v", err)
	}

	if zapatillaBorrada.Estado != false {
		t.Fatal("Se esperaba que el estado de la zapatilla fuera false después del borrado lógico")
	}

	updated, _ := queries.ListZapatillas(ctx)
	if updated[0].KmsAcumulados != 12.5 {
		t.Fatalf("kms_acumulados = %v, se esperaba 12.5", updated[0].KmsAcumulados)
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
		Lugar:       sql.NullString{String: "Campus UNICEN", Valid: true},
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
