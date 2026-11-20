package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

var testDatabase *sql.DB

func TestMain(m *testing.M) {
	databaseSource := os.Getenv("TEST_DATABASE_URL")
	if databaseSource == "" {
		fmt.Fprintln(os.Stderr, "Falta TEST_DATABASE_URL. Ejecutá make test para usar la base de pruebas.")
		os.Exit(1)
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

	zapatilla, err := queries.CreateZapatilla(ctx, CreateZapatillaParams{
		MarcaModelo:   "Nike Pegasus 40",
		KmsAcumulados: 0,
		Estado:        true,
	})
	if err != nil {
		t.Fatalf("CreateZapatilla() error = %v", err)
	}

	zapatillas, err := queries.ListZapatillas(ctx)
	if err != nil {
		t.Fatalf("ListZapatillas() falló con error: %v", err)
	}
	if len(zapatillas) == 0 {
		t.Fatalf("ListZapatillas() devolvió 0 elementos, se esperaba al menos 1")
	}

	err = queries.UpdateZapatillaKms(ctx, UpdateZapatillaKmsParams{
		ID:            zapatilla.ID,
		KmsAcumulados: 12.5,
	})
	if err != nil {
		t.Fatalf("UpdateZapatillaKms() error = %v", err)
	}

	err = queries.DeleteZapatilla(ctx, zapatilla.ID)
	if err != nil {
		t.Fatalf("DeleteZapatilla() error = %v", err)
	}
}

func TestEntrenamientoCRUD(t *testing.T) {
	cleanDatabase(t)
	queries := New(testDatabase)
	ctx := context.Background()

	zapatilla, _ := queries.CreateZapatilla(ctx, CreateZapatillaParams{
		MarcaModelo: "Asics Novablast 3",
		Estado:      true,
	})

	entrenamiento, err := queries.CreateEntrenamiento(ctx, CreateEntrenamientoParams{
		Fecha:       time.Now(),
		DistanciaKm: 10.0,
		TiempoMin:   50,
		Tipo:        "Fondo",
		ZapatillaID: zapatilla.ID,
		Ritmo:       5.0,
		Calorias:    600,
		Lugar:       sql.NullString{String: "Parque", Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateEntrenamiento() error = %v", err)
	}

	guardado, err := queries.GetEntrenamiento(ctx, entrenamiento.ID)
	if err != nil || guardado.ID != entrenamiento.ID {
		t.Fatalf("GetEntrenamiento() falló al recuperar el registro")
	}

	if err := queries.DeleteEntrenamiento(ctx, entrenamiento.ID); err != nil {
		t.Fatalf("DeleteEntrenamiento() error = %v", err)
	}

	_, err = queries.GetEntrenamiento(ctx, entrenamiento.ID)
	if err == nil {
		t.Fatal("GetEntrenamiento() esperaba un error después de eliminar, pero devolvió el registro")
	}
}
