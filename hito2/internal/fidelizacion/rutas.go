package fidelizacion

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ConfigurarRutas define token y permisos en un solo lugar.
func ConfigurarRutas(r *gin.Engine, db *gorm.DB, auth *Autenticacion) {
	a := &API{Servicio: &Servicio{Repo: &Repositorio{DB: db}}, Autenticacion: auth}
	r.POST("/login", a.Login)
	g := r.Group("", auth.Middleware())
	g.GET("/clientes/:id/puntos", a.ConsultarPuntos)
	g.GET("/recompensas", a.ListarRecompensas)
	g.POST("/compras", SoloRol(RolAdministrador), a.RegistrarCompra)
	g.POST("/canjes", SoloRol(RolCliente), a.SolicitarCanje)
	g.GET("/canjes", SoloRol(RolAdministrador), a.ListarCanjes)
	g.GET("/canjes/:id", a.ObtenerCanje)
	g.PATCH("/canjes/:id/estado", SoloRol(RolAdministrador), a.ActualizarEstadoCanje)
}
