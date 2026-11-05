package fidelizacion

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	correoDemoHito2  = "cliente.hito2@example.invalid"
	cedulaDemoHito2  = "9999999999"
	facturaDemoHito2 = "DEMO-HITO2-INICIAL-2026"
	puntosDemoHito2  = 100000
)

// Conectar usa PostgreSQL. SQLite solo se utiliza dentro de las pruebas aisladas.
func Conectar() (*gorm.DB, error) {
	driver := strings.ToLower(strings.TrimSpace(os.Getenv("DB_DRIVER")))
	if driver != "" && driver != "postgres" {
		return nil, fmt.Errorf("DB_DRIVER debe ser postgres; el proyecto usa PostgreSQL")
	}
	if strings.TrimSpace(os.Getenv("DB_NAME")) == "" {
		return nil, errors.New("DB_NAME es obligatorio; configure la base PostgreSQL existente")
	}
	host, puerto := os.Getenv("DB_HOST"), os.Getenv("DB_PORT")
	if host == "" {
		host = "localhost"
	}
	if puerto == "" {
		puerto = "5432"
	}
	conexion := url.URL{Scheme: "postgres", Host: net.JoinHostPort(host, puerto), Path: "/" + os.Getenv("DB_NAME"), User: url.UserPassword(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"))}
	// La barra es válida en una query URI. GORM lee TimeZone sin decodificar el DSN.
	conexion.RawQuery = "sslmode=disable&TimeZone=America/Guayaquil"
	return gorm.Open(postgres.Open(conexion.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
}

// PrepararBase hace indivisibles la migración y los datos de demostración.
// Si algo entra en conflicto, PostgreSQL revierte ambos cambios.
func PrepararBase(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := bloquearPreparacion(tx); err != nil {
			return err
		}
		if err := migrarDentro(tx); err != nil {
			return err
		}
		if strings.EqualFold(os.Getenv("DEMO_DATA"), "true") {
			return sembrarDentro(tx)
		}
		return nil
	})
}

func bloquearPreparacion(tx *gorm.DB) error {
	if tx.Dialector.Name() != "postgres" {
		return nil
	}
	// Serializa los arranques de esta aplicación sin bloquear otras bases.
	return tx.Exec("SELECT pg_advisory_xact_lock(20261028, 2)").Error
}

// Migrar actualiza el esquema sin borrar ni recrear las tablas existentes.
func Migrar(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := bloquearPreparacion(tx); err != nil {
			return err
		}
		return migrarDentro(tx)
	})
}

func migrarDentro(tx *gorm.DB) error {
	if err := tx.AutoMigrate(&Sucursal{}); err != nil {
		return err
	}
	principal := Sucursal{ID: 1, Nombre: "Centro", Direccion: "Local principal; actualizar dirección", Activa: true}
	if err := tx.Where("id = ?", 1).FirstOrCreate(&principal).Error; err != nil {
		return err
	}
	// SucursalID tiene default:1: los registros anteriores se atribuyen al local principal.
	if err := tx.AutoMigrate(&Cliente{}, &Compra{}, &Recompensa{}, &Canje{}, &MovimientoPuntos{}, &Cuenta{}); err != nil {
		return err
	}
	return sincronizarSecuencias(tx)
}

// Sembrar solo añade registros dedicados de demostración, sin utilizar IDs históricos.
func Sembrar(db *gorm.DB) error {
	if !strings.EqualFold(os.Getenv("DEMO_DATA"), "true") {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := bloquearPreparacion(tx); err != nil {
			return err
		}
		return sembrarDentro(tx)
	})
}

