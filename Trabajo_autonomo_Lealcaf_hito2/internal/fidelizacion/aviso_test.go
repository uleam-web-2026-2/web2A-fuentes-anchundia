package fidelizacion

import (
	"errors"
	"testing"
)

type entregaFalsa struct {
	propietario uint
	aviso       Aviso
}
type avisadorFalso struct {
	entregas  []entregaFalsa
	verificar func()
}

func (a *avisadorFalso) Enviar(id uint, aviso Aviso) {
	if a.verificar != nil {
		a.verificar()
	}
	a.entregas = append(a.entregas, entregaFalsa{id, aviso})
}

func TestCambioPermitidoAvisaTrasCommit(t *testing.T) {
	repo := &repositorioFalso{canje: Canje{ID: 9, ClienteID: 7, Estado: EstadoSolicitado, PuntosUtilizados: 40}}
	fake := &avisadorFalso{verificar: func() {
		if repo.canje.Estado != EstadoConfirmado || len(repo.movimientos) != 1 {
			t.Fatal("aviso antes del commit")
		}
	}}
	s := &Servicio{Repo: repo}
	s.AvisarCon(fake)
	if _, err := s.ActualizarEstadoCanje(9, EstadoConfirmado); err != nil {
		t.Fatal(err)
	}
	expected := entregaFalsa{7, Aviso{Tipo: "estado_cambiado", ID: 9, Estado: EstadoConfirmado}}
	if len(fake.entregas) != 1 || fake.entregas[0] != expected {
		t.Fatalf("aviso incorrecto: %+v", fake.entregas)
	}
}

func TestTransicionProhibidaNoAvisa(t *testing.T) {
	repo := &repositorioFalso{canje: Canje{ID: 9, ClienteID: 7, Estado: EstadoConfirmado}}
	fake := &avisadorFalso{}
	s := &Servicio{Repo: repo}
	s.AvisarCon(fake)
	if _, err := s.ActualizarEstadoCanje(9, EstadoCancelado); !errors.Is(err, ErrTransicionProhibida) {
		t.Fatal(err)
	}
	if len(fake.entregas) != 0 {
		t.Fatal("avisó una transición prohibida")
	}
}

func TestErrorPersistenciaNoAvisa(t *testing.T) {
	repo := &repositorioFalso{canje: Canje{ID: 9, ClienteID: 7, Estado: EstadoSolicitado, PuntosUtilizados: 40}, fallarMovimiento: errors.New("persistencia fallida")}
	fake := &avisadorFalso{}
	s := &Servicio{Repo: repo}
	s.AvisarCon(fake)
	if _, err := s.ActualizarEstadoCanje(9, EstadoConfirmado); err == nil {
		t.Fatal("se esperaba error")
	}
	if len(fake.entregas) != 0 || repo.canje.Estado != EstadoSolicitado {
		t.Fatal("aviso o escritura antes del rollback")
	}
}

type repoCommitFallido struct{ *repositorioFalso }

func (r *repoCommitFallido) EnTransaccion(f func(RepositorioTransaccion) error) error {
	tx := *r.repositorioFalso
	if err := f(&tx); err != nil {
		return err
	}
	return errors.New("commit fallido")
}
func TestErrorCommitNoAvisa(t *testing.T) {
	repo := &repoCommitFallido{&repositorioFalso{canje: Canje{ID: 9, ClienteID: 7, Estado: EstadoSolicitado, PuntosUtilizados: 40}}}
	fake := &avisadorFalso{}
	s := &Servicio{Repo: repo}
	s.AvisarCon(fake)
	if _, err := s.ActualizarEstadoCanje(9, EstadoConfirmado); err == nil {
		t.Fatal("se esperaba error de commit")
	}
	if len(fake.entregas) != 0 {
		t.Fatal("avisó un commit fallido")
	}
}
