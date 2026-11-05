package fidelizacion

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

// ahora permite controlar el reloj en las pruebas.
var ahora = time.Now
var (
	ErrDatosInvalidos      = errors.New("datos inválidos")
	ErrEstadoInvalido      = errors.New("estado no válido")
	ErrNoEncontrado        = errors.New("registro no encontrado")
	ErrProhibido           = errors.New("acción no permitida")
	ErrFacturaDuplicada    = errors.New("esa factura ya fue registrada")
	ErrTransicionProhibida = errors.New("transición no permitida")
	ErrPuntosInsuficientes = errors.New("puntos insuficientes")
	ErrSucursalInactiva    = errors.New("la sucursal está inactiva")
	ErrClienteInactivo     = errors.New("el cliente está inactivo")
)

// ErrorDominio conserva el motivo del negocio sin conocer códigos HTTP.
type ErrorDominio struct {
	Causa      error
	Mensaje    string
	Saldo      *int
	Necesarios *int
}

func (e *ErrorDominio) Error() string { return e.Mensaje }
func (e *ErrorDominio) Unwrap() error { return e.Causa }
func errorConMensaje(causa error, mensaje string) error {
	return &ErrorDominio{Causa: causa, Mensaje: mensaje}
}
func puntosInsuficientes(saldo, necesarios int) error {
	return &ErrorDominio{Causa: ErrPuntosInsuficientes, Mensaje: ErrPuntosInsuficientes.Error(), Saldo: &saldo, Necesarios: &necesarios}
}

// ValidarTransicion es la máquina de estados pura; no depende de HTTP ni de la base.
func ValidarTransicion(actual, destino string) error {
	if destino != EstadoSolicitado && destino != EstadoConfirmado && destino != EstadoCancelado {
		return ErrEstadoInvalido
	}
	if actual != EstadoSolicitado || destino == EstadoSolicitado {
		return fmt.Errorf("%w: de %s a %s", ErrTransicionProhibida, actual, destino)
	}
	return nil
}

type Servicio struct{ Repo Repositorio }
type EntradaCompra struct {
	ClienteID     uint
	SucursalID    uint
	Total         float64
	NumeroFactura string
}
type EntradaCanje struct {
	RecompensaID uint
	SucursalID   uint
}
type ResultadoCompra struct {
	Compra           Compra
	PuntosGenerados  int
	FechaVencimiento time.Time
}
type ConsultaPuntos struct {
	ClienteID   uint
	Cedula      string
	Nombre      string
	Saldo       int
	Movimientos []MovimientoPuntos
}

func (s *Servicio) ConsultarPuntos(id uint, rol string, clienteID uint) (ConsultaPuntos, error) {
	var resultado ConsultaPuntos
	if rol == RolCliente && clienteID != id {
		return resultado, errorConMensaje(ErrProhibido, "solo puede consultar sus propios puntos")
	}
	cliente, err := s.Repo.Cliente(id, false)
	if err != nil {
		return resultado, err
	}
	saldo, err := s.Repo.Saldo(id)
	if err != nil {
		return resultado, err
	}
	movimientos, err := s.Repo.Movimientos(id)
	if err != nil {
		return resultado, err
	}
	return ConsultaPuntos{ClienteID: cliente.ID, Cedula: cliente.Cedula, Nombre: cliente.Nombre, Saldo: saldo, Movimientos: movimientos}, nil
}
func (s *Servicio) ListarRecompensas() ([]Recompensa, error) { return s.Repo.RecompensasActivas() }
func (s *Servicio) sucursalActiva(id uint) error {
	return validarSucursalActiva(s.Repo, id)
}

func validarSucursalActiva(repo interface{ Sucursal(uint) (Sucursal, error) }, id uint) error {
	sucursal, err := repo.Sucursal(id)
	if err != nil {
		return err
	}
	if !sucursal.Activa {
		return ErrSucursalInactiva
	}
	return nil
}

