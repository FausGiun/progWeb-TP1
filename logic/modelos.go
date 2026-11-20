package logic

import "time"

type Zapatilla struct {
	ID            int32
	MarcaModelo   string
	KmsAcumulados float64
	Estado        bool
}

type Entrenamiento struct {
	ID          int32
	Fecha       time.Time
	DistanciaKm float64
	TiempoMin   int32
	Tipo        string
	ZapatillaID int32
	Ritmo       float64
	Calorias    int32
	Lugar       *string
}
