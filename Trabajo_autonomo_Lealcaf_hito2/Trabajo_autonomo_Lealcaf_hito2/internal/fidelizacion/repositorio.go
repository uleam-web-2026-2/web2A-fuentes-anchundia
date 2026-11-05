package fidelizacion

// Repositorio declara lo que usa el servicio, sin depender de una base concreta.
type Repositorio interface {
	Cliente(id uint, bloquear bool) (Cliente, error)
	Sucursal(id uint) (Sucursal, error)
	RecompensaActiva(id uint) (Recompensa, error)
	Saldo(clienteID uint) (int, error)
	Movimientos(clienteID uint) ([]MovimientoPuntos, error)
	RecompensasActivas() ([]Recompensa, error)
	ExisteFactura(numero string) (bool, error)
	CrearCompra(compra *Compra) error
	CrearMovimiento(movimiento *MovimientoPuntos) error
	CrearCanje(canje *Canje) error
	Canjes(estado string) ([]Canje, error)
	Canje(id uint, bloquear bool) (Canje, error)
	CambiarEstadoCanje(id uint, desde, hasta string) error
	EnTransaccion(operacion func(RepositorioTransaccion) error) error
}

// RepositorioTransaccion limita el callback a las operaciones de los dos casos
// atómicos. Cada adaptador las ejecuta sobre su propia transacción.
type RepositorioTransaccion interface {
	Cliente(id uint, bloquear bool) (Cliente, error)
	Sucursal(id uint) (Sucursal, error)
	Saldo(clienteID uint) (int, error)
	ExisteFactura(numero string) (bool, error)
	CrearCompra(compra *Compra) error
	CrearMovimiento(movimiento *MovimientoPuntos) error
	Canje(id uint, bloquear bool) (Canje, error)
	CambiarEstadoCanje(id uint, desde, hasta string) error
}
