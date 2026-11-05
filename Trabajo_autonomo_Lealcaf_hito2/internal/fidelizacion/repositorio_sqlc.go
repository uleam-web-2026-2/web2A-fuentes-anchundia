package fidelizacion

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"fidelizacion/internal/almacen/filas"
)

// RepositorioSqlc consulta canjes y actualiza su estado mediante sqlc.
// Las demás operaciones fuera de una transacción usan GORM.
type RepositorioSqlc struct {
	*RepositorioGORM
	conexion *sql.DB
	q        *filas.Queries
}

var _ Repositorio = (*RepositorioSqlc)(nil)

func NuevoRepositorioSqlc(g *RepositorioGORM, conexion *sql.DB) *RepositorioSqlc {
	return &RepositorioSqlc{RepositorioGORM: g, conexion: conexion, q: filas.New(conexion)}
}

func (r *RepositorioSqlc) Canje(id uint, bloquear bool) (Canje, error) {
	return buscarCanjeSQL(r.q, id, bloquear)
}

func (r *RepositorioSqlc) CambiarEstadoCanje(id uint, desde, hasta string) error {
	return cambiarEstadoSQL(r.q, id, desde, hasta)
}

func (r *RepositorioSqlc) EnTransaccion(operacion func(RepositorioTransaccion) error) error {
	tx, err := r.conexion.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// WithTx enlaza las consultas del callback a la misma transacción SQL.
	repo := &repositorioTransaccionSqlc{q: r.q.WithTx(tx)}
	if err := operacion(repo); err != nil {
		return err
	}
	return tx.Commit()
}

// repositorioTransaccionSqlc ejecuta las consultas del callback sobre la conexión transaccional.
type repositorioTransaccionSqlc struct{ q *filas.Queries }

var _ RepositorioTransaccion = (*repositorioTransaccionSqlc)(nil)

func errorRegistroSQL(err error, entidad string) error {
	if errors.Is(err, sql.ErrNoRows) {
		return errorConMensaje(ErrNoEncontrado, entidad+" no encontrado")
	}
	return err
}

func buscarCanjeSQL(q *filas.Queries, id uint, bloquear bool) (Canje, error) {
	var f filas.Canje
	var err error
	if bloquear {
		f, err = q.BloquearCanje(context.Background(), int64(id))
	} else {
		f, err = q.BuscarCanje(context.Background(), int64(id))
	}
	if err != nil {
		return Canje{}, errorRegistroSQL(err, "canje")
	}
	return Canje{ID: uint(f.ID), SucursalID: uint(f.SucursalID), ClienteID: uint(f.ClienteID),
		RecompensaID: uint(f.RecompensaID), PuntosUtilizados: int(f.PuntosUtilizados),
		Fecha: f.Fecha, Estado: f.Estado}, nil
}

func cambiarEstadoSQL(q *filas.Queries, id uint, desde, hasta string) error {
	filasAfectadas, err := q.CambiarEstadoCanje(context.Background(), filas.CambiarEstadoCanjeParams{
		ID: int64(id), Desde: desde, Hasta: hasta,
	})
	if err != nil {
		return err
	}
	if filasAfectadas != 1 {
		return ErrTransicionProhibida
	}
	return nil
}

func (r *repositorioTransaccionSqlc) Canje(id uint, bloquear bool) (Canje, error) {
	return buscarCanjeSQL(r.q, id, bloquear)
}
func (r *repositorioTransaccionSqlc) CambiarEstadoCanje(id uint, desde, hasta string) error {
	return cambiarEstadoSQL(r.q, id, desde, hasta)
}
func (r *repositorioTransaccionSqlc) Cliente(id uint, bloquear bool) (Cliente, error) {
	if bloquear {
		f, err := r.q.BloquearClienteTransaccion(context.Background(), int64(id))
		return Cliente{ID: uint(f.ID), Activo: f.Activo}, errorRegistroSQL(err, "cliente")
	}
	f, err := r.q.BuscarClienteTransaccion(context.Background(), int64(id))
	return Cliente{ID: uint(f.ID), Activo: f.Activo}, errorRegistroSQL(err, "cliente")
}
func (r *repositorioTransaccionSqlc) Sucursal(id uint) (Sucursal, error) {
	f, err := r.q.BuscarSucursalTransaccion(context.Background(), int64(id))
	if errors.Is(err, sql.ErrNoRows) {
		return Sucursal{}, errorConMensaje(ErrNoEncontrado, "sucursal no encontrada")
	}
	return Sucursal{ID: uint(f.ID), Nombre: f.Nombre, Direccion: f.Direccion, Activa: f.Activa}, err
}
func (r *repositorioTransaccionSqlc) Saldo(clienteID uint) (int, error) {
	saldo, err := r.q.SaldoTransaccion(context.Background(), filas.SaldoTransaccionParams{
		ClienteID: int64(clienteID), Fecha: sql.NullTime{Time: ahora(), Valid: true},
	})
	if saldo < 0 {
		saldo = 0
	}
	return int(saldo), err
}
func (r *repositorioTransaccionSqlc) ExisteFactura(numero string) (bool, error) {
	return r.q.ExisteFacturaTransaccion(context.Background(), numero)
}
func (r *repositorioTransaccionSqlc) CrearCompra(compra *Compra) error {
	id, err := r.q.CrearCompraTransaccion(context.Background(), filas.CrearCompraTransaccionParams{
		SucursalID: int64(compra.SucursalID), ClienteID: int64(compra.ClienteID), Fecha: compra.Fecha,
		Total: strconv.FormatFloat(compra.Total, 'f', -1, 64), NumeroFactura: compra.NumeroFactura,
	})
	if err != nil {
		var estadoSQL interface{ SQLState() string }
		if errors.As(err, &estadoSQL) && estadoSQL.SQLState() == "23505" {
			return ErrFacturaDuplicada
		}
		return err
	}
	compra.ID = uint(id)
	return nil
}
func (r *repositorioTransaccionSqlc) CrearMovimiento(movimiento *MovimientoPuntos) error {
	var compraID, canjeID sql.NullInt64
	var vencimiento sql.NullTime
	if movimiento.CompraID != nil {
		compraID = sql.NullInt64{Int64: int64(*movimiento.CompraID), Valid: true}
	}
	if movimiento.CanjeID != nil {
		canjeID = sql.NullInt64{Int64: int64(*movimiento.CanjeID), Valid: true}
	}
	if movimiento.FechaVencimiento != nil {
		vencimiento = sql.NullTime{Time: *movimiento.FechaVencimiento, Valid: true}
	}
	id, err := r.q.CrearMovimientoTransaccion(context.Background(), filas.CrearMovimientoTransaccionParams{
		ClienteID: int64(movimiento.ClienteID), CompraID: compraID, CanjeID: canjeID,
		Tipo: movimiento.Tipo, Puntos: int64(movimiento.Puntos), Fecha: movimiento.Fecha, FechaVencimiento: vencimiento,
	})
	if err != nil {
		return err
	}
	movimiento.ID = uint(id)
	return nil
}
