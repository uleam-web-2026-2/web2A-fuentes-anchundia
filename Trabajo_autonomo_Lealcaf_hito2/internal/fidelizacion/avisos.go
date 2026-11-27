package fidelizacion

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
)

// Avisos conserva un canal por conexión: varias pestañas no se reemplazan.
type Avisos struct {
	auth    *Autenticacion
	mu      sync.Mutex
	canales map[string]map[chan Aviso]Identidad
}

func NuevosAvisos(auth *Autenticacion) *Avisos {
	return &Avisos{auth: auth, canales: make(map[string]map[chan Aviso]Identidad)}
}

func claveCliente(id uint) string { return "cliente:" + strconv.FormatUint(uint64(id), 10) }
func claveConexion(id Identidad) string {
	if id.Rol == RolCliente {
		return claveCliente(id.ClienteID)
	}
	return "cuenta:" + strconv.FormatUint(uint64(id.CuentaID), 10)
}

func (a *Avisos) suscribir(id Identidad) chan Aviso {
	canal := make(chan Aviso, 8)
	a.mu.Lock()
	defer a.mu.Unlock()
	key := claveConexion(id)
	if a.canales[key] == nil {
		a.canales[key] = make(map[chan Aviso]Identidad)
	}
	a.canales[key][canal] = id
	return canal
}

func (a *Avisos) quitar(id Identidad, canal chan Aviso) {
	a.mu.Lock()
	defer a.mu.Unlock()
	key := claveConexion(id)
	delete(a.canales[key], canal)
	if len(a.canales[key]) == 0 {
		delete(a.canales, key)
	}
	// No cerrar el canal: el receptor sale con el contexto de la conexión.
}

func (a *Avisos) Enviar(clienteID uint, av Aviso) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for canal := range a.canales[claveCliente(clienteID)] {
		select {
		case canal <- av:
		default: /* Cliente lento: se descarta este aviso. */
		}
	}
}

// NuevoCanje aplica la selección del rol administrativo en el transporte.
// El cliente nunca recibe avisos que después tenga que filtrar.
func (a *Avisos) NuevoCanje(av Aviso) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, conexiones := range a.canales {
		for canal, identidad := range conexiones {
			if identidad.Rol != RolAdministrador {
				continue
			}
			select {
			case canal <- av:
			default:
			}
		}
	}
}

// Conectar valida la sesión antes de Accept, fuera del guardia Bearer.
func (a *Avisos) Conectar(c *gin.Context) {
	token := c.GetString("token_ws")
	if token == "" {
		token = c.Query("token")
	}
	id, err := a.auth.ValidarToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "sin sesión: token ausente, inválido o vencido"})
		return
	}
	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{OriginPatterns: []string{OrigenFront}})
	if err != nil {
		return
	}
	defer conn.CloseNow()
	canal := a.suscribir(id)
	defer a.quitar(id, canal)
	ctx := conn.CloseRead(context.Background())
	tick := time.NewTicker(30 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case av := <-canal:
			writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := wsjson.Write(writeCtx, conn, av)
			cancel()
			if err != nil {
				return
			}
		case <-tick.C:
			pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := conn.Ping(pingCtx)
			cancel()
			if err != nil {
				return
			}
		}
	}
}

// OcultarTokenWS evita que Gin escriba una credencial en el log desde el origen.
// El JWT original solo vive en el contexto de esta petición.
func OcultarTokenWS() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/ws" {
			q := c.Request.URL.Query()
			if token := q.Get("token"); token != "" {
				c.Set("token_ws", token)
				q.Set("token", "BORRADO")
				c.Request.URL.RawQuery = q.Encode()
			}
		}
		c.Next()
	}
}
