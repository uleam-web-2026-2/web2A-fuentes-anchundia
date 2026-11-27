package fidelizacion

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestPropietarioCanjeSaleDeSesion(t *testing.T) {
	e := nuevoEntorno(t)
	owner := e.cliente(t, "0900000001", "owner@example.invalid")
	other := e.cliente(t, "0900000002", "other@example.invalid")
	e.acumular(t, owner.ID, 100, time.Now().AddDate(0, 6, 0))
	reward := e.recompensa(t, "Café", 40)
	code, body := e.llamar("POST", "/canjes", RolCliente, int(owner.ID),
		fmt.Sprintf(`{"cliente_id":%d,"recompensa_id":%d,"sucursal_id":%d}`, other.ID, reward.ID, e.norte.ID))
	exigirCodigo(t, code, http.StatusCreated, body)
	if body["cliente_id"] != float64(owner.ID) {
		t.Fatalf("aceptó propietario del cuerpo: %v", body)
	}
	var stored Canje
	if err := e.db.First(&stored, idJSON(t, body)).Error; err != nil || stored.ClienteID != owner.ID {
		t.Fatalf("propietario persistido incorrecto: %+v %v", stored, err)
	}
	code, body = e.llamar("GET", fmt.Sprintf("/canjes/%d", stored.ID), RolCliente, int(other.ID), "")
	exigirCodigo(t, code, http.StatusForbidden, body)
}
