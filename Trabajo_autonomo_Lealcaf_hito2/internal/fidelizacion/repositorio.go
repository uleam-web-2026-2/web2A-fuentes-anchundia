package fidelizacion

import (
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
)

// Repositorio contiene el acceso a datos; el servicio no construye SQL.
type Repositorio struct{ DB *gorm.DB }

func errorRegistro(err error, entidad string) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errorConMensaje(ErrNoEncontrado, entidad+" no encontrado")
	}
	return err
}
func (r *Repositorio) EnTransaccion(operacion func(*Repositorio) error) error {
	return r.DB.Transaction(func(tx *gorm.DB) error { return operacion(&Repositorio{DB: tx}) })
}
func (r *Repositorio) Cliente(id uint, bloquear bool) (Cliente, error) {
	var cliente Cliente
	consulta := r.DB
	if bloquear {
		consulta = consulta.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := consulta.First(&cliente, id).Error
	return cliente, errorRegistro(err, "cliente")
}
func (r *Repositorio) Sucursal(id uint) (Sucursal, error) {
	var sucursal Sucursal
	err := r.DB.First(&sucursal, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return sucursal, errorConMensaje(ErrNoEncontrado, "sucursal no encontrada")
	}
	return sucursal, err
}
func (r *Repositorio) RecompensaActiva(id uint) (Recompensa, error) {
	var recompensa Recompensa
	err := r.DB.Where("activa = ?", true).First(&recompensa, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return recompensa, errorConMensaje(ErrNoEncontrado, "recompensa no disponible")
	}
	return recompensa, err
}
func (r *Repositorio) Saldo(clienteID uint) (int, error) {
	var suma int64
	err := r.DB.Model(&MovimientoPuntos{}).Select("COALESCE(SUM(puntos), 0)").Where("cliente_id = ?", clienteID).Where("(tipo = ? AND fecha_vencimiento > ?) OR tipo = ?", TipoAcumulacion, ahora(), TipoCanje).Scan(&suma).Error
	if suma < 0 {
		suma = 0
	}
	return int(suma), err
}
func (r *Repositorio) Movimientos(clienteID uint) ([]MovimientoPuntos, error) {
	movimientos := make([]MovimientoPuntos, 0)
	err := r.DB.Where("cliente_id = ?", clienteID).Order("fecha DESC, id DESC").Find(&movimientos).Error
	return movimientos, err
}
func (r *Repositorio) RecompensasActivas() ([]Recompensa, error) {
	recompensas := make([]Recompensa, 0)
	err := r.DB.Where("activa = ?", true).Order("puntos_necesarios, id").Find(&recompensas).Error
	return recompensas, err
}
func (r *Repositorio) ExisteFactura(numero string) (bool, error) {
	var cantidad int64
	err := r.DB.Model(&Compra{}).Where("numero_factura = ?", numero).Count(&cantidad).Error
	return cantidad > 0, err
}
func (r *Repositorio) CrearCompra(compra *Compra) error {
	err := r.DB.Create(compra).Error
	if err == nil {
		return nil
	}
	// PostgreSQL informa SQLSTATE 23505; SQLite se utiliza en las pruebas.
	var estadoSQL interface{ SQLState() string }
	if errors.Is(err, gorm.ErrDuplicatedKey) || (errors.As(err, &estadoSQL) && estadoSQL.SQLState() == "23505") || strings.Contains(strings.ToLower(err.Error()), "unique constraint failed") {
		return ErrFacturaDuplicada
	}
	return err
}
func (r *Repositorio) CrearMovimiento(movimiento *MovimientoPuntos) error {
	return r.DB.Create(movimiento).Error
}
func (r *Repositorio) CrearCanje(canje *Canje) error { return r.DB.Create(canje).Error }
func (r *Repositorio) Canjes(estado string) ([]Canje, error) {
	canjes := make([]Canje, 0)
	consulta := r.DB.Order("fecha DESC, id DESC")
	if estado != "" {
		consulta = consulta.Where("estado = ?", estado)
	}
	err := consulta.Find(&canjes).Error
	return canjes, err
}
func (r *Repositorio) Canje(id uint, bloquear bool) (Canje, error) {
	var canje Canje
	consulta := r.DB
	if bloquear {
		consulta = consulta.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := consulta.First(&canje, id).Error
	return canje, errorRegistro(err, "canje")
}
func (r *Repositorio) CambiarEstadoCanje(id uint, desde, hasta string) error {
	resultado := r.DB.Model(&Canje{}).Where("id = ? AND estado = ?", id, desde).Update("estado", hasta)
	if resultado.Error != nil {
		return resultado.Error
	}
	if resultado.RowsAffected != 1 {
		return ErrTransicionProhibida
	}
	return nil
}
