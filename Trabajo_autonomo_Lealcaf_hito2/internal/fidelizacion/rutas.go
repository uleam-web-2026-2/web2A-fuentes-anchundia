package fidelizacion

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ConfigurarRutas define token y permisos en un solo lugar.
func ConfigurarRutas(r *gin.Engine, db *gorm.DB, auth *Autenticacion) {
	ConfigurarRutasConRepositorio(r, &RepositorioGORM{DB: db}, auth)
}

// ConfigurarRutasConRepositorio conecta las rutas con el repositorio del servicio.
func ConfigurarRutasConRepositorio(r *gin.Engine, repo Repositorio, auth *Autenticacion) {
	a := &API{Servicio: &Servicio{Repo: repo}, Autenticacion: auth}
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
