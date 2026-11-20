package fidelizacion

import (
	"errors"
	"net/http"
	"testing"
)

func TestFiltroConComillaEsUnValor(t *testing.T) {
	e := nuevoEntorno(t)
	client := e.cliente(t, "0900000088", "sql@example.invalid")
	reward := e.recompensa(t, "Café SQL", 40)
	canje := Canje{ClienteID: client.ID, RecompensaID: reward.ID, SucursalID: e.norte.ID, Estado: EstadoSolicitado}
	if err := e.db.Create(&canje).Error; err != nil {
		t.Fatal(err)
	}
	repo := &RepositorioGORM{DB: e.db}
	normal, err := repo.Canjes(EstadoSolicitado)
	if err != nil || len(normal) != 1 {
		t.Fatalf("filtro normal: cantidad=%d error=%v", len(normal), err)
	}
	attack, err := repo.Canjes("' OR 1=1 --")
	if err != nil || len(attack) != 0 {
		t.Fatalf("el texto alteró el filtro: cantidad=%d error=%v", len(attack), err)
	}
	_, err = (&Servicio{Repo: repo}).ListarCanjes("' OR 1=1 --")
	if !errors.Is(err, ErrEstadoInvalido) {
		t.Fatalf("el servicio debe validar el estado: %v", err)
	}
}

func TestIDDeRutaRechazaCondicionSQL(t *testing.T) {
	e := nuevoEntorno(t)
	code, body := e.solicitar("GET", "/canjes/1%20OR%201=1", e.token(RolAdministrador, 0), "")
	exigirCodigo(t, code, http.StatusBadRequest, body)
}
