package application

import (
	"context"
	"strings"
	"time"

	"tp1/logic"
)

type Registro struct {
	repo Repositorio
}

func NuevoRegistro(repo Repositorio) *Registro {
	return &Registro{repo: repo}
}

func (s *Registro) ListarZapatillas(ctx context.Context) ([]logic.Zapatilla, error) {
	return s.repo.ListarZapatillas(ctx)
}

func (s *Registro) ListarEntrenamientos(ctx context.Context) ([]logic.Entrenamiento, error) {
	entrenamientos, err := s.repo.ListarEntrenamientos(ctx)
	if err != nil {
		return nil, err
	}
	for i := range entrenamientos {
		entrenamiento := &entrenamientos[i]
		entrenamiento.Ritmo = logic.CalcularRitmo(entrenamiento.DistanciaKm, int(entrenamiento.TiempoMin))
	}
	return entrenamientos, nil
}

func (s *Registro) AgregarZapatilla(ctx context.Context, marca string, kilometros float64) error {
	marca = strings.TrimSpace(marca)
	if err := logic.ValidateZapatilla(marca, kilometros); err != nil {
		return ErrorValidacion(err.Error())
	}
	return s.repo.GuardarZapatilla(ctx, logic.Zapatilla{
		MarcaModelo: marca, KmsAcumulados: kilometros, Estado: true,
	})
}

func (s *Registro) RetirarZapatilla(ctx context.Context, id int32) error {
	return s.repo.EnTransaccion(ctx, func(cambios Cambios) error {
		if _, err := cambios.BuscarZapatillaParaActualizar(ctx, id); err != nil {
			return err
		}
		return cambios.RetirarZapatilla(ctx, id)
	})
}

type NuevoEntrenamiento struct {
	Fecha       time.Time
	DistanciaKm float64
	TiempoMin   int
	Tipo        string
	ZapatillaID int
	Lugar       string
}

func (s *Registro) RegistrarEntrenamiento(ctx context.Context, datos NuevoEntrenamiento) error {
	tipo := strings.TrimSpace(datos.Tipo)
	if err := logic.ValidateEntrenamiento(datos.Fecha, datos.DistanciaKm, datos.TiempoMin, tipo, datos.ZapatillaID); err != nil {
		return ErrorValidacion(err.Error())
	}

	entrenamiento := logic.Entrenamiento{
		Fecha:       datos.Fecha,
		DistanciaKm: datos.DistanciaKm,
		TiempoMin:   int32(datos.TiempoMin),
		Tipo:        tipo,
		ZapatillaID: int32(datos.ZapatillaID),
		Ritmo:       logic.CalcularRitmo(datos.DistanciaKm, datos.TiempoMin),
		Calorias:    logic.CalcularCalorias(datos.DistanciaKm),
	}
	if lugar := strings.TrimSpace(datos.Lugar); lugar != "" {
		entrenamiento.Lugar = &lugar
	}

	return s.repo.EnTransaccion(ctx, func(cambios Cambios) error {
		zapatilla, err := cambios.BuscarZapatillaParaActualizar(ctx, entrenamiento.ZapatillaID)
		if err != nil {
			return err
		}
		if !zapatilla.Estado {
			return ErrorValidacion("La zapatilla seleccionada ya no está activa")
		}
		if err := cambios.GuardarEntrenamiento(ctx, entrenamiento); err != nil {
			return err
		}
		return cambios.CambiarKilometros(ctx, zapatilla.ID, entrenamiento.DistanciaKm)
	})
}

func (s *Registro) EliminarEntrenamiento(ctx context.Context, id int32) error {
	return s.repo.EnTransaccion(ctx, func(cambios Cambios) error {
		entrenamiento, err := cambios.BuscarEntrenamientoParaActualizar(ctx, id)
		if err != nil {
			return err
		}
		if _, err := cambios.BuscarZapatillaParaActualizar(ctx, entrenamiento.ZapatillaID); err != nil {
			return err
		}
		if err := cambios.BorrarEntrenamiento(ctx, id); err != nil {
			return err
		}
		return cambios.CambiarKilometros(ctx, entrenamiento.ZapatillaID, -entrenamiento.DistanciaKm)
	})
}
