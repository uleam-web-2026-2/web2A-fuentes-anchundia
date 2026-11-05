package fidelizacion

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Estas formas reproducen las tablas previas, sin sucursal_id ni asociaciones nuevas.
type compraHito1 struct {
	ID            uint      `gorm:"primaryKey"`
	ClienteID     uint      `gorm:"not null"`
	Fecha         time.Time `gorm:"not null"`
	Total         float64   `gorm:"not null"`
	NumeroFactura string    `gorm:"not null;unique"`
}

func (compraHito1) TableName() string { return "compras" }

type canjeHito1 struct {
	ID               uint      `gorm:"primaryKey"`
	ClienteID        uint      `gorm:"not null"`
	RecompensaID     uint      `gorm:"not null"`
	PuntosUtilizados int       `gorm:"not null"`
	Fecha            time.Time `gorm:"not null"`
	Estado           string    `gorm:"not null;default:solicitado"`
}

func (canjeHito1) TableName() string { return "canjes" }

type movimientoHito1 struct {
	ID               uint `gorm:"primaryKey"`
	ClienteID        uint `gorm:"not null"`
	CompraID         *uint
	CanjeID          *uint
	Tipo             string    `gorm:"not null"`
	Puntos           int       `gorm:"not null"`
	Fecha            time.Time `gorm:"not null"`
	FechaVencimiento *time.Time
}

func (movimientoHito1) TableName() string { return "movimiento_puntos" }

type registrosHito1 struct {
	clientes    []Cliente
	compras     []compraHito1
	canjes      []canjeHito1
	recompensas []Recompensa
	movimientos []movimientoHito1
}

func leerRegistrosHito1(t *testing.T, db *gorm.DB) registrosHito1 {
	t.Helper()
	var resultado registrosHito1
	for _, destino := range []any{&resultado.clientes, &resultado.compras, &resultado.canjes, &resultado.recompensas, &resultado.movimientos} {
		if err := db.Order("id").Find(destino).Error; err != nil {
			t.Fatal(err)
		}
	}
	return resultado
}

func exigirRegistrosAnterioresSinCambios(t *testing.T, db *gorm.DB, anteriores registrosHito1) {
	t.Helper()
	// Se consultan por ID solamente los registros anteriores: los nuevos datos de prueba
	// no deben ocultar que se alteró una fila existente.
	for _, conjunto := range []any{anteriores.clientes, anteriores.compras, anteriores.canjes, anteriores.recompensas, anteriores.movimientos} {
		valores := reflect.ValueOf(conjunto)
		for i := 0; i < valores.Len(); i++ {
			anterior := valores.Index(i)
			actual := reflect.New(anterior.Type())
			id := anterior.FieldByName("ID").Uint()
			if err := db.First(actual.Interface(), id).Error; err != nil {
				t.Fatalf("se perdió %s ID %d: %v", anterior.Type(), id, err)
			}
			if !reflect.DeepEqual(anterior.Interface(), actual.Elem().Interface()) {
				t.Fatalf("se modificó %s ID %d: anterior=%+v actual=%+v", anterior.Type(), id, anterior.Interface(), actual.Elem().Interface())
			}
		}
	}
}

