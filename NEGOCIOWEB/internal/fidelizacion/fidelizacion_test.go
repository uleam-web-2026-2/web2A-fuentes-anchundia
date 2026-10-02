package fidelizacion

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type entorno struct {
	r  *gin.Engine
	db *gorm.DB
}

// nuevoEntorno crea una base SQLite en memoria (una por prueba) con las mismas tablas.
func nuevoEntorno(t *testing.T) *entorno {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := Migrar(db); err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	ConfigurarRutas(r, db)
	return &entorno{r: r, db: db}
}

func (e *entorno) llamar(metodo, ruta, rol string, clienteID int, cuerpo string) (int, map[string]any) {
	req := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Rol", rol)
	if clienteID > 0 {
		req.Header.Set("X-Cliente-ID", strconv.Itoa(clienteID))
	}
	w := httptest.NewRecorder()
	e.r.ServeHTTP(w, req)
	var salida map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &salida)
	return w.Code, salida
}

func (e *entorno) cliente(t *testing.T, cedula, correo string) Cliente {
	t.Helper()
	c := Cliente{Cedula: cedula, Nombre: "Cliente " + cedula, Correo: correo}
	if err := e.db.Create(&c).Error; err != nil {
		t.Fatal(err)
	}
	return c
}

func (e *entorno) acumular(t *testing.T, clienteID uint, puntos int, vence time.Time) {
	t.Helper()
	m := MovimientoPuntos{ClienteID: clienteID, Tipo: TipoAcumulacion, Puntos: puntos, Fecha: time.Now(), FechaVencimiento: &vence}
	if err := e.db.Create(&m).Error; err != nil {
		t.Fatal(err)
	}
}

func (e *entorno) recompensa(t *testing.T, nombre string, puntos int) Recompensa {
	t.Helper()
	r := Recompensa{Nombre: nombre, PuntosNecesarios: puntos, Activa: true}
	if err := e.db.Create(&r).Error; err != nil {
		t.Fatal(err)
	}
	return r
}

func TestConsultarPuntos(t *testing.T) {
	e := nuevoEntorno(t)
	c := e.cliente(t, "0900000001", "a@example.com")
	e.acumular(t, c.ID, 40, time.Now().AddDate(0, 6, 0))
	e.db.Create(&MovimientoPuntos{ClienteID: c.ID, Tipo: TipoCanje, Puntos: -10, Fecha: time.Now()})

	codigo, cuerpo := e.llamar("GET", "/clientes/1/puntos", RolCliente, 1, "")
	if codigo != http.StatusOK {
		t.Fatalf("esperaba 200, obtuvo %d", codigo)
	}
	if cuerpo["saldo"].(float64) != 30 {
		t.Fatalf("esperaba saldo 30, obtuvo %v", cuerpo["saldo"])
	}
}

func TestClienteNoPuedeVerPuntosAjenos(t *testing.T) {
	e := nuevoEntorno(t)
	e.cliente(t, "0900000001", "a@example.com")
	e.cliente(t, "0900000002", "b@example.com")

	codigo, _ := e.llamar("GET", "/clientes/1/puntos", RolCliente, 2, "")
	if codigo != http.StatusForbidden {
		t.Fatalf("esperaba 403, obtuvo %d", codigo)
	}
}

func TestPuntosVencidosNoCuentanEnElSaldo(t *testing.T) {
	e := nuevoEntorno(t)
	c := e.cliente(t, "0900000001", "a@example.com")
	e.acumular(t, c.ID, 50, time.Now().AddDate(0, 0, -1)) // venció ayer

	_, cuerpo := e.llamar("GET", "/clientes/1/puntos", RolAdministrador, 0, "")
	if cuerpo["saldo"].(float64) != 0 {
		t.Fatalf("esperaba saldo 0, obtuvo %v", cuerpo["saldo"])
	}
}

func TestRegistrarCompraGeneraPuntosConVencimiento(t *testing.T) {
	e := nuevoEntorno(t)
	e.cliente(t, "0900000001", "a@example.com")

	codigo, cuerpo := e.llamar("POST", "/compras", RolAdministrador, 0,
		`{"cliente_id":1,"total":50.75,"numero_factura":"001-001-000000123"}`)
	if codigo != http.StatusCreated {
		t.Fatalf("esperaba 201, obtuvo %d", codigo)
	}
	if cuerpo["puntos_generados"].(float64) != 50 {
		t.Fatalf("esperaba 50 puntos, obtuvo %v", cuerpo["puntos_generados"])
	}
	var m MovimientoPuntos
	if err := e.db.First(&m).Error; err != nil {
		t.Fatal(err)
	}
	if m.Tipo != TipoAcumulacion || m.FechaVencimiento == nil || !m.FechaVencimiento.After(m.Fecha.AddDate(0, 5, 0)) {
		t.Fatalf("movimiento mal registrado: %+v", m)
	}
}

func TestRegistrarCompraConFacturaDuplicada(t *testing.T) {
	e := nuevoEntorno(t)
	e.cliente(t, "0900000001", "a@example.com")
	cuerpo := `{"cliente_id":1,"total":10,"numero_factura":"001-001-000000999"}`

	e.llamar("POST", "/compras", RolAdministrador, 0, cuerpo)
	codigo, _ := e.llamar("POST", "/compras", RolAdministrador, 0, cuerpo)
	if codigo != http.StatusConflict {
		t.Fatalf("esperaba 409, obtuvo %d", codigo)
	}
}

func TestClienteNoPuedeRegistrarCompras(t *testing.T) {
	e := nuevoEntorno(t)
	e.cliente(t, "0900000001", "a@example.com")

	codigo, _ := e.llamar("POST", "/compras", RolCliente, 1,
		`{"cliente_id":1,"total":10,"numero_factura":"001-001-000000555"}`)
	if codigo != http.StatusForbidden {
		t.Fatalf("esperaba 403, obtuvo %d", codigo)
	}
}

