package fidelizacion

import "time"

// Valores cerrados ("uno de:") que usa el dominio.
const (
	TipoAcumulacion = "acumulacion"
	TipoCanje       = "canje"

	EstadoSolicitado = "solicitado"
	EstadoConfirmado = "confirmado"
	EstadoCancelado  = "cancelado"

	RolCliente       = "cliente"
	RolAdministrador = "administrador"
)

// Reglas del programa (adaptadas del método de TI POINTS de DIPASO).
const (
	PuntosPorDolar = 1 // puntos que se ganan por cada dólar completo de la compra
	MesesVigencia  = 6 // vigencia de los puntos desde la compra que los generó
)

type Cliente struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Cedula        string    `gorm:"type:varchar(10);not null;unique" json:"cedula"`
	Nombre        string    `gorm:"not null" json:"nombre"`
	Correo        string    `gorm:"not null;unique" json:"correo"`
	Telefono      string    `json:"telefono"`
	FechaRegistro time.Time `gorm:"autoCreateTime" json:"fecha_registro"`
	Activo        bool      `gorm:"not null;default:true" json:"activo"`
}

type Sucursal struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Nombre    string `gorm:"not null;uniqueIndex" json:"nombre"`
	Direccion string `gorm:"not null" json:"direccion"`
	Activa    bool   `gorm:"not null;default:true" json:"activa"`
}

// Cuenta se usa exclusivamente para autenticacion, aparte del cliente del negocio.
type Cuenta struct {
	ID           uint     `gorm:"primaryKey" json:"id"`
	Usuario      string   `gorm:"not null;uniqueIndex" json:"usuario"`
	PasswordHash string   `gorm:"not null" json:"-"`
	Rol          string   `gorm:"not null" json:"rol"`
	ClienteID    *uint    `json:"cliente_id"`
	Cliente      *Cliente `gorm:"foreignKey:ClienteID" json:"-"`
}

type Compra struct {
	SucursalID    uint      `gorm:"not null;default:1;index" json:"sucursal_id"`
	Sucursal      Sucursal  `gorm:"foreignKey:SucursalID" json:"-"`
	ID            uint      `gorm:"primaryKey" json:"id"`
	ClienteID     uint      `gorm:"not null" json:"cliente_id"`
	Cliente       Cliente   `gorm:"foreignKey:ClienteID" json:"-"`
	Fecha         time.Time `gorm:"not null" json:"fecha"`
	Total         float64   `gorm:"not null" json:"total"`
	NumeroFactura string    `gorm:"not null;unique" json:"numero_factura"`
}

type Recompensa struct {
	ID               uint   `gorm:"primaryKey" json:"id"`
	Nombre           string `gorm:"not null" json:"nombre"`
	Descripcion      string `json:"descripcion"`
	PuntosNecesarios int    `gorm:"not null" json:"puntos_necesarios"`
	Activa           bool   `gorm:"not null;default:true" json:"activa"`
}

type Canje struct {
	SucursalID       uint       `gorm:"not null;default:1;index" json:"sucursal_id"`
	Sucursal         Sucursal   `gorm:"foreignKey:SucursalID" json:"-"`
	ID               uint       `gorm:"primaryKey" json:"id"`
	ClienteID        uint       `gorm:"not null" json:"cliente_id"`
	Cliente          Cliente    `gorm:"foreignKey:ClienteID" json:"-"`
	RecompensaID     uint       `gorm:"not null" json:"recompensa_id"`
	Recompensa       Recompensa `gorm:"foreignKey:RecompensaID" json:"-"`
	PuntosUtilizados int        `gorm:"not null" json:"puntos_utilizados"`
	Fecha            time.Time  `gorm:"not null" json:"fecha"`
	Estado           string     `gorm:"not null;default:solicitado" json:"estado"`
}

type MovimientoPuntos struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	ClienteID        uint       `gorm:"not null" json:"cliente_id"`
	Cliente          Cliente    `gorm:"foreignKey:ClienteID" json:"-"`
	CompraID         *uint      `json:"compra_id"`
	Compra           *Compra    `gorm:"foreignKey:CompraID" json:"-"`
	CanjeID          *uint      `json:"canje_id"`
	Canje            *Canje     `gorm:"foreignKey:CanjeID" json:"-"`
	Tipo             string     `gorm:"not null" json:"tipo"`
	Puntos           int        `gorm:"not null" json:"puntos"`
	Fecha            time.Time  `gorm:"not null" json:"fecha"`
	FechaVencimiento *time.Time `json:"fecha_vencimiento"`
}