func baseConDatosDelHito1(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared&_pragma=foreign_keys(1)", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&Cliente{}, &Recompensa{}, &compraHito1{}, &canjeHito1{}, &movimientoHito1{}); err != nil {
		t.Fatal(err)
	}
	fecha := time.Date(2026, time.October, 1, 10, 15, 30, 123000000, time.UTC)
	cliente := Cliente{Cedula: "0900000001", Nombre: "Cliente anterior", Correo: "anterior@example.com", Telefono: "0991112222", FechaRegistro: fecha, Activo: true}
	recompensa := Recompensa{Nombre: "Café anterior", Descripcion: "Catálogo previo", PuntosNecesarios: 40, Activa: true}
	if err := db.Create(&cliente).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&recompensa).Error; err != nil {
		t.Fatal(err)
	}
	compra := compraHito1{ClienteID: cliente.ID, Fecha: fecha, Total: 80.75, NumeroFactura: "001-001-000000123"}
	if err := db.Create(&compra).Error; err != nil {
		t.Fatal(err)
	}
	canje := canjeHito1{ClienteID: cliente.ID, RecompensaID: recompensa.ID, PuntosUtilizados: 40, Fecha: fecha.Add(time.Minute), Estado: EstadoSolicitado}
	if err := db.Create(&canje).Error; err != nil {
		t.Fatal(err)
	}
	vencimiento := fecha.AddDate(0, 6, 0)
	movimiento := movimientoHito1{ClienteID: cliente.ID, CompraID: &compra.ID, Tipo: TipoAcumulacion, Puntos: 80, Fecha: fecha, FechaVencimiento: &vencimiento}
	if err := db.Create(&movimiento).Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func TestMigrarConservaDatosHito1YAsignaSucursalPrincipal(t *testing.T) {
	db := baseConDatosDelHito1(t)
	anteriores := leerRegistrosHito1(t, db)
	if db.Migrator().HasColumn(&compraHito1{}, "sucursal_id") {
		t.Fatal("la tabla inicial ya tenía la columna nueva")
	}
	for i := 0; i < 2; i++ {
		if err := Migrar(db); err != nil {
			t.Fatalf("migración %d: %v", i+1, err)
		}
		exigirRegistrosAnterioresSinCambios(t, db, anteriores)
	}
	var compra Compra
	var canje Canje
	if err := db.First(&compra, anteriores.compras[0].ID).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.First(&canje, anteriores.canjes[0].ID).Error; err != nil {
		t.Fatal(err)
	}
	if compra.SucursalID != 1 || canje.SucursalID != 1 {
		t.Fatalf("faltan referencias de sucursal: compra=%d canje=%d", compra.SucursalID, canje.SucursalID)
	}
	var sucursal Sucursal
	if err := db.First(&sucursal, 1).Error; err != nil || !sucursal.Activa {
		t.Fatalf("falta sucursal principal: %+v %v", sucursal, err)
	}
	if !db.Migrator().HasTable(&Cuenta{}) {
		t.Fatal("falta la tabla de autenticación")
	}
}

func TestSembrarSoloCargaDatosConOpcionExplicita(t *testing.T) {
	e := nuevoEntorno(t)
	t.Setenv("DEMO_DATA", "false")
	if err := Sembrar(e.db); err != nil {
		t.Fatal(err)
	}
	var clientes, cuentas int64
	if err := e.db.Model(&Cliente{}).Count(&clientes).Error; err != nil {
		t.Fatal(err)
	}
	if err := e.db.Model(&Cuenta{}).Count(&cuentas).Error; err != nil {
		t.Fatal(err)
	}
	if clientes != 0 || cuentas != 0 {
		t.Fatalf("se crearon datos demo sin pedirlos: clientes=%d cuentas=%d", clientes, cuentas)
	}
}

func cuentasDemo(t *testing.T, db *gorm.DB) (Cuenta, Cuenta) {
	t.Helper()
	var cliente, admin Cuenta
	if err := db.Where("usuario = ?", "cliente").First(&cliente).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Where("usuario = ?", "admin").First(&admin).Error; err != nil {
		t.Fatal(err)
	}
	if cliente.ClienteID == nil {
		t.Fatal("cuenta cliente sin identidad de negocio")
	}
	return cliente, admin
}

func conteosDemo(t *testing.T, db *gorm.DB) []int64 {
	t.Helper()
	cantidades := make([]int64, 0, 7)
	for _, modelo := range []any{&Sucursal{}, &Cliente{}, &Recompensa{}, &Compra{}, &Canje{}, &MovimientoPuntos{}, &Cuenta{}} {
		var cantidad int64
		if err := db.Model(modelo).Count(&cantidad).Error; err != nil {
			t.Fatal(err)
		}
		cantidades = append(cantidades, cantidad)
	}
	return cantidades
}

