package fidelizacion

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ahora se puede reemplazar en pruebas.
var ahora = time.Now

var errPuntosInsuficientes = errors.New("puntos insuficientes")

type API struct {
	DB *gorm.DB
}

// Autenticar es una simplificación del Hito 1: el rol llega en el encabezado
// X-Rol y, para el cliente, su identificador en X-Cliente-ID. Sin login todavía.
func (a *API) Autenticar() gin.HandlerFunc {
	return func(c *gin.Context) {
		rol := c.GetHeader("X-Rol")
		if rol != RolCliente && rol != RolAdministrador {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "rol no identificado: envíe el encabezado X-Rol"})
			return
		}
		c.Set("rol", rol)
		if rol == RolCliente {
			id, err := strconv.ParseUint(c.GetHeader("X-Cliente-ID"), 10, 64)
			if err != nil || id == 0 {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "cliente no identificado: envíe el encabezado X-Cliente-ID"})
				return
			}
			c.Set("cliente_id", uint(id))
		}
		c.Next()
	}
}

// SoloRol deja pasar únicamente al rol indicado.
func SoloRol(rol string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("rol") != rol {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "su rol no tiene permiso para esta acción"})
			return
		}
		c.Next()
	}
}

// saldoDe suma las acumulaciones vigentes (no vencidas) y los canjes (negativos).
func saldoDe(db *gorm.DB, clienteID uint) (int, error) {
	var suma int64
	err := db.Model(&MovimientoPuntos{}).
		Select("COALESCE(SUM(puntos), 0)").
		Where("cliente_id = ?", clienteID).
		Where("(tipo = ? AND fecha_vencimiento > ?) OR tipo = ?", TipoAcumulacion, ahora(), TipoCanje).
		Scan(&suma).Error
	if err != nil {
		return 0, err
	}
	if suma < 0 {
		suma = 0
	}
	return int(suma), nil
}

func errorBD(c *gin.Context, err error, noEncontrado string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": noEncontrado})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
}

// GET /clientes/{id}/puntos — cliente (solo el suyo) y administrador.
func (a *API) ConsultarPuntos(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	id := uint(id64)
	if c.GetString("rol") == RolCliente && c.GetUint("cliente_id") != id {
		c.JSON(http.StatusForbidden, gin.H{"error": "solo puede consultar sus propios puntos"})
		return
	}

	var cliente Cliente
	if err := a.DB.First(&cliente, id).Error; err != nil {
		errorBD(c, err, "cliente no encontrado")
		return
	}
	saldo, err := saldoDe(a.DB, id)
	if err != nil {
		errorBD(c, err, "")
		return
	}
	var movimientos []MovimientoPuntos
	if err := a.DB.Where("cliente_id = ?", id).Order("fecha DESC, id DESC").Find(&movimientos).Error; err != nil {
		errorBD(c, err, "")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"cliente_id":  cliente.ID,
		"cedula":      cliente.Cedula,
		"nombre":      cliente.Nombre,
		"saldo":       saldo,
		"movimientos": movimientos,
	})
}

// GET /recompensas — ambos roles; solo las activas.
func (a *API) ListarRecompensas(c *gin.Context) {
	var recompensas []Recompensa
	if err := a.DB.Where("activa = ?", true).Order("puntos_necesarios").Find(&recompensas).Error; err != nil {
		errorBD(c, err, "")
		return
	}
	c.JSON(http.StatusOK, gin.H{"recompensas": recompensas})
}

