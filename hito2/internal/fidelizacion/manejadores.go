package fidelizacion

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// API adapta HTTP al servicio; no consulta la base de datos.
type API struct {
	Servicio      *Servicio
	Autenticacion *Autenticacion
}

func SoloRol(rol string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("rol") != rol {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "su rol no tiene permiso para esta acción"})
			return
		}
		c.Next()
	}
}

func responderError(c *gin.Context, err error) {
	codigo := http.StatusInternalServerError
	switch {
	case errors.Is(err, ErrDatosInvalidos), errors.Is(err, ErrEstadoInvalido):
		codigo = http.StatusBadRequest
	case errors.Is(err, ErrProhibido):
		codigo = http.StatusForbidden
	case errors.Is(err, ErrNoEncontrado):
		codigo = http.StatusNotFound
	case errors.Is(err, ErrFacturaDuplicada):
		codigo = http.StatusConflict
	case errors.Is(err, ErrTransicionProhibida), errors.Is(err, ErrPuntosInsuficientes), errors.Is(err, ErrSucursalInactiva), errors.Is(err, ErrClienteInactivo):
		codigo = http.StatusUnprocessableEntity
	}
	if codigo == http.StatusInternalServerError {
		c.JSON(codigo, gin.H{"error": "error interno"})
		return
	}
	cuerpo := gin.H{"error": err.Error()}
	var detalle *ErrorDominio
	if errors.As(err, &detalle) && detalle.Saldo != nil && detalle.Necesarios != nil {
		cuerpo["saldo"] = *detalle.Saldo
		cuerpo["necesarios"] = *detalle.Necesarios
	}
	c.JSON(codigo, cuerpo)
}

func idDeRuta(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, strconv.IntSize)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return 0, false
	}
	return uint(id), true
}

func (a *API) ConsultarPuntos(c *gin.Context) {
	id, ok := idDeRuta(c)
	if !ok {
		return
	}
	resultado, err := a.Servicio.ConsultarPuntos(id, c.GetString("rol"), c.GetUint("cliente_id"))
	if err != nil {
		responderError(c, err)
		return
	}
	c.JSON(http.StatusOK, resultado)
}

func (a *API) ListarRecompensas(c *gin.Context) {
	recompensas, err := a.Servicio.ListarRecompensas()
	if err != nil {
		responderError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"recompensas": recompensas})
}

func (a *API) RegistrarCompra(c *gin.Context) {
	var entrada EntradaCompra
	if err := c.ShouldBindJSON(&entrada); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos inválidos: cliente_id, sucursal_id, total (> 0) y numero_factura son obligatorios"})
		return
	}
	resultado, err := a.Servicio.RegistrarCompra(entrada)
	if err != nil {
		responderError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resultado)
}

func (a *API) SolicitarCanje(c *gin.Context) {
	var entrada EntradaCanje
	if err := c.ShouldBindJSON(&entrada); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos inválidos: recompensa_id y sucursal_id son obligatorios"})
		return
	}
	canje, err := a.Servicio.SolicitarCanje(c.GetUint("cliente_id"), entrada)
	if err != nil {
		responderError(c, err)
		return
	}
	c.JSON(http.StatusCreated, canje)
}

func (a *API) ListarCanjes(c *gin.Context) {
	canjes, err := a.Servicio.ListarCanjes(c.Query("estado"))
	if err != nil {
		responderError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"canjes": canjes})
}

func (a *API) ObtenerCanje(c *gin.Context) {
	id, ok := idDeRuta(c)
	if !ok {
		return
	}
	canje, err := a.Servicio.ObtenerCanje(id, c.GetString("rol"), c.GetUint("cliente_id"))
	if err != nil {
		responderError(c, err)
		return
	}
	c.JSON(http.StatusOK, canje)
}

func (a *API) ActualizarEstadoCanje(c *gin.Context) {
	id, ok := idDeRuta(c)
	if !ok {
		return
	}
	var entrada struct {
		Estado string `json:"estado" binding:"required"`
	}
	if err := c.ShouldBindJSON(&entrada); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos inválidos: estado es obligatorio"})
		return
	}
	canje, err := a.Servicio.ActualizarEstadoCanje(id, entrada.Estado)
	if err != nil {
		responderError(c, err)
		return
	}
	c.JSON(http.StatusOK, canje)
}