func TestSembrarDemoEsIdempotenteYPreparaLasCredenciales(t *testing.T) {
	e := nuevoEntorno(t)
	configurarDemoPrueba(t)
	if err := PrepararBase(e.db); err != nil {
		t.Fatal(err)
	}
	cliente, admin := cuentasDemo(t, e.db)
	antes := leerRegistrosHito1(t, e.db)
	conteos := conteosDemo(t, e.db)
	for i := 0; i < 2; i++ {
		if err := PrepararBase(e.db); err != nil {
			t.Fatalf("preparación repetida: %v", err)
		}
		exigirRegistrosAnterioresSinCambios(t, e.db, antes)
		if actuales := conteosDemo(t, e.db); !reflect.DeepEqual(conteos, actuales) {
			t.Fatalf("se duplicaron datos: antes=%v después=%v", conteos, actuales)
		}
	}
	esperado := []int64{2, 1, 1, 1, 1, 2, 2}
	if !reflect.DeepEqual(conteos, esperado) {
		t.Fatalf("datos demo inesperados: %v, esperaba %v", conteos, esperado)
	}
	clienteDespues, adminDespues := cuentasDemo(t, e.db)
	if !reflect.DeepEqual(cliente, clienteDespues) || !reflect.DeepEqual(admin, adminDespues) {
		t.Fatal("se cambiaron las cuentas o sus hashes al repetir")
	}
	var canje Canje
	if err := e.db.Where("cliente_id = ? AND estado = ?", *cliente.ClienteID, EstadoConfirmado).First(&canje).Error; err != nil {
		t.Fatal(err)
	}
	for _, dato := range []struct {
		cuenta        Cuenta
		password, rol string
	}{
		{cliente, "cliente-2026", RolCliente}, {admin, "admin-2026", RolAdministrador},
	} {
		if dato.cuenta.PasswordHash == dato.password || bcrypt.CompareHashAndPassword([]byte(dato.cuenta.PasswordHash), []byte(dato.password)) != nil {
			t.Fatal("la contraseña demo no se guardó con bcrypt")
		}
		codigo, cuerpo := e.solicitar("POST", "/login", "", fmt.Sprintf(`{"usuario":%q,"password":%q}`, dato.cuenta.Usuario, dato.password))
		exigirCodigo(t, codigo, http.StatusOK, cuerpo)
		if cuerpo["rol"] != dato.rol {
			t.Fatalf("rol incorrecto: %v", cuerpo)
		}
		if _, existe := cuerpo["password_hash"]; existe {
			t.Fatal("login expuso el hash de contraseña")
		}
	}
	codigo, cuerpo := e.llamar("GET", fmt.Sprintf("/clientes/%d/puntos", *cliente.ClienteID), RolAdministrador, 0, "")
	exigirCodigo(t, codigo, http.StatusOK, cuerpo)
	if cuerpo["saldo"] != float64(100000-canje.PuntosUtilizados) {
		t.Fatalf("saldo demo inesperado: %v", cuerpo)
	}
}

func TestPrepararBaseConIdsOcupadosConservaTodoElHito1(t *testing.T) {
	db := baseConDatosDelHito1(t)
	anteriores := leerRegistrosHito1(t, db)
	configurarDemoPrueba(t)
	if err := PrepararBase(db); err != nil {
		t.Fatal(err)
	}
	cliente, _ := cuentasDemo(t, db)
	if cliente.ClienteID == nil || *cliente.ClienteID == anteriores.clientes[0].ID {
		t.Fatal("se reutilizó la identidad del cliente anterior")
	}
	var demo Cliente
	if err := db.First(&demo, *cliente.ClienteID).Error; err != nil {
		t.Fatal(err)
	}
	if demo.Correo != "cliente.hito2@example.invalid" {
		t.Fatalf("cliente de demostración incorrecto: %s", demo.Correo)
	}
	var canje Canje
	if err := db.Where("cliente_id = ? AND estado = ?", demo.ID, EstadoConfirmado).First(&canje).Error; err != nil {
		t.Fatal(err)
	}
	if canje.ID == anteriores.canjes[0].ID {
		t.Fatal("se reutilizó el canje anterior")
	}
	if canje.RecompensaID != anteriores.recompensas[0].ID {
		t.Fatal("no se aprovechó la recompensa activa existente")
	}
	exigirRegistrosAnterioresSinCambios(t, db, anteriores)
	var regalos int64
	if err := db.Model(&Recompensa{}).Count(&regalos).Error; err != nil {
		t.Fatal(err)
	}
	if regalos != int64(len(anteriores.recompensas)) {
		t.Fatal("se duplicó el catálogo existente")
	}
	despuesPrimera := leerRegistrosHito1(t, db)
	conteos := conteosDemo(t, db)
	if err := PrepararBase(db); err != nil {
		t.Fatal(err)
	}
	exigirRegistrosAnterioresSinCambios(t, db, despuesPrimera)
	if actuales := conteosDemo(t, db); !reflect.DeepEqual(conteos, actuales) {
		t.Fatalf("segunda preparación duplicó filas: %v -> %v", conteos, actuales)
	}
}

