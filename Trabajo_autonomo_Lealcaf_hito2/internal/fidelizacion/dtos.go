package fidelizacion

import "time"

// Los DTO son la frontera HTTP. Solo estos tipos definen los nombres del JSON.
// Ni el id, ni el cliente, ni los puntos, ni el estado del canje vienen del cuerpo.
type CanjePeticion struct {
	RecompensaID uint `json:"recompensa_id" binding:"required"`
	SucursalID   uint `json:"sucursal_id" binding:"required"`
}

func (p CanjePeticion) aEntrada() EntradaCanje {
	return EntradaCanje{RecompensaID: p.RecompensaID, SucursalID: p.SucursalID}
}

type EstadoCanjePeticion struct {
	Estado string `json:"estado" binding:"required"`
}

type CompraPeticion struct {
	ClienteID     uint    `json:"cliente_id" binding:"required"`
	SucursalID    uint    `json:"sucursal_id" binding:"required"`
	Total         float64 `json:"total" binding:"required,gt=0"`
	NumeroFactura string  `json:"numero_factura" binding:"required"`
}

func (p CompraPeticion) aEntrada() EntradaCompra {
	return EntradaCompra{ClienteID: p.ClienteID, SucursalID: p.SucursalID,
		Total: p.Total, NumeroFactura: p.NumeroFactura}
}

type CanjeRespuesta struct {
	ID               uint      `json:"id"`
	SucursalID       uint      `json:"sucursal_id"`
	ClienteID        uint      `json:"cliente_id"`
	RecompensaID     uint      `json:"recompensa_id"`
	PuntosUtilizados int       `json:"puntos_utilizados"`
	Fecha            time.Time `json:"fecha"`
	Estado           string    `json:"estado"`
}

func nuevaCanjeRespuesta(c Canje) CanjeRespuesta {
	return CanjeRespuesta{ID: c.ID, SucursalID: c.SucursalID, ClienteID: c.ClienteID,
		RecompensaID: c.RecompensaID, PuntosUtilizados: c.PuntosUtilizados,
		Fecha: c.Fecha, Estado: c.Estado}
}

func nuevasCanjesRespuesta(canjes []Canje) []CanjeRespuesta {
	respuestas := make([]CanjeRespuesta, len(canjes))
	for i, canje := range canjes {
		respuestas[i] = nuevaCanjeRespuesta(canje)
	}
	return respuestas
}

type RecompensaRespuesta struct {
	ID               uint   `json:"id"`
	Nombre           string `json:"nombre"`
	Descripcion      string `json:"descripcion"`
	PuntosNecesarios int    `json:"puntos_necesarios"`
	Activa           bool   `json:"activa"`
}

func nuevasRecompensasRespuesta(recompensas []Recompensa) []RecompensaRespuesta {
	respuestas := make([]RecompensaRespuesta, len(recompensas))
	for i, r := range recompensas {
		respuestas[i] = RecompensaRespuesta{ID: r.ID, Nombre: r.Nombre,
			Descripcion: r.Descripcion, PuntosNecesarios: r.PuntosNecesarios, Activa: r.Activa}
	}
	return respuestas
}

type CompraRespuesta struct {
	ID            uint      `json:"id"`
	SucursalID    uint      `json:"sucursal_id"`
	ClienteID     uint      `json:"cliente_id"`
	Fecha         time.Time `json:"fecha"`
	Total         float64   `json:"total"`
	NumeroFactura string    `json:"numero_factura"`
}

type ResultadoCompraRespuesta struct {
	Compra           CompraRespuesta `json:"compra"`
	PuntosGenerados  int             `json:"puntos_generados"`
	FechaVencimiento time.Time       `json:"fecha_vencimiento"`
}

func nuevoResultadoCompraRespuesta(r ResultadoCompra) ResultadoCompraRespuesta {
	c := r.Compra
	return ResultadoCompraRespuesta{
		Compra: CompraRespuesta{ID: c.ID, SucursalID: c.SucursalID, ClienteID: c.ClienteID,
			Fecha: c.Fecha, Total: c.Total, NumeroFactura: c.NumeroFactura},
		PuntosGenerados: r.PuntosGenerados, FechaVencimiento: r.FechaVencimiento,
	}
}

type MovimientoPuntosRespuesta struct {
	ID               uint       `json:"id"`
	ClienteID        uint       `json:"cliente_id"`
	CompraID         *uint      `json:"compra_id"`
	CanjeID          *uint      `json:"canje_id"`
	Tipo             string     `json:"tipo"`
	Puntos           int        `json:"puntos"`
	Fecha            time.Time  `json:"fecha"`
	FechaVencimiento *time.Time `json:"fecha_vencimiento"`
}

type ConsultaPuntosRespuesta struct {
	ClienteID   uint                        `json:"cliente_id"`
	Cedula      string                      `json:"cedula"`
	Nombre      string                      `json:"nombre"`
	Saldo       int                         `json:"saldo"`
	Movimientos []MovimientoPuntosRespuesta `json:"movimientos"`
}

func nuevaConsultaPuntosRespuesta(r ConsultaPuntos) ConsultaPuntosRespuesta {
	movimientos := make([]MovimientoPuntosRespuesta, len(r.Movimientos))
	for i, m := range r.Movimientos {
		movimientos[i] = MovimientoPuntosRespuesta{ID: m.ID, ClienteID: m.ClienteID,
			CompraID: m.CompraID, CanjeID: m.CanjeID, Tipo: m.Tipo, Puntos: m.Puntos,
			Fecha: m.Fecha, FechaVencimiento: m.FechaVencimiento}
	}
	return ConsultaPuntosRespuesta{ClienteID: r.ClienteID, Cedula: r.Cedula,
		Nombre: r.Nombre, Saldo: r.Saldo, Movimientos: movimientos}
}
