package fidelizacion

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func (e *entorno) solicitarCanje(clienteID, recompensaID, sucursalID uint) (int, map[string]any) {
	e.t.Helper()
	return e.llamar("POST", "/canjes", RolCliente, int(clienteID), fmt.Sprintf(`{"recompensa_id":%d,"sucursal_id":%d}`, recompensaID, sucursalID))
}

func TestIntegracionSeisPeticionesHito2(t *testing.T) {
	e := nuevoEntorno(t)
	c := e.cliente(t, "0900000001", "a@example.com")
	e.acumular(t, c.ID, 100, time.Now().AddDate(0, 6, 0))
	r := e.recompensa(t, "Café gratis", 40)
	e.cuenta("cliente-demo", "cliente-2026", RolCliente, c.ID)
	tokenAdmin := e.token(RolAdministrador, 0)
	codigo, canje := e.solicitarCanje(c.ID, r.ID, e.norte.ID)
	exigirCodigo(t, codigo, http.StatusCreated, canje)
	ruta := fmt.Sprintf("/canjes/%d", idJSON(t, canje))
	codigo, cuerpo := e.solicitar("GET", ruta, "", "")
	exigirCodigo(t, codigo, http.StatusUnauthorized, cuerpo)
	codigo, cuerpo = e.solicitar("POST", "/login", "", `{"usuario":"cliente-demo","password":"cliente-2026"}`)
	exigirCodigo(t, codigo, http.StatusOK, cuerpo)
	tokenCliente, ok := cuerpo["token"].(string)
	if !ok || tokenCliente == "" {
		t.Fatalf("login sin token: %v", cuerpo)
	}
	codigo, cuerpo = e.solicitar("GET", ruta, tokenCliente, "")
	exigirCodigo(t, codigo, http.StatusOK, cuerpo)
	codigo, cuerpo = e.solicitar("PATCH", ruta+"/estado", tokenCliente, `{"estado":"confirmado"}`)
	exigirCodigo(t, codigo, http.StatusForbidden, cuerpo)
	codigo, cuerpo = e.solicitar("PATCH", ruta+"/estado", tokenAdmin, `{"estado":"confirmado"}`)
	exigirCodigo(t, codigo, http.StatusOK, cuerpo)
	codigo, cuerpo = e.solicitar("PATCH", ruta+"/estado", tokenAdmin, `{"estado":"cancelado"}`)
	exigirCodigo(t, codigo, http.StatusUnprocessableEntity, cuerpo)
}

func TestPuntosCompartidosEntreSucursales(t *testing.T) {
	e := nuevoEntorno(t)
	c := e.cliente(t, "0900000001", "a@example.com")
	r := e.recompensa(t, "Café gratis", 40)
	codigo, cuerpo := e.llamar("POST", "/compras", RolAdministrador, 0,
		fmt.Sprintf(`{"cliente_id":%d,"total":100,"numero_factura":"001-001-000000001","sucursal_id":%d}`, c.ID, e.centro.ID))
	exigirCodigo(t, codigo, http.StatusCreated, cuerpo)
	var compra Compra
	if err := e.db.First(&compra).Error; err != nil || compra.SucursalID != e.centro.ID {
		t.Fatalf("compra sin sucursal Centro: %+v %v", compra, err)
	}
	codigo, canje := e.solicitarCanje(c.ID, r.ID, e.norte.ID)
	exigirCodigo(t, codigo, http.StatusCreated, canje)
	if canje["sucursal_id"] != float64(e.norte.ID) {
		t.Fatalf("el canje debe identificar Norte: %v", canje)
	}
	codigo, cuerpo = e.llamar("PATCH", fmt.Sprintf("/canjes/%d/estado", idJSON(t, canje)), RolAdministrador, 0, `{"estado":"confirmado"}`)
	exigirCodigo(t, codigo, http.StatusOK, cuerpo)
	_, puntos := e.llamar("GET", fmt.Sprintf("/clientes/%d/puntos", c.ID), RolCliente, int(c.ID), "")
	if puntos["saldo"] != float64(60) {
		t.Fatalf("puntos no compartidos entre sucursales: %v", puntos)
	}
}

