package fidelizacion

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
)

func servidorAvisos(t *testing.T) (*entorno, *Avisos, string) {
	t.Helper()
	e := nuevoEntorno(t)
	a := NuevosAvisos(e.auth)
	r := gin.New()
	r.Use(CORS(), OcultarTokenWS())
	r.GET("/ws", a.Conectar)
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return e, a, "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
}
func abrirWS(t *testing.T, url, token, origin string) (*websocket.Conn, *http.Response, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	headers := http.Header{}
	if origin != "" {
		headers.Set("Origin", origin)
	}
	if token != "" {
		url += "?token=" + token
	}
	return websocket.Dial(ctx, url, &websocket.DialOptions{HTTPHeader: headers})
}
func esperarConexiones(t *testing.T, a *Avisos, n int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		a.mu.Lock()
		count := 0
		for _, list := range a.canales {
			count += len(list)
		}
		a.mu.Unlock()
		if count == n {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("no se alcanzaron %d conexiones", n)
}
func leerAviso(t *testing.T, c *websocket.Conn, want Aviso) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var got Aviso
	if err := wsjson.Read(ctx, c, &got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("aviso: %+v; esperado %+v", got, want)
	}
}
func sinAviso(t *testing.T, c *websocket.Conn) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	var got Aviso
	err := wsjson.Read(ctx, c, &got)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("se esperaba silencio hasta deadline; recibido %+v, error %v", got, err)
	}
}
func TestWebSocketAutenticacionYOrigen(t *testing.T) {
	e, _, url := servidorAvisos(t)
	client := e.cliente(t, "0900000001", "one@example.invalid")
	token := e.token(RolCliente, client.ID)
	for _, tc := range []struct {
		name, token, origin string
		code                int
	}{
		{"sin_token", "", OrigenFront, 401}, {"token_invalido", "incorrecto", OrigenFront, 401},
		{"otro_origen", token, "http://malo.example", 403}, {"esquema_distinto", token, "https://localhost:5173", 403},
		{"permitido", token, OrigenFront, 101}, {"sin_origin_con_token", token, "", 101},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, resp, err := abrirWS(t, url, tc.token, tc.origin)
			if c != nil {
				defer c.CloseNow()
			}
			if resp == nil || resp.StatusCode != tc.code {
				t.Fatalf("esperado %d, respuesta %v error %v", tc.code, resp, err)
			}
			if tc.code == 101 && err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestWebSocketAvisoIndividualVariasConexionesYLimpieza(t *testing.T) {
	e, a, url := servidorAvisos(t)
	owner := e.cliente(t, "0900000001", "one@example.invalid")
	other := e.cliente(t, "0900000002", "two@example.invalid")
	token := e.token(RolCliente, owner.ID)
	first, _, err := abrirWS(t, url, token, OrigenFront)
	if err != nil {
		t.Fatal(err)
	}
	defer first.CloseNow()
	second, _, err := abrirWS(t, url, token, OrigenFront)
	if err != nil {
		t.Fatal(err)
	}
	defer second.CloseNow()
	foreign, _, err := abrirWS(t, url, e.token(RolCliente, other.ID), OrigenFront)
	if err != nil {
		t.Fatal(err)
	}
	defer foreign.CloseNow()
	esperarConexiones(t, a, 3)
	av := Aviso{Tipo: "estado_cambiado", ID: 7, Estado: EstadoConfirmado}
	a.Enviar(owner.ID, av)
	leerAviso(t, first, av)
	leerAviso(t, second, av)
	sinAviso(t, foreign)
	first.CloseNow()
	second.CloseNow()
	foreign.CloseNow()
	esperarConexiones(t, a, 0)
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.canales) != 0 {
		t.Fatal("quedaron identidades sin conexiones")
	}
}
func TestAvisosClienteLentoNoBloqueaYOtrosUsuarios(t *testing.T) {
	a := NuevosAvisos(nil)
	owner := Identidad{CuentaID: 1, ClienteID: 1, Rol: RolCliente}
	other := Identidad{CuentaID: 2, ClienteID: 2, Rol: RolCliente}
	slow := a.suscribir(owner)
	foreign := a.suscribir(other)
	defer a.quitar(owner, slow)
	defer a.quitar(other, foreign)
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			a.Enviar(1, Aviso{ID: uint(i + 1)})
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("el consumidor lento bloqueó el servicio")
	}
	if len(slow) != 8 || len(foreign) != 0 {
		t.Fatal("cola sin límite o fuga a otro usuario")
	}
}
func TestAvisosConcurrenciaSuscribirEnviarQuitar(t *testing.T) {
	a := NuevosAvisos(nil)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := Identidad{CuentaID: 1, ClienteID: 1, Rol: RolCliente}
			for j := 0; j < 50; j++ {
				c := a.suscribir(id)
				a.Enviar(1, Aviso{ID: 1})
				a.quitar(id, c)
			}
		}()
	}
	wg.Wait()
	if len(a.canales) != 0 {
		t.Fatal("fuga de conexiones")
	}
}
func TestWebSocketTokenVencidoNoAbre(t *testing.T) {
	e, _, url := servidorAvisos(t)
	client := e.cliente(t, "0900000001", "one@example.invalid")
	account := e.cuenta("expired", "ficticia", RolCliente, client.ID)
	auth, err := NuevaAutenticacion(e.db, secretoPrueba, time.Nanosecond)
	if err != nil {
		t.Fatal(err)
	}
	token, err := auth.EmitirToken(account)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond)
	c, resp, err := abrirWS(t, url, token, OrigenFront)
	if c != nil {
		c.CloseNow()
	}
	if err == nil || resp == nil || resp.StatusCode != 401 {
		t.Fatal("aceptó token vencido")
	}
}
