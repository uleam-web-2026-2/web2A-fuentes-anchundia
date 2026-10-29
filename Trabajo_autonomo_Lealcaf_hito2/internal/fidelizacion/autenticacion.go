package fidelizacion

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	ErrCredencialesInvalidas = errors.New("usuario o contraseña incorrectos")
	ErrTokenInvalido         = errors.New("token inválido o vencido")
	claveFicticia, _         = bcrypt.GenerateFromPassword([]byte("cuenta-inexistente-no-iniciar"), bcrypt.DefaultCost)
)

type Identidad struct {
	CuentaID  uint
	Rol       string
	ClienteID uint
}

type claimsAcceso struct {
	Rol       string `json:"rol"`
	ClienteID uint   `json:"cliente_id"`
	jwt.RegisteredClaims
}

type Autenticacion struct {
	db       *gorm.DB
	secreto  []byte
	duracion time.Duration
}

func NuevaAutenticacion(db *gorm.DB, secreto string, duracion time.Duration) (*Autenticacion, error) {
	if db == nil {
		return nil, errors.New("la autenticación necesita una base de datos")
	}
	if len([]byte(secreto)) < 32 || strings.TrimSpace(secreto) == "" {
		return nil, errors.New("JWT_SECRET debe contener al menos 32 bytes; genere una clave local y no la suba al repositorio")
	}
	if duracion <= 0 {
		return nil, errors.New("JWT_TTL debe ser una duración positiva")
	}
	return &Autenticacion{db: db, secreto: []byte(secreto), duracion: duracion}, nil
}

func (a *Autenticacion) IniciarSesion(usuario, password string) (Cuenta, error) {
	var cuenta Cuenta
	err := a.db.Where("usuario = ?", strings.TrimSpace(usuario)).First(&cuenta).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		_ = bcrypt.CompareHashAndPassword(claveFicticia, []byte(password))
		return Cuenta{}, ErrCredencialesInvalidas
	}
	if err != nil {
		return Cuenta{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(cuenta.PasswordHash), []byte(password)) != nil {
		return Cuenta{}, ErrCredencialesInvalidas
	}
	return cuenta, nil
}

func (a *Autenticacion) EmitirToken(cuenta Cuenta) (string, error) {
	if cuenta.ID == 0 || (cuenta.Rol != RolCliente && cuenta.Rol != RolAdministrador) ||
		(cuenta.Rol == RolCliente && (cuenta.ClienteID == nil || *cuenta.ClienteID == 0)) {
		return "", errors.New("la cuenta no tiene una identidad válida")
	}
	var clienteID uint
	if cuenta.ClienteID != nil {
		clienteID = *cuenta.ClienteID
	}
	inicio := time.Now()
	claims := claimsAcceso{Rol: cuenta.Rol, ClienteID: clienteID, RegisteredClaims: jwt.RegisteredClaims{
		Issuer: "lealcaf", Subject: strconv.FormatUint(uint64(cuenta.ID), 10),
		Audience: jwt.ClaimStrings{"lealcaf-api"}, IssuedAt: jwt.NewNumericDate(inicio),
		ExpiresAt: jwt.NewNumericDate(inicio.Add(a.duracion)),
	}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(a.secreto)
}

func (a *Autenticacion) ValidarToken(texto string) (Identidad, error) {
	if len(texto) == 0 || len(texto) > 8192 {
		return Identidad{}, ErrTokenInvalido
	}
	claims := &claimsAcceso{}
	token, err := jwt.ParseWithClaims(texto, claims, func(t *jwt.Token) (any, error) {
		return a.secreto, nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired(),
		jwt.WithIssuer("lealcaf"), jwt.WithAudience("lealcaf-api"), jwt.WithIssuedAt())
	if err != nil || token == nil || !token.Valid || claims.IssuedAt == nil || claims.ExpiresAt == nil ||
		!claims.ExpiresAt.After(claims.IssuedAt.Time) {
		return Identidad{}, ErrTokenInvalido
	}
	id64, err := strconv.ParseUint(claims.Subject, 10, strconv.IntSize)
	if err != nil || id64 == 0 {
		return Identidad{}, ErrTokenInvalido
	}
	var cuenta Cuenta
	if err := a.db.First(&cuenta, uint(id64)).Error; err != nil {
		return Identidad{}, ErrTokenInvalido
	}
	var clienteID uint
	if cuenta.ClienteID != nil {
		clienteID = *cuenta.ClienteID
	}
	if (cuenta.Rol != RolCliente && cuenta.Rol != RolAdministrador) || claims.Rol != cuenta.Rol ||
		claims.ClienteID != clienteID || (cuenta.Rol == RolCliente && clienteID == 0) {
		return Identidad{}, ErrTokenInvalido
	}
	return Identidad{CuentaID: cuenta.ID, Rol: cuenta.Rol, ClienteID: clienteID}, nil
}

// Middleware es el unico origen de rol y cliente_id para las rutas protegidas.
func (a *Autenticacion) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		partes := strings.Fields(c.GetHeader("Authorization"))
		if len(partes) != 2 || !strings.EqualFold(partes[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "se requiere un token Bearer"})
			return
		}
		identidad, err := a.ValidarToken(partes[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrTokenInvalido.Error()})
			return
		}
		c.Set("cuenta_id", identidad.CuentaID)
		c.Set("rol", identidad.Rol)
		c.Set("cliente_id", identidad.ClienteID)
		c.Next()
	}
}

func (a *API) Login(c *gin.Context) {
	var entrada struct {
		Usuario  string `json:"usuario" binding:"required,max=100"`
		Password string `json:"password" binding:"required,max=72"`
	}
	if c.ShouldBindJSON(&entrada) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "usuario y password son obligatorios"})
		return
	}
	cuenta, err := a.Autenticacion.IniciarSesion(entrada.Usuario, entrada.Password)
	if errors.Is(err, ErrCredencialesInvalidas) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrCredencialesInvalidas.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	token, err := a.Autenticacion.EmitirToken(cuenta)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		return
	}
	var clienteID uint
	if cuenta.ClienteID != nil {
		clienteID = *cuenta.ClienteID
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "tipo": "Bearer", "expira_en": int(a.Autenticacion.duracion.Seconds()), "rol": cuenta.Rol, "cliente_id": clienteID})
}