func TestSucursalObligatoriaExistenteYActiva(t *testing.T) {
	for _, caso := range []struct {
		nombre   string
		id       uint
		inactiva bool
		esperado int
	}{
		{"sin_sucursal", 0, false, http.StatusBadRequest},
		{"sucursal_inexistente", 999, false, http.StatusNotFound},
		{"sucursal_inactiva", 1, true, http.StatusUnprocessableEntity},
	} {
		t.Run(caso.nombre, func(t *testing.T) {
			e := nuevoEntorno(t)
			c := e.cliente(t, "0900000001", "a@example.com")
			e.acumular(t, c.ID, 100, time.Now().AddDate(0, 6, 0))
			r := e.recompensa(t, "Café gratis", 40)
			if caso.inactiva {
				if err := e.db.Model(&e.centro).Update("activa", false).Error; err != nil {
					t.Fatal(err)
				}
			}
			codigo, cuerpo := e.solicitarCanje(c.ID, r.ID, caso.id)
			exigirCodigo(t, codigo, caso.esperado, cuerpo)
			codigo, cuerpo = e.llamar("POST", "/compras", RolAdministrador, 0,
				fmt.Sprintf(`{"cliente_id":%d,"total":50,"numero_factura":"001-001-000000002","sucursal_id":%d}`, c.ID, caso.id))
			exigirCodigo(t, codigo, caso.esperado, cuerpo)
			var compras, canjes int64
			e.db.Model(&Compra{}).Count(&compras)
			e.db.Model(&Canje{}).Count(&canjes)
			if compras != 0 || canjes != 0 {
				t.Fatalf("el rechazo guardó operaciones: compras=%d canjes=%d", compras, canjes)
			}
		})
	}
}

func TestCancelarSolicitadoNoDescuentaPuntos(t *testing.T) {
	e := nuevoEntorno(t)
	c := e.cliente(t, "0900000001", "a@example.com")
	e.acumular(t, c.ID, 100, time.Now().AddDate(0, 6, 0))
	r := e.recompensa(t, "Café gratis", 40)
	codigo, canje := e.solicitarCanje(c.ID, r.ID, e.centro.ID)
	exigirCodigo(t, codigo, http.StatusCreated, canje)
	ruta := fmt.Sprintf("/canjes/%d/estado", idJSON(t, canje))
	codigo, cuerpo := e.llamar("PATCH", ruta, RolAdministrador, 0, `{"estado":"cancelado"}`)
	exigirCodigo(t, codigo, http.StatusOK, cuerpo)
	codigo, cuerpo = e.llamar("PATCH", ruta, RolAdministrador, 0, `{"estado":"confirmado"}`)
	exigirCodigo(t, codigo, http.StatusUnprocessableEntity, cuerpo)
	_, puntos := e.llamar("GET", fmt.Sprintf("/clientes/%d/puntos", c.ID), RolCliente, int(c.ID), "")
	if puntos["saldo"] != float64(100) {
		t.Fatalf("cancelar descontó puntos: %v", puntos)
	}
}

