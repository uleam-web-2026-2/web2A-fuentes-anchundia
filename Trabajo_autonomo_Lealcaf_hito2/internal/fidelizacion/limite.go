package fidelizacion

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type limiteLogin struct {
	mu       sync.Mutex
	intentos map[string][]time.Time
	reloj    func() time.Time
	limpieza time.Time
}

// LimitarLogin mantiene una ventana móvil por IP de conexión en este proceso.
func LimitarLogin() gin.HandlerFunc {
	return nuevoLimiteLogin(time.Now).middleware
}

func nuevoLimiteLogin(reloj func() time.Time) *limiteLogin {
	return &limiteLogin{intentos: make(map[string][]time.Time), reloj: reloj, limpieza: reloj()}
}

func (l *limiteLogin) contar(ip string) (int, int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.reloj()
	cutoff := now.Add(-time.Minute)
	if !now.Before(l.limpieza.Add(time.Minute)) {
		for key, attempts := range l.intentos {
			if len(attempts) == 0 || !attempts[len(attempts)-1].After(cutoff) {
				delete(l.intentos, key)
			}
		}
		l.limpieza = now
	}
	attempts := l.intentos[ip]
	for len(attempts) > 0 && !attempts[0].After(cutoff) {
		attempts = attempts[1:]
	}
	if len(attempts) >= 5 {
		l.intentos[ip] = attempts
		return 0, int(math.Ceil(attempts[0].Add(time.Minute).Sub(now).Seconds()))
	}
	l.intentos[ip] = append(attempts, now)
	return 5 - len(l.intentos[ip]), 0
}

func (l *limiteLogin) middleware(c *gin.Context) {
	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	ip := net.ParseIP(host)
	if err != nil || ip == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "IP de conexión inválida"})
		return
	}
	remaining, retry := l.contar(ip.String())
	c.Header("X-Ratelimit-Limit", "5")
	c.Header("X-Ratelimit-Remaining", strconv.Itoa(remaining))
	if retry > 0 {
		c.Header("Retry-After", strconv.Itoa(retry))
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "demasiados intentos de inicio de sesión"})
		return
	}
	c.Next()
}
