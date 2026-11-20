package fidelizacion

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestLimiteLoginVentanaMovilConRelojControlado(t *testing.T) {
	now := time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC)
	l := nuevoLimiteLogin(func() time.Time { return now })
	for i := 0; i < 5; i++ {
		remaining, retry := l.contar("192.0.2.1")
		if remaining != 4-i || retry != 0 {
			t.Fatalf("intento %d: restantes=%d espera=%d", i+1, remaining, retry)
		}
		now = now.Add(time.Second)
	}
	if _, retry := l.contar("192.0.2.1"); retry != 55 {
		t.Fatalf("el sexto debe esperar 55 segundos: %d", retry)
	}
	if remaining, retry := l.contar("192.0.2.2"); remaining != 4 || retry != 0 {
		t.Fatal("la segunda IP debe tener su propia ventana")
	}
	now = now.Add(55 * time.Second)
	if remaining, retry := l.contar("192.0.2.1"); remaining != 0 || retry != 0 {
		t.Fatal("a los 60 segundos debe liberarse solo el intento más antiguo")
	}
	now = now.Add(2 * time.Minute)
	l.contar("192.0.2.3")
	if len(l.intentos) != 1 {
		t.Fatal("las IP inactivas deben retirarse del contador")
	}
}

func TestLimiteLoginIgnoraCabecerasYRespetaOtrasRutas(t *testing.T) {
	r := gin.New()
	r.POST("/login", LimitarLogin(), func(c *gin.Context) { c.Status(http.StatusUnauthorized) })
	r.GET("/canjes", func(c *gin.Context) { c.Status(http.StatusOK) })
	for i := 0; i < 7; i++ {
		req := httptest.NewRequest("POST", "/login", nil)
		req.RemoteAddr = "[::ffff:192.0.2.1]:12345"
		req.Header.Set("X-Forwarded-For", "198.51.100."+string(rune('1'+i)))
		req.Header.Set("X-Real-IP", "198.51.100.99")
		response := httptest.NewRecorder()
		r.ServeHTTP(response, req)
		want := http.StatusUnauthorized
		if i >= 5 {
			want = http.StatusTooManyRequests
			if response.Header().Get("Retry-After") == "" {
				t.Fatal("falta Retry-After")
			}
		}
		if response.Code != want {
			t.Fatalf("intento %d: HTTP %d, esperado %d", i+1, response.Code, want)
		}
	}
	response := httptest.NewRecorder()
	r.ServeHTTP(response, httptest.NewRequest("GET", "/canjes", nil))
	if response.Code != http.StatusOK {
		t.Fatal("el límite debe afectar solo al login")
	}
}

func TestLimiteLoginConcurrente(t *testing.T) {
	l := nuevoLimiteLogin(time.Now)
	var passed atomic.Int32
	var group sync.WaitGroup
	for i := 0; i < 30; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if _, retry := l.contar("192.0.2.1"); retry == 0 {
				passed.Add(1)
			}
		}()
	}
	group.Wait()
	if passed.Load() != 5 {
		t.Fatalf("se aceptaron %d solicitudes concurrentes", passed.Load())
	}
}