func TestCanjeConfirmadoNoSeModificaNiSeDescuentaDosVeces(t *testing.T) {
	e := nuevoEntorno(t)
	c := e.cliente(t, "0900000001", "a@example.com")
	e.acumular(t, c.ID, 100, time.Now().AddDate(0, 6, 0))
	r := e.recompensa(t, "Café gratis", 40)
	codigo, canje := e.solicitarCanje(c.ID, r.ID, e.centro.ID)
	exigirCodigo(t, codigo, http.StatusCreated, canje)
	id := idJSON(t, canje)
	ruta := fmt.Sprintf("/canjes/%d/estado", id)
	codigo, cuerpo := e.llamar("PATCH", ruta, RolAdministrador, 0, `{"estado":"confirmado"}`)
	exigirCodigo(t, codigo, http.StatusOK, cuerpo)
	for _, destino := range []string{EstadoCancelado, EstadoConfirmado} {
		codigo, cuerpo = e.llamar("PATCH", ruta, RolAdministrador, 0, fmt.Sprintf(`{"estado":%q}`, destino))
		exigirCodigo(t, codigo, http.StatusUnprocessableEntity, cuerpo)
	}
	var guardado Canje
	if err := e.db.First(&guardado, id).Error; err != nil || guardado.Estado != EstadoConfirmado {
		t.Fatalf("rechazo alteró estado: %+v %v", guardado, err)
	}
	var cantidad int64
	e.db.Model(&MovimientoPuntos{}).Where("canje_id = ?", id).Count(&cantidad)
	if cantidad != 1 {
		t.Fatalf("debe existir un único descuento: %d", cantidad)
	}
	_, puntos := e.llamar("GET", fmt.Sprintf("/clientes/%d/puntos", c.ID), RolCliente, int(c.ID), "")
	if puntos["saldo"] != float64(60) {
		t.Fatalf("el rechazo modificó puntos: %v", puntos)
	}
}

func TestConfirmacionRevalidaSaldoYNoDejaCambiosParciales(t *testing.T) {
	e := nuevoEntorno(t)
	c := e.cliente(t, "0900000001", "a@example.com")
	e.acumular(t, c.ID, 60, time.Now().AddDate(0, 6, 0))
	r := e.recompensa(t, "Café gratis", 40)
	codigo, primero := e.solicitarCanje(c.ID, r.ID, e.centro.ID)
	exigirCodigo(t, codigo, http.StatusCreated, primero)
	codigo, segundo := e.solicitarCanje(c.ID, r.ID, e.centro.ID)
	exigirCodigo(t, codigo, http.StatusCreated, segundo)
	codigo, cuerpo := e.llamar("PATCH", fmt.Sprintf("/canjes/%d/estado", idJSON(t, primero)), RolAdministrador, 0, `{"estado":"confirmado"}`)
	exigirCodigo(t, codigo, http.StatusOK, cuerpo)
	idSegundo := idJSON(t, segundo)
	codigo, cuerpo = e.llamar("PATCH", fmt.Sprintf("/canjes/%d/estado", idSegundo), RolAdministrador, 0, `{"estado":"confirmado"}`)
	exigirCodigo(t, codigo, http.StatusUnprocessableEntity, cuerpo)
	var guardado Canje
	if err := e.db.First(&guardado, idSegundo).Error; err != nil || guardado.Estado != EstadoSolicitado {
		t.Fatalf("el rechazo debe conservar solicitado: %+v %v", guardado, err)
	}
	var cantidad int64
	e.db.Model(&MovimientoPuntos{}).Where("canje_id = ?", idSegundo).Count(&cantidad)
	if cantidad != 0 {
		t.Fatal("confirmación rechazada descontó puntos")
	}
}

func TestEstadosYDatosInvalidos(t *testing.T) {
	e := nuevoEntorno(t)
	c := e.cliente(t, "0900000001", "a@example.com")
	e.acumular(t, c.ID, 100, time.Now().AddDate(0, 6, 0))
	r := e.recompensa(t, "Café gratis", 40)
	codigo, canje := e.solicitarCanje(c.ID, r.ID, e.centro.ID)
	exigirCodigo(t, codigo, http.StatusCreated, canje)
	ruta := fmt.Sprintf("/canjes/%d/estado", idJSON(t, canje))
	for _, p := range []struct{ metodo, ruta, cuerpo string }{
		{"PATCH", ruta, `{"estado":"inventado"}`},
		{"PATCH", ruta, `{`},
		{"PATCH", "/canjes/no-es-id/estado", `{"estado":"confirmado"}`},
		{"GET", "/canjes?estado=inventado", ""},
	} {
		codigo, cuerpo := e.llamar(p.metodo, p.ruta, RolAdministrador, 0, p.cuerpo)
		exigirCodigo(t, codigo, http.StatusBadRequest, cuerpo)
	}
}
