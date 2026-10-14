package logic

import (
	"testing"
	"time"
)

func TestValidateZapatilla(t *testing.T) {
	if err := ValidateZapatilla("", 10); err == nil {
		t.Error("Se esperaba error por marca vacía, pero no falló")
	}
	if err := ValidateZapatilla("Nike Pegasus", -5); err == nil {
		t.Error("Se esperaba error por kms negativos, pero no falló")
	}
	if err := ValidateZapatilla("Nike Pegasus", 15.5); err != nil {
		t.Errorf("No se esperaba error para datos válidos, se obtuvo: %v", err)
	}
}

func TestValidateEntrenamiento(t *testing.T) {
	maniana := time.Now().Add(24 * time.Hour)
	if err := ValidateEntrenamiento(maniana, 5, 30, "Fondo", 1); err == nil {
		t.Error("Se esperaba error por fecha futura, pero no falló")
	}

	hoy := time.Now().Add(-1 * time.Hour)
	if err := ValidateEntrenamiento(hoy, 10.0, 60, "Fondo", 1); err != nil {
		t.Errorf("No se esperaba error para entrenamiento válido: %v", err)
	}
}
