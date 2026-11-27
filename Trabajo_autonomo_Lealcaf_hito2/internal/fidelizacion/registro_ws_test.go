package fidelizacion

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestWebSocketRegistroOcultaTokenAntesDelLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var log bytes.Buffer
	r := gin.New()
	r.Use(OcultarTokenWS(), gin.LoggerWithWriter(&log))
	r.GET("/ws", func(c *gin.Context) {
		if c.GetString("token_ws") != "credencial-ficticia" || c.Query("token") != "BORRADO" {
			t.Fatal("no preservó el token solo en memoria")
		}
		c.Status(http.StatusUnauthorized)
	})
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/ws?token=credencial-ficticia", nil))
	if strings.Contains(log.String(), "credencial-ficticia") || !strings.Contains(log.String(), "token=BORRADO") {
		t.Fatalf("registro incorrecto: %s", log.String())
	}
}

func TestSegundoClienteDemoTieneIdentidadDiferente(t *testing.T) {
	e := nuevoEntorno(t)
	configurarDemoPrueba(t)
	if err := PrepararBase(e.db); err != nil {
		t.Fatal(err)
	}
	first, err := e.auth.IniciarSesion("cliente", "cliente-2026")
	if err != nil {
		t.Fatal(err)
	}
	second, err := e.auth.IniciarSesion("cliente2", "cliente-2026")
	if err != nil {
		t.Fatal(err)
	}
	if first.ClienteID == nil || second.ClienteID == nil || *first.ClienteID == *second.ClienteID || second.Rol != RolCliente {
		t.Fatal("el segundo cliente no tiene identidad propia")
	}
}
