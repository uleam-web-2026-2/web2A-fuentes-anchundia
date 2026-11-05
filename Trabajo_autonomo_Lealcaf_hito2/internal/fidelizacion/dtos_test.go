package fidelizacion

import (
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"testing"
	"time"
)

func exigirCamposDTO(t *testing.T, cuerpo map[string]any, campos ...string) {
	t.Helper()
	obtenidos := make([]string, 0, len(cuerpo))
	for campo := range cuerpo {
		obtenidos = append(obtenidos, campo)
	}
	sort.Strings(obtenidos)
	sort.Strings(campos)
	if !reflect.DeepEqual(obtenidos, campos) {
		t.Fatalf("cambió el contrato JSON: campos=%v, esperados=%v", obtenidos, campos)
	}
}

func exigirCanjeDTO(t *testing.T, cuerpo map[string]any) {
	t.Helper()
	exigirCamposDTO(t, cuerpo, "id", "sucursal_id", "cliente_id", "recompensa_id",
		"puntos_utilizados", "fecha", "estado")
}

func TestDTOCanjeIgnoraCamposQueDecideElServidor(t *testing.T) {
	e := nuevoEntorno(t)
	cliente := e.cliente(t, "0900000001", "dto@example.com")
	e.acumular(t, cliente.ID, 100, time.Now().AddDate(0, 6, 0))
	recompensa := e.recompensa(t, "Café DTO", 40)
	cuerpo := fmt.Sprintf(`{"recompensa_id":%d,"sucursal_id":%d,"id":77,"estado":"confirmado","cliente_id":999,"puntos_utilizados":1,"fecha":"2000-01-01T00:00:00Z"}`, recompensa.ID, e.centro.ID)
	codigo, creado := e.llamar("POST", "/canjes", RolCliente, int(cliente.ID), cuerpo)
	exigirCodigo(t, codigo, http.StatusCreated, creado)
	exigirCanjeDTO(t, creado)
	id := idJSON(t, creado)
	if id == 77 || creado["estado"] != EstadoSolicitado || creado["cliente_id"] != float64(cliente.ID) || creado["puntos_utilizados"] != float64(40) {
		t.Fatalf("el cliente controló campos internos: %v", creado)
	}
	var guardado Canje
	if err := e.db.First(&guardado, id).Error; err != nil {
		t.Fatal(err)
	}
	if guardado.Estado != EstadoSolicitado || guardado.ClienteID != cliente.ID || guardado.PuntosUtilizados != 40 || guardado.Fecha.Year() == 2000 {
		t.Fatalf("la persistencia aceptó campos internos: %+v", guardado)
	}
	var descuentos int64
	e.db.Model(&MovimientoPuntos{}).Where("canje_id = ?", id).Count(&descuentos)
	if descuentos != 0 {
		t.Fatalf("solicitar creó un descuento: %d", descuentos)
	}

	ruta := fmt.Sprintf("/canjes/%d", id)
	codigo, detalle := e.llamar("GET", ruta, RolCliente, int(cliente.ID), "")
	exigirCodigo(t, codigo, http.StatusOK, detalle)
	exigirCanjeDTO(t, detalle)
	codigo, listado := e.llamar("GET", "/canjes", RolAdministrador, 0, "")
	exigirCodigo(t, codigo, http.StatusOK, listado)
	canjes, ok := listado["canjes"].([]any)
	if !ok || len(canjes) != 1 {
		t.Fatalf("lista inesperada: %v", listado)
	}
	exigirCanjeDTO(t, canjes[0].(map[string]any))
	codigo, actualizado := e.llamar("PATCH", ruta+"/estado", RolAdministrador, 0, `{"estado":"confirmado"}`)
	exigirCodigo(t, codigo, http.StatusOK, actualizado)
	exigirCanjeDTO(t, actualizado)
	if actualizado["estado"] != EstadoConfirmado {
		t.Fatalf("PATCH no confirmó el canje: %v", actualizado)
	}
}

func TestDTORespuestasAnidadasConservanContrato(t *testing.T) {
	e := nuevoEntorno(t)
	cliente := e.cliente(t, "0900000001", "anidado@example.com")
	codigo, resultado := e.llamar("POST", "/compras", RolAdministrador, 0,
		fmt.Sprintf(`{"cliente_id":%d,"sucursal_id":%d,"total":75.50,"numero_factura":"DTO-001","id":77}`, cliente.ID, e.centro.ID))
	exigirCodigo(t, codigo, http.StatusCreated, resultado)
	exigirCamposDTO(t, resultado, "compra", "puntos_generados", "fecha_vencimiento")
	compra, ok := resultado["compra"].(map[string]any)
	if !ok {
		t.Fatalf("falta compra anidada: %v", resultado)
	}
	exigirCamposDTO(t, compra, "id", "sucursal_id", "cliente_id", "fecha", "total", "numero_factura")
	if idJSON(t, compra) == 77 {
		t.Fatal("POST /compras aceptó el id enviado por el cliente")
	}
	codigo, puntos := e.llamar("GET", fmt.Sprintf("/clientes/%d/puntos", cliente.ID), RolCliente, int(cliente.ID), "")
	exigirCodigo(t, codigo, http.StatusOK, puntos)
	exigirCamposDTO(t, puntos, "cliente_id", "cedula", "nombre", "saldo", "movimientos")
	movimientos, ok := puntos["movimientos"].([]any)
	if !ok || len(movimientos) != 1 {
		t.Fatalf("movimientos inesperados: %v", puntos)
	}
	exigirCamposDTO(t, movimientos[0].(map[string]any), "id", "cliente_id", "compra_id", "canje_id",
		"tipo", "puntos", "fecha", "fecha_vencimiento")
}

func TestDTOCanjeNoTieneEtiquetasJSON(t *testing.T) {
	tipo := reflect.TypeOf(Canje{})
	for i := 0; i < tipo.NumField(); i++ {
		campo := tipo.Field(i)
		if etiqueta, existe := campo.Tag.Lookup("json"); existe {
			t.Errorf("la entidad Canje.%s todavía tiene json:%q", campo.Name, etiqueta)
		}
	}
}