func sembrarDentro(tx *gorm.DB) error {
	passwordCliente := os.Getenv("DEMO_CLIENT_PASSWORD")
	passwordAdmin := os.Getenv("DEMO_ADMIN_PASSWORD")
	if strings.TrimSpace(passwordCliente) == "" || strings.TrimSpace(passwordAdmin) == "" {
		return errors.New("DEMO_DATA=true requiere DEMO_CLIENT_PASSWORD y DEMO_ADMIN_PASSWORD en el entorno")
	}
	if len(passwordCliente) > 72 || len(passwordAdmin) > 72 {
		return errors.New("las contraseñas de demostración no pueden superar 72 bytes")
	}
	norte := Sucursal{ID: 2, Nombre: "Norte", Direccion: "Local Norte de demostración", Activa: true}
	if err := tx.Where("id = ?", 2).FirstOrCreate(&norte).Error; err != nil {
		return err
	}
	if !norte.Activa {
		return errors.New("la sucursal de demostración Norte está inactiva; no se modificó")
	}
	if err := sincronizarSecuencias(tx); err != nil {
		return err
	}

	cliente, err := clienteDemostracion(tx)
	if err != nil {
		return err
	}
	for _, dato := range []struct {
		usuario, password, rol string
		clienteID              *uint
	}{
		{"cliente", passwordCliente, RolCliente, &cliente.ID},
		{"admin", passwordAdmin, RolAdministrador, nil},
	} {
		if err := asegurarCuentaDemo(tx, dato.usuario, dato.password, dato.rol, dato.clienteID); err != nil {
			return err
		}
	}

	var recompensa Recompensa
	err = tx.Where("activa = ? AND puntos_necesarios > 0", true).Order("id").First(&recompensa).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		recompensa = Recompensa{Nombre: "Café de demostración Hito 2", Descripcion: "Recompensa ficticia para practicar la exposición", PuntosNecesarios: 40, Activa: true}
		if err := tx.Create(&recompensa).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	var compra Compra
	err = tx.Where("numero_factura = ?", facturaDemoHito2).First(&compra).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		compra = Compra{ClienteID: cliente.ID, SucursalID: 1, Total: puntosDemoHito2, NumeroFactura: facturaDemoHito2, Fecha: ahora()}
		if err := tx.Create(&compra).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if compra.ClienteID != cliente.ID || compra.SucursalID != 1 || compra.Total != puntosDemoHito2 {
		return errors.New("la factura reservada de demostración ya pertenece a otros datos; no se modificó")
	}
	var acumulacion MovimientoPuntos
	err = tx.Where("compra_id = ? AND tipo = ?", compra.ID, TipoAcumulacion).First(&acumulacion).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		vence := compra.Fecha.AddDate(0, MesesVigencia, 0)
		acumulacion = MovimientoPuntos{ClienteID: cliente.ID, CompraID: &compra.ID, Tipo: TipoAcumulacion, Puntos: puntosDemoHito2, Fecha: compra.Fecha, FechaVencimiento: &vence}
		if err := tx.Create(&acumulacion).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if acumulacion.ClienteID != cliente.ID || acumulacion.Puntos != puntosDemoHito2 || acumulacion.FechaVencimiento == nil {
		return errors.New("el movimiento de demostración existente no coincide con su compra; no se modificó")
	}

	var confirmado Canje
	err = tx.Where("cliente_id = ? AND estado = ?", cliente.ID, EstadoConfirmado).Order("id").First(&confirmado).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		saldo, err := (&RepositorioGORM{DB: tx}).Saldo(cliente.ID)
		if err != nil {
			return err
		}
		if saldo < recompensa.PuntosNecesarios {
			return errors.New("el cliente de demostración no tiene saldo vigente suficiente para preparar el canje confirmado")
		}
		confirmado = Canje{ClienteID: cliente.ID, RecompensaID: recompensa.ID, SucursalID: norte.ID, PuntosUtilizados: recompensa.PuntosNecesarios, Fecha: ahora(), Estado: EstadoConfirmado}
		if err := tx.Create(&confirmado).Error; err != nil {
			return err
		}
		movimiento := MovimientoPuntos{ClienteID: cliente.ID, CanjeID: &confirmado.ID, Tipo: TipoCanje, Puntos: -confirmado.PuntosUtilizados, Fecha: confirmado.Fecha}
		if err := tx.Create(&movimiento).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	// Un canje ya confirmado debe conservar su descuento; nunca lo repetimos.
	var descuento MovimientoPuntos
	if err := tx.Where("canje_id = ? AND tipo = ?", confirmado.ID, TipoCanje).First(&descuento).Error; err != nil {
		return errors.New("el canje confirmado de demostración no tiene su movimiento de descuento; no se modificó")
	}
	if descuento.ClienteID != cliente.ID || descuento.Puntos != -confirmado.PuntosUtilizados {
		return errors.New("el descuento del canje de demostración no coincide con sus datos; no se modificó")
	}
	return sincronizarSecuencias(tx)
}