// POST /compras — administrador. Registra la compra y genera los puntos.
func (a *API) RegistrarCompra(c *gin.Context) {
	var in struct {
		ClienteID     uint    `json:"cliente_id" binding:"required"`
		Total         float64 `json:"total" binding:"required,gt=0"`
		NumeroFactura string  `json:"numero_factura" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos inválidos: cliente_id, total (> 0) y numero_factura son obligatorios"})
		return
	}

	var cliente Cliente
	if err := a.DB.First(&cliente, in.ClienteID).Error; err != nil {
		errorBD(c, err, "cliente no encontrado")
		return
	}
	if !cliente.Activo {
		c.JSON(http.StatusBadRequest, gin.H{"error": "el cliente está inactivo"})
		return
	}

	var existe int64
	a.DB.Model(&Compra{}).Where("numero_factura = ?", in.NumeroFactura).Count(&existe)
	if existe > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "esa factura ya fue registrada"})
		return
	}

	fecha := ahora()
	vence := fecha.AddDate(0, MesesVigencia, 0)
	puntos := int(in.Total) * PuntosPorDolar
	compra := Compra{ClienteID: in.ClienteID, Fecha: fecha, Total: in.Total, NumeroFactura: in.NumeroFactura}

	err := a.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&compra).Error; err != nil {
			return err
		}
		if puntos == 0 {
			return nil
		}
		m := MovimientoPuntos{
			ClienteID: in.ClienteID, CompraID: &compra.ID, Tipo: TipoAcumulacion,
			Puntos: puntos, Fecha: fecha, FechaVencimiento: &vence,
		}
		return tx.Create(&m).Error
	})
	if err != nil {
		errorBD(c, err, "")
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"compra":            compra,
		"puntos_generados":  puntos,
		"fecha_vencimiento": vence,
	})
}

// POST /canjes — cliente. Crea el canje en estado "solicitado".
func (a *API) SolicitarCanje(c *gin.Context) {
	var in struct {
		RecompensaID uint `json:"recompensa_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos inválidos: recompensa_id es obligatorio"})
		return
	}
	clienteID := c.GetUint("cliente_id")

	var cliente Cliente
	if err := a.DB.First(&cliente, clienteID).Error; err != nil {
		errorBD(c, err, "cliente no encontrado")
		return
	}
	var rec Recompensa
	if err := a.DB.Where("activa = ?", true).First(&rec, in.RecompensaID).Error; err != nil {
		errorBD(c, err, "recompensa no disponible")
		return
	}
	saldo, err := saldoDe(a.DB, clienteID)
	if err != nil {
		errorBD(c, err, "")
		return
	}
	if saldo < rec.PuntosNecesarios {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "puntos insuficientes", "saldo": saldo, "necesarios": rec.PuntosNecesarios,
		})
		return
	}
	canje := Canje{
		ClienteID: clienteID, RecompensaID: rec.ID,
		PuntosUtilizados: rec.PuntosNecesarios, Fecha: ahora(), Estado: EstadoSolicitado,
	}
	if err := a.DB.Create(&canje).Error; err != nil {
		errorBD(c, err, "")
		return
	}
	c.JSON(http.StatusCreated, canje)
}

// GET /canjes?estado=solicitado — administrador.
func (a *API) ListarCanjes(c *gin.Context) {
	q := a.DB.Model(&Canje{}).Order("fecha DESC, id DESC")
	if estado := c.Query("estado"); estado != "" {
		if estado != EstadoSolicitado && estado != EstadoConfirmado && estado != EstadoCancelado {
			c.JSON(http.StatusBadRequest, gin.H{"error": "estado no válido"})
			return
		}
		q = q.Where("estado = ?", estado)
	}
	var canjes []Canje
	if err := q.Find(&canjes).Error; err != nil {
		errorBD(c, err, "")
		return
	}
	c.JSON(http.StatusOK, gin.H{"canjes": canjes})
}

// GET /canjes/{id} — cliente (solo los suyos) y administrador (cualquiera).
func (a *API) ObtenerCanje(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	var canje Canje
	if err := a.DB.First(&canje, uint(id64)).Error; err != nil {
		errorBD(c, err, "canje no encontrado")
		return
	}
	if c.GetString("rol") == RolCliente && c.GetUint("cliente_id") != canje.ClienteID {
		c.JSON(http.StatusForbidden, gin.H{"error": "solo puede consultar sus propios canjes"})
		return
	}
	c.JSON(http.StatusOK, canje)
}

// PATCH /canjes/{id}/estado — administrador. Máquina de estados del canje:
// solicitado -> confirmado | cancelado. Cualquier otra transición se rechaza.
func (a *API) ActualizarEstadoCanje(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	var in struct {
		Estado string `json:"estado" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos inválidos: estado es obligatorio"})
		return
	}
	if in.Estado != EstadoConfirmado && in.Estado != EstadoCancelado {
		c.JSON(http.StatusBadRequest, gin.H{"error": "estado no válido: use confirmado o cancelado"})
		return
	}

	var canje Canje
	if err := a.DB.First(&canje, uint(id64)).Error; err != nil {
		errorBD(c, err, "canje no encontrado")
		return
	}
	if canje.Estado != EstadoSolicitado {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("transición no permitida: de %s a %s", canje.Estado, in.Estado)})
		return
	}

	err = a.DB.Transaction(func(tx *gorm.DB) error {
		if in.Estado == EstadoConfirmado {
			saldo, err := saldoDe(tx, canje.ClienteID)
			if err != nil {
				return err
			}
			if saldo < canje.PuntosUtilizados {
				return errPuntosInsuficientes
			}
			m := MovimientoPuntos{
				ClienteID: canje.ClienteID, CanjeID: &canje.ID, Tipo: TipoCanje,
				Puntos: -canje.PuntosUtilizados, Fecha: ahora(),
			}
			if err := tx.Create(&m).Error; err != nil {
				return err
			}
		}
		return tx.Model(&canje).Update("estado", in.Estado).Error
	})
	if errors.Is(err, errPuntosInsuficientes) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "puntos insuficientes"})
		return
	}
	if err != nil {
		errorBD(c, err, "")
		return
	}
	c.JSON(http.StatusOK, canje)
}
