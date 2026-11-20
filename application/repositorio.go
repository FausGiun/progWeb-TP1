package application

import (
	"context"
	"errors"

	"tp1/logic"
)

var (
	ErrZapatillaNoEncontrada     = errors.New("Zapatilla no encontrada")
	ErrEntrenamientoNoEncontrado = errors.New("Entrenamiento no encontrado")
)

type ErrorValidacion string

func (e ErrorValidacion) Error() string { return string(e) }

type Repositorio interface {
	ListarZapatillas(ctx context.Context) ([]logic.Zapatilla, error)
	ListarEntrenamientos(ctx context.Context) ([]logic.Entrenamiento, error)
	GuardarZapatilla(ctx context.Context, zapatilla logic.Zapatilla) error
	EnTransaccion(ctx context.Context, operacion func(Cambios) error) error
}

type Cambios interface {
	BuscarZapatillaParaActualizar(ctx context.Context, id int32) (logic.Zapatilla, error)
	BuscarEntrenamientoParaActualizar(ctx context.Context, id int32) (logic.Entrenamiento, error)
	GuardarEntrenamiento(ctx context.Context, entrenamiento logic.Entrenamiento) error
	CambiarKilometros(ctx context.Context, id int32, distancia float64) error
	RetirarZapatilla(ctx context.Context, id int32) error
	BorrarEntrenamiento(ctx context.Context, id int32) error
}
