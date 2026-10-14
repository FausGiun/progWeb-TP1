package logic

import (
	"errors"
	"strings"
	"time"
)

func ValidateZapatilla(marcaModelo string, kmsAcumulados float64) error {
	if strings.TrimSpace(marcaModelo) == "" {
		return errors.New("la marca y modelo de la zapatilla no pueden estar vacíos")
	}
	if kmsAcumulados < 0 {
		return errors.New("los kilómetros acumulados no pueden ser negativos")
	}
	return nil
}

func ValidateEntrenamiento(fecha time.Time, distanciaKm float64, tiempoMin int, tipo string, zapatillaID int) error {
	if fecha.IsZero() {
		return errors.New("la fecha del entrenamiento es obligatoria")
	}
	if fecha.After(time.Now()) {
		return errors.New("la fecha del entrenamiento no puede ser en el futuro")
	}
	if distanciaKm <= 0 {
		return errors.New("la distancia debe ser mayor a 0 km")
	}
	if tiempoMin <= 0 {
		return errors.New("el tiempo debe ser mayor a 0 minutos")
	}
	if strings.TrimSpace(tipo) == "" {
		return errors.New("el tipo de entrenamiento no puede estar vacío")
	}
	if zapatillaID <= 0 {
		return errors.New("debe especificar un ID de zapatilla válido")
	}
	return nil
}

func CalcularRitmo(distanciaKm float64, tiempoMin int) float64 {
	if tiempoMin <= 0 {
		return 0
	}
	horas := float64(tiempoMin) / 60.0
	return distanciaKm / horas
}

func CalcularCalorias(distanciaKm float64) int32 {
	return int32(distanciaKm * 65.0)
}
