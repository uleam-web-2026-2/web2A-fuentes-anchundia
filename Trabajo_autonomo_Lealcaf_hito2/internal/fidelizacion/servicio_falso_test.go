package fidelizacion

import (
	"errors"
	"testing"
)

// Este falso usa una copia del estado; no abre ninguna base de datos.
type repositorioFalso struct {
	Repositorio
	canje            Canje
	compras          []Compra
	movimientos      []MovimientoPuntos
	fallarMovimiento error
	transacciones    int
}

func (r *repositorioFalso) Cliente(id uint, bloquear bool) (Cliente, error) {
	return Cliente{ID: id, Activo: true}, nil
}
func (r *repositorioFalso) Sucursal(id uint) (Sucursal, error) {
	return Sucursal{ID: id, Activa: true}, nil
}
func (r *repositorioFalso) Saldo(uint) (int, error)            { return 100, nil }
func (r *repositorioFalso) ExisteFactura(string) (bool, error) { return false, nil }
func (r *repositorioFalso) Canje(uint, bool) (Canje, error)    { return r.canje, nil }
func (r *repositorioFalso) CambiarEstadoCanje(id uint, desde, hasta string) error {
	if r.canje.ID != id || r.canje.Estado != desde {
		return ErrTransicionProhibida
	}
	r.canje.Estado = hasta
	return nil
}
func (r *repositorioFalso) CrearCompra(c *Compra) error {
	c.ID = uint(len(r.compras) + 1)
	r.compras = append(r.compras, *c)
	return nil
}
func (r *repositorioFalso) CrearMovimiento(m *MovimientoPuntos) error {
	if r.fallarMovimiento != nil {
		return r.fallarMovimiento
	}
	r.movimientos = append(r.movimientos, *m)
	return nil
}
func (r *repositorioFalso) EnTransaccion(operacion func(RepositorioTransaccion) error) error {
	r.transacciones++
	tx := *r
	tx.compras = append([]Compra(nil), r.compras...)
	tx.movimientos = append([]MovimientoPuntos(nil), r.movimientos...)
	if err := operacion(&tx); err != nil {
		return err
	}
	r.canje, r.compras, r.movimientos = tx.canje, tx.compras, tx.movimientos
	return nil
}

func TestServicioConfirmacionConFalsoSinBase(t *testing.T) {
	fallo := errors.New("fallo al insertar el movimiento")
	repo := &repositorioFalso{canje: Canje{ID: 9, ClienteID: 1, PuntosUtilizados: 40, Estado: EstadoSolicitado}, fallarMovimiento: fallo}
	servicio := &Servicio{Repo: repo}
	if _, err := servicio.ActualizarEstadoCanje(9, EstadoConfirmado); !errors.Is(err, fallo) {
		t.Fatalf("se perdió el error del repositorio: %v", err)
	}
	if repo.canje.Estado != EstadoSolicitado || len(repo.movimientos) != 0 || repo.transacciones != 1 {
		t.Fatalf("confirmación parcial: %+v", repo)
	}
	repo.fallarMovimiento = nil
	canje, err := servicio.ActualizarEstadoCanje(9, EstadoConfirmado)
	if err != nil || canje.Estado != EstadoConfirmado || len(repo.movimientos) != 1 || repo.movimientos[0].Puntos != -40 {
		t.Fatalf("confirmación incorrecta: %+v %v", canje, err)
	}
	if _, err := servicio.ActualizarEstadoCanje(9, EstadoConfirmado); !errors.Is(err, ErrTransicionProhibida) || len(repo.movimientos) != 1 {
		t.Fatalf("descontó dos veces: %v", err)
	}
}

func TestServicioCompraConFalsoSinBase(t *testing.T) {
	fallo := errors.New("fallo al acreditar puntos")
	repo := &repositorioFalso{fallarMovimiento: fallo}
	servicio := &Servicio{Repo: repo}
	entrada := EntradaCompra{ClienteID: 1, SucursalID: 1, Total: 50, NumeroFactura: "FAKE-1"}
	if _, err := servicio.RegistrarCompra(entrada); !errors.Is(err, fallo) {
		t.Fatal(err)
	}
	if len(repo.compras) != 0 || len(repo.movimientos) != 0 {
		t.Fatal("quedó una compra sin sus puntos")
	}
	repo.fallarMovimiento = nil
	resultado, err := servicio.RegistrarCompra(entrada)
	if err != nil || resultado.PuntosGenerados != 50 || len(repo.compras) != 1 || len(repo.movimientos) != 1 {
		t.Fatalf("compra incorrecta: %+v %v", resultado, err)
	}
}