func (s *Servicio) RegistrarCompra(entrada EntradaCompra) (ResultadoCompra, error) {
	var resultado ResultadoCompra
	entrada.NumeroFactura = strings.TrimSpace(entrada.NumeroFactura)
	// Evita desbordar int al convertir un monto recibido por JSON a puntos.
	if entrada.ClienteID == 0 || entrada.SucursalID == 0 || entrada.NumeroFactura == "" || entrada.Total <= 0 || math.IsNaN(entrada.Total) || math.IsInf(entrada.Total, 0) || entrada.Total >= float64(int(^uint(0)>>1)) {
		return resultado, ErrDatosInvalidos
	}
	err := s.Repo.EnTransaccion(func(repo RepositorioTransaccion) error {
		cliente, err := repo.Cliente(entrada.ClienteID, true)
		if err != nil {
			return err
		}
		if !cliente.Activo {
			return ErrClienteInactivo
		}
		if err := validarSucursalActiva(repo, entrada.SucursalID); err != nil {
			return err
		}
		existe, err := repo.ExisteFactura(entrada.NumeroFactura)
		if err != nil {
			return err
		}
		if existe {
			return ErrFacturaDuplicada
		}
		fecha := ahora()
		vence := fecha.AddDate(0, MesesVigencia, 0)
		puntos := int(entrada.Total) * PuntosPorDolar
		compra := Compra{ClienteID: entrada.ClienteID, SucursalID: entrada.SucursalID, Fecha: fecha, Total: entrada.Total, NumeroFactura: entrada.NumeroFactura}
		if err := repo.CrearCompra(&compra); err != nil {
			return err
		}
		if puntos > 0 {
			movimiento := MovimientoPuntos{ClienteID: entrada.ClienteID, CompraID: &compra.ID, Tipo: TipoAcumulacion, Puntos: puntos, Fecha: fecha, FechaVencimiento: &vence}
			if err := repo.CrearMovimiento(&movimiento); err != nil {
				return err
			}
		}
		resultado = ResultadoCompra{Compra: compra, PuntosGenerados: puntos, FechaVencimiento: vence}
		return nil
	})
	return resultado, err
}

func (s *Servicio) SolicitarCanje(clienteID uint, entrada EntradaCanje) (Canje, error) {
	var canje Canje
	if clienteID == 0 || entrada.RecompensaID == 0 || entrada.SucursalID == 0 {
		return canje, ErrDatosInvalidos
	}
	cliente, err := s.Repo.Cliente(clienteID, false)
	if err != nil {
		return canje, err
	}
	if !cliente.Activo {
		return canje, ErrClienteInactivo
	}
	if err := s.sucursalActiva(entrada.SucursalID); err != nil {
		return canje, err
	}
	recompensa, err := s.Repo.RecompensaActiva(entrada.RecompensaID)
	if err != nil {
		return canje, err
	}
	saldo, err := s.Repo.Saldo(clienteID)
	if err != nil {
		return canje, err
	}
	if saldo < recompensa.PuntosNecesarios {
		return canje, puntosInsuficientes(saldo, recompensa.PuntosNecesarios)
	}
	canje = Canje{ClienteID: clienteID, SucursalID: entrada.SucursalID, RecompensaID: recompensa.ID, PuntosUtilizados: recompensa.PuntosNecesarios, Fecha: ahora(), Estado: EstadoSolicitado}
	if err := s.Repo.CrearCanje(&canje); err != nil {
		return Canje{}, err
	}
	return canje, nil
}

func (s *Servicio) ListarCanjes(estado string) ([]Canje, error) {
	if estado != "" && estado != EstadoSolicitado && estado != EstadoConfirmado && estado != EstadoCancelado {
		return nil, ErrEstadoInvalido
	}
	return s.Repo.Canjes(estado)
}
func (s *Servicio) ObtenerCanje(id uint, rol string, clienteID uint) (Canje, error) {
	canje, err := s.Repo.Canje(id, false)
	if err != nil {
		return canje, err
	}
	if rol == RolCliente && clienteID != canje.ClienteID {
		return Canje{}, errorConMensaje(ErrProhibido, "solo puede consultar sus propios canjes")
	}
	return canje, nil
}

func (s *Servicio) ActualizarEstadoCanje(id uint, destino string) (Canje, error) {
	var resultado Canje
	if destino != EstadoSolicitado && destino != EstadoConfirmado && destino != EstadoCancelado {
		return resultado, ErrEstadoInvalido
	}
	err := s.Repo.EnTransaccion(func(repo RepositorioTransaccion) error {
		canje, err := repo.Canje(id, false)
		if err != nil {
			return err
		}
		// Todos los canjes del cliente usan el mismo bloqueo, para no gastar dos veces el saldo.
		cliente, err := repo.Cliente(canje.ClienteID, true)
		if err != nil {
			return err
		}
		canje, err = repo.Canje(id, true)
		if err != nil {
			return err
		}
		if err := ValidarTransicion(canje.Estado, destino); err != nil {
			return err
		}
		if destino == EstadoConfirmado {
			if !cliente.Activo {
				return ErrClienteInactivo
			}
			saldo, err := repo.Saldo(canje.ClienteID)
			if err != nil {
				return err
			}
			if saldo < canje.PuntosUtilizados {
				return puntosInsuficientes(saldo, canje.PuntosUtilizados)
			}
		}
		if err := repo.CambiarEstadoCanje(canje.ID, EstadoSolicitado, destino); err != nil {
			return err
		}
		if destino == EstadoConfirmado {
			movimiento := MovimientoPuntos{ClienteID: canje.ClienteID, CanjeID: &canje.ID, Tipo: TipoCanje, Puntos: -canje.PuntosUtilizados, Fecha: ahora()}
			if err := repo.CrearMovimiento(&movimiento); err != nil {
				return err
			}
		}
		canje.Estado = destino
		resultado = canje
		return nil
	})
	return resultado, err
}
