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
	if err := ValidateEntrenamiento(hoy, 10, 60, "Fondo", 4294967297); err == nil {
		t.Error("Se aceptó un ID que se convertiría en otra zapatilla al almacenarlo")
	}
}

func TestCalcularRitmoEnMinutosPorKilometro(t *testing.T) {
	for _, caso := range []struct {
		distancia float64
		minutos   int
		esperado  float64
	}{
		{10, 50, 5},
		{5, 30, 6},
		{0, 30, 0},
		{5, 0, 0},
	} {
		if resultado := CalcularRitmo(caso.distancia, caso.minutos); resultado != caso.esperado {
			t.Errorf("CalcularRitmo(%v, %v) = %v; esperado %v min/km", caso.distancia, caso.minutos, resultado, caso.esperado)
		}
	}
}
