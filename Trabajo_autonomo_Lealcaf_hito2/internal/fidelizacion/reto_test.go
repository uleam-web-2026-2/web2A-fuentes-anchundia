package fidelizacion

import (
	"errors"
	"testing"
	"time"
)

func TestRetoAvisaADosAdministradoresYNingunCliente(t *testing.T) {
	e, a, url := servidorAvisos(t)
	owner := e.cliente(t, "0900000001", "owner@example.invalid")
	other := e.cliente(t, "0900000002", "other@example.invalid")
	e.acumular(t, owner.ID, 100, time.Now().AddDate(0, 6, 0))
	reward := e.recompensa(t, "Café del reto", 40)
	admin1 := e.cuenta("admin1", "ficticia", RolAdministrador, 0)
	admin2 := e.cuenta("admin2", "ficticia", RolAdministrador, 0)
	token1, err := e.auth.EmitirToken(admin1)
	if err != nil {
		t.Fatal(err)
	}
	token2, err := e.auth.EmitirToken(admin2)
	if err != nil {
		t.Fatal(err)
	}
	c1, _, err := abrirWS(t, url, token1, OrigenFront)
	if err != nil {
		t.Fatal(err)
	}
	defer c1.CloseNow()
	c2, _, err := abrirWS(t, url, token2, OrigenFront)
	if err != nil {
		t.Fatal(err)
	}
	defer c2.CloseNow()
	own, _, err := abrirWS(t, url, e.token(RolCliente, owner.ID), OrigenFront)
	if err != nil {
		t.Fatal(err)
	}
	defer own.CloseNow()
	foreign, _, err := abrirWS(t, url, e.token(RolCliente, other.ID), OrigenFront)
	if err != nil {
		t.Fatal(err)
	}
	defer foreign.CloseNow()
	esperarConexiones(t, a, 4)
	s := &Servicio{Repo: &RepositorioGORM{DB: e.db}}
	s.AvisarCon(a)
	s.AvisarNuevosCon(a)
	canje, err := s.SolicitarCanje(owner.ID, EntradaCanje{RecompensaID: reward.ID, SucursalID: e.norte.ID})
	if err != nil {
		t.Fatal(err)
	}
	expected := Aviso{Tipo: "canje_nuevo", ID: canje.ID, Estado: EstadoSolicitado}
	leerAviso(t, c1, expected)
	leerAviso(t, c2, expected)
	sinAviso(t, own)
	sinAviso(t, foreign)
}

type falsoNuevos struct{ avisos []Aviso }

func (a *falsoNuevos) NuevoCanje(av Aviso) { a.avisos = append(a.avisos, av) }

type repoNuevo struct {
	*repositorioFalso
	fallo error
}

func (r *repoNuevo) RecompensaActiva(id uint) (Recompensa, error) {
	return Recompensa{ID: id, PuntosNecesarios: 40, Activa: true}, nil
}
func (r *repoNuevo) CrearCanje(c *Canje) error {
	if r.fallo != nil {
		return r.fallo
	}
	c.ID = 12
	r.canje = *c
	return nil
}
func TestRetoNoAvisaSiCrearFalla(t *testing.T) {
	repo := &repoNuevo{repositorioFalso: &repositorioFalso{}, fallo: errors.New("fallo de escritura")}
	a := &falsoNuevos{}
	s := &Servicio{Repo: repo}
	s.AvisarNuevosCon(a)
	if _, err := s.SolicitarCanje(1, EntradaCanje{RecompensaID: 1, SucursalID: 1}); err == nil {
		t.Fatal("se esperaba error")
	}
	if len(a.avisos) != 0 {
		t.Fatal("avisó un canje no guardado")
	}
	repo.fallo = nil
	if _, err := s.SolicitarCanje(1, EntradaCanje{RecompensaID: 1, SucursalID: 1}); err != nil {
		t.Fatal(err)
	}
	if len(a.avisos) != 1 || a.avisos[0].ID != 12 {
		t.Fatalf("aviso incorrecto: %+v", a.avisos)
	}
}

func TestRetoEstadoSoloAlPropietarioSinAvisarAdministradores(t *testing.T) {
	e, a, url := servidorAvisos(t)
	owner := e.cliente(t, "0900000001", "owner@example.invalid")
	own, _, err := abrirWS(t, url, e.token(RolCliente, owner.ID), OrigenFront)
	if err != nil {
		t.Fatal(err)
	}
	defer own.CloseNow()
	admin, _, err := abrirWS(t, url, e.token(RolAdministrador, 0), OrigenFront)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.CloseNow()
	esperarConexiones(t, a, 2)
	repo := &repositorioFalso{canje: Canje{ID: 9, ClienteID: owner.ID, Estado: EstadoSolicitado, PuntosUtilizados: 40}}
	s := &Servicio{Repo: repo}
	s.AvisarCon(a)
	s.AvisarNuevosCon(a)
	if _, err := s.ActualizarEstadoCanje(9, EstadoConfirmado); err != nil {
		t.Fatal(err)
	}
	leerAviso(t, own, Aviso{Tipo: "estado_cambiado", ID: 9, Estado: EstadoConfirmado})
	sinAviso(t, admin)
}
