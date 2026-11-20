package fidelizacion

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORSPreflightAntesDeAutenticacion(t *testing.T) {
	r := gin.New()
	r.Use(CORS())
	r.Use(func(c *gin.Context) { c.AbortWithStatus(http.StatusUnauthorized) })
	for _, origin := range []string{"http://localhost:5173", "http://malo.example"} {
		req := httptest.NewRequest("OPTIONS", "/canjes/1", nil)
		req.Header.Set("Origin", origin)
		req.Header.Set("Access-Control-Request-Method", "PATCH")
		req.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)
		if response.Code != http.StatusOK {
			t.Fatalf("preflight %s: HTTP %d", origin, response.Code)
		}
		allowed := response.Header().Get("Access-Control-Allow-Origin")
		if (origin == "http://localhost:5173" && allowed != origin) || (origin != "http://localhost:5173" && allowed != "") {
			t.Fatalf("permiso de origen incorrecto: %q", allowed)
		}
	}
}

func TestCORSNoSustituyeLosPermisosDelServidor(t *testing.T) {
	r := gin.New()
	r.Use(CORS())
	r.PATCH("/canjes/1/estado", func(c *gin.Context) { c.Status(http.StatusOK) })
	req := httptest.NewRequest("PATCH", "/canjes/1/estado", nil)
	req.Header.Set("Origin", "http://malo.example")
	response := httptest.NewRecorder()
	r.ServeHTTP(response, req)
	if response.Code != http.StatusOK || response.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("el servidor debe atender sin otorgar permiso CORS al navegador")
	}
}