func clienteDemostracion(tx *gorm.DB) (Cliente, error) {
	var cliente Cliente
	err := tx.Where("correo = ?", correoDemoHito2).First(&cliente).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		var cantidad int64
		if err := tx.Model(&Cliente{}).Where("cedula = ?", cedulaDemoHito2).Count(&cantidad).Error; err != nil {
			return cliente, err
		}
		if cantidad != 0 {
			return cliente, errors.New("la cédula ficticia reservada ya existe con otro correo; no se modificó ese cliente")
		}
		cliente = Cliente{Cedula: cedulaDemoHito2, Nombre: "Cliente de demostración Hito 2", Correo: correoDemoHito2, Activo: true}
		if err := tx.Create(&cliente).Error; err != nil {
			return cliente, err
		}
	} else if err != nil {
		return cliente, err
	}
	if cliente.Cedula != cedulaDemoHito2 || !cliente.Activo {
		return cliente, errors.New("el cliente reservado de demostración tiene otra identificación o está inactivo; no se modificó")
	}
	return cliente, nil
}

func asegurarCuentaDemo(tx *gorm.DB, usuario, password, rol string, clienteID *uint) error {
	var cuenta Cuenta
	err := tx.Where("usuario = ?", usuario).First(&cuenta).Error
	if err == nil {
		mismoCliente := (clienteID == nil && cuenta.ClienteID == nil) || (clienteID != nil && cuenta.ClienteID != nil && *clienteID == *cuenta.ClienteID)
		if cuenta.Rol != rol || !mismoCliente || bcrypt.CompareHashAndPassword([]byte(cuenta.PasswordHash), []byte(password)) != nil {
			return fmt.Errorf("la cuenta %s ya existe con otra configuración o contraseña; no se reemplazó", usuario)
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	cuenta = Cuenta{Usuario: usuario, PasswordHash: string(hash), Rol: rol, ClienteID: clienteID}
	return tx.Create(&cuenta).Error
}

// Corrige únicamente secuencias atrasadas. Una secuencia adelantada no se reduce.
func sincronizarSecuencias(tx *gorm.DB) error {
	if tx.Dialector.Name() != "postgres" {
		return nil
	}
	for _, modelo := range []string{"Sucursal", "Cliente", "Compra", "Recompensa", "Canje", "MovimientoPuntos", "Cuenta"} {
		tabla := tx.NamingStrategy.TableName(modelo)
		var secuencia string
		consulta := "SELECT format('%I.%I', n.nspname, c.relname) FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE c.oid = pg_get_serial_sequence(?, 'id')::regclass"
		if err := tx.Raw(consulta, tabla).Scan(&secuencia).Error; err != nil {
			return err
		}
		if secuencia == "" {
			continue
		}
		var estado struct {
			LastValue int64
			IsCalled  bool
		}
		// El identificador viene escapado con format('%I.%I', ...) por PostgreSQL.
		if err := tx.Raw("SELECT last_value, is_called FROM " + secuencia).Scan(&estado).Error; err != nil {
			return err
		}
		var maximo int64
		if err := tx.Table(tabla).Select("COALESCE(MAX(id), 0)").Scan(&maximo).Error; err != nil {
			return err
		}
		if maximo > 0 && (maximo > estado.LastValue || (maximo == estado.LastValue && !estado.IsCalled)) {
			if err := tx.Exec("SELECT setval(?::regclass, ?, true)", secuencia, maximo).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
