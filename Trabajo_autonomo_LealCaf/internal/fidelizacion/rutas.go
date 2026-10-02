package fidelizacion

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ConfigurarRutas registra los siete endpoints del Hito 1.
// Rol: encabezado X-Rol (cliente | administrador); el cliente además envía X-Cliente-ID.
func ConfigurarRutas(r *gin.Engine, db *gorm.DB) {
	a := &API{DB: db}
	g := r.Group("", a.Autenticar())

	g.GET("/clientes/:id/puntos", a.ConsultarPuntos)
	g.GET("/recompensas", a.ListarRecompensas)

	g.POST("/compras", SoloRol(RolAdministrador), a.RegistrarCompra)

	g.POST("/canjes", SoloRol(RolCliente), a.SolicitarCanje)
	g.GET("/canjes", SoloRol(RolAdministrador), a.ListarCanjes)
	g.GET("/canjes/:id", a.ObtenerCanje)
	g.PATCH("/canjes/:id/estado", SoloRol(RolAdministrador), a.ActualizarEstadoCanje)
}
