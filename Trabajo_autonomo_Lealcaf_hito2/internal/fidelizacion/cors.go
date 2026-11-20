package fidelizacion

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/cors"
)

// CORS adapta el middleware net/http del kit a la cadena de Gin.
func CORS() gin.HandlerFunc {
	middleware := cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173"},
		AllowedMethods: []string{"GET", "POST", "PATCH"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		MaxAge:         300,
	})
	return func(c *gin.Context) {
		middleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			c.Next()
		})).ServeHTTP(c.Writer, c.Request)
		// Si CORS atendió OPTIONS, Gin no debe ejecutar más manejadores.
		c.Abort()
	}
}