func TestCanjeConPuntosInsuficientes(t *testing.T) {
	e := nuevoEntorno(t)
	c := e.cliente(t, "0900000001", "a@example.com")
	e.acumular(t, c.ID, 20, time.Now().AddDate(0, 6, 0))
	r := e.recompensa(t, "Café gratis", 40)

	codigo, cuerpo := e.llamar("POST", "/canjes", RolCliente, int(c.ID), fmt.Sprintf(`{"recompensa_id":%d}`, r.ID))
	if codigo != http.StatusBadRequest {
		t.Fatalf("esperaba 400, obtuvo %d", codigo)
	}
	if cuerpo["error"] != "puntos insuficientes" {
		t.Fatalf("mensaje inesperado: %v", cuerpo["error"])
	}
}

func TestCanjeCorrectoDescuentaPuntosAlConfirmar(t *testing.T) {
	e := nuevoEntorno(t)
	c := e.cliente(t, "0900000001", "a@example.com")
	e.acumular(t, c.ID, 100, time.Now().AddDate(0, 6, 0))
	r := e.recompensa(t, "Café gratis", 40)

	codigo, canje := e.llamar("POST", "/canjes", RolCliente, int(c.ID), fmt.Sprintf(`{"recompensa_id":%d}`, r.ID))
	if codigo != http.StatusCreated || canje["estado"] != EstadoSolicitado {
		t.Fatalf("esperaba 201 solicitado, obtuvo %d %v", codigo, canje["estado"])
	}
	// Mientras está solicitado no se descuentan puntos.
	_, p := e.llamar("GET", "/clientes/1/puntos", RolCliente, int(c.ID), "")
	if p["saldo"].(float64) != 100 {
		t.Fatalf("el saldo no debe cambiar al solicitar, obtuvo %v", p["saldo"])
	}

	ruta := fmt.Sprintf("/canjes/%d/estado", int(canje["id"].(float64)))
	codigo, conf := e.llamar("PATCH", ruta, RolAdministrador, 0, `{"estado":"confirmado"}`)
	if codigo != http.StatusOK || conf["estado"] != EstadoConfirmado {
		t.Fatalf("esperaba 200 confirmado, obtuvo %d %v", codigo, conf["estado"])
	}
	_, p = e.llamar("GET", "/clientes/1/puntos", RolCliente, int(c.ID), "")
	if p["saldo"].(float64) != 60 {
		t.Fatalf("esperaba saldo 60, obtuvo %v", p["saldo"])
	}
}

func TestTransicionProhibidaDeConfirmadoACancelado(t *testing.T) {
	e := nuevoEntorno(t)
	c := e.cliente(t, "0900000001", "a@example.com")
	e.acumular(t, c.ID, 100, time.Now().AddDate(0, 6, 0))
	r := e.recompensa(t, "Café gratis", 40)

	_, canje := e.llamar("POST", "/canjes", RolCliente, int(c.ID), fmt.Sprintf(`{"recompensa_id":%d}`, r.ID))
	ruta := fmt.Sprintf("/canjes/%d/estado", int(canje["id"].(float64)))
	e.llamar("PATCH", ruta, RolAdministrador, 0, `{"estado":"confirmado"}`)

	codigo, _ := e.llamar("PATCH", ruta, RolAdministrador, 0, `{"estado":"cancelado"}`)
	if codigo != http.StatusConflict {
		t.Fatalf("esperaba 409, obtuvo %d", codigo)
	}
}

func TestRecompensasSoloMuestraLasActivas(t *testing.T) {
	e := nuevoEntorno(t)
	e.cliente(t, "0900000001", "a@example.com")
	e.recompensa(t, "Activa", 40)
	inactiva := e.recompensa(t, "Inactiva", 60)
	e.db.Model(&inactiva).Update("activa", false)

	codigo, cuerpo := e.llamar("GET", "/recompensas", RolCliente, 1, "")
	if codigo != http.StatusOK {
		t.Fatalf("esperaba 200, obtuvo %d", codigo)
	}
	if n := len(cuerpo["recompensas"].([]any)); n != 1 {
		t.Fatalf("esperaba 1 recompensa activa, obtuvo %d", n)
	}
}

func TestVerCanjePropioYNoAjeno(t *testing.T) {
	e := nuevoEntorno(t)
	a := e.cliente(t, "0900000001", "a@example.com")
	b := e.cliente(t, "0900000002", "b@example.com")
	e.acumular(t, a.ID, 100, time.Now().AddDate(0, 6, 0))
	r := e.recompensa(t, "Café gratis", 40)

	_, canje := e.llamar("POST", "/canjes", RolCliente, int(a.ID), fmt.Sprintf(`{"recompensa_id":%d}`, r.ID))
	ruta := fmt.Sprintf("/canjes/%d", int(canje["id"].(float64)))

	codigo, cuerpo := e.llamar("GET", ruta, RolCliente, int(a.ID), "")
	if codigo != http.StatusOK || cuerpo["estado"] != EstadoSolicitado {
		t.Fatalf("esperaba 200 solicitado, obtuvo %d %v", codigo, cuerpo["estado"])
	}
	codigo, _ = e.llamar("GET", ruta, RolCliente, int(b.ID), "")
	if codigo != http.StatusForbidden {
		t.Fatalf("esperaba 403 para un canje ajeno, obtuvo %d", codigo)
	}
	codigo, _ = e.llamar("GET", "/canjes/999", RolAdministrador, 0, "")
	if codigo != http.StatusNotFound {
		t.Fatalf("esperaba 404 para un canje inexistente, obtuvo %d", codigo)
	}
}