func TestConflictoDeCuentaDemoConservaCredencialesYReviertePreparacion(t *testing.T) {
	e := nuevoEntorno(t)
	legado := e.cliente(t, "0900000001", "cuenta-original@example.com")
	original := e.cuenta("cliente", "clave-anterior", RolCliente, legado.ID)
	anteriores := leerRegistrosHito1(t, e.db)
	conteos := conteosDemo(t, e.db)
	configurarDemoPrueba(t)
	if err := PrepararBase(e.db); err == nil {
		t.Fatal("se aceptó apropiarse de una cuenta cliente anterior")
	}
	exigirRegistrosAnterioresSinCambios(t, e.db, anteriores)
	if actuales := conteosDemo(t, e.db); !reflect.DeepEqual(conteos, actuales) {
		t.Fatalf("la preparación fallida dejó datos parciales: %v -> %v", conteos, actuales)
	}
	var guardada Cuenta
	if err := e.db.First(&guardada, original.ID).Error; err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(original, guardada) {
		t.Fatal("se cambiaron datos de la cuenta anterior")
	}
	if bcrypt.CompareHashAndPassword([]byte(guardada.PasswordHash), []byte("clave-anterior")) != nil {
		t.Fatal("se cambió la contraseña anterior")
	}
}

func TestConectarRechazaSQLiteSinCrearUnaBase(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "no-crear.db")
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DB_SQLITE_PATH", ruta)
	db, err := Conectar()
	if err == nil || db != nil {
		t.Fatal("el arranque de la aplicación debe exigir PostgreSQL")
	}
	if _, err := os.Stat(ruta); !os.IsNotExist(err) {
		t.Fatalf("se creó o consultó una base SQLite: %v", err)
	}
}
func TestPreparacionFallidaRevierteTambienElEsquema(t *testing.T) {
	db := baseConDatosDelHito1(t)
	anteriores := leerRegistrosHito1(t, db)
	if err := db.AutoMigrate(&Cuenta{}); err != nil {
		t.Fatal(err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("clave-legada"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	id := anteriores.clientes[0].ID
	cuenta := Cuenta{Usuario: "cliente", PasswordHash: string(hash), Rol: RolCliente, ClienteID: &id}
	if err := db.Create(&cuenta).Error; err != nil {
		t.Fatal(err)
	}
	if db.Migrator().HasTable(&Sucursal{}) || db.Migrator().HasColumn(&compraHito1{}, "sucursal_id") {
		t.Fatal("la base legacy ya contenía el esquema nuevo")
	}
	configurarDemoPrueba(t)
	if err := PrepararBase(db); err == nil {
		t.Fatal("la cuenta previa incompatible debió impedir la preparación")
	}
	if db.Migrator().HasTable(&Sucursal{}) || db.Migrator().HasColumn(&compraHito1{}, "sucursal_id") || db.Migrator().HasColumn(&canjeHito1{}, "sucursal_id") {
		t.Fatal("la preparación fallida dejó parte del esquema nuevo")
	}
	exigirRegistrosAnterioresSinCambios(t, db, anteriores)
	var guardada Cuenta
	if err := db.First(&guardada, cuenta.ID).Error; err != nil {
		t.Fatal(err)
	}
	if guardada.PasswordHash != cuenta.PasswordHash || guardada.ClienteID == nil || *guardada.ClienteID != id {
		t.Fatal("la cuenta anterior cambió")
	}
}

// Credenciales ficticias exclusivas de las pruebas; el servidor las lee del entorno.
func configurarDemoPrueba(t *testing.T) {
    t.Helper()
    t.Setenv("DEMO_DATA", "true")
    t.Setenv("DEMO_CLIENT_PASSWORD", "cliente-2026")
    t.Setenv("DEMO_ADMIN_PASSWORD", "admin-2026")
}

func TestDemoSinPasswordsReviertePreparacion(t *testing.T) {
    e := nuevoEntorno(t)
    t.Setenv("DEMO_DATA", "true")
    t.Setenv("DEMO_CLIENT_PASSWORD", "")
    t.Setenv("DEMO_ADMIN_PASSWORD", "")
    antes := conteosDemo(t, e.db)
    if err := PrepararBase(e.db); err == nil {
        t.Fatal("se aceptó sembrar cuentas sin contraseñas configuradas")
    }
    if despues := conteosDemo(t, e.db); !reflect.DeepEqual(antes, despues) {
        t.Fatalf("la preparación fallida dejó filas: %v -> %v", antes, despues)
    }
}
