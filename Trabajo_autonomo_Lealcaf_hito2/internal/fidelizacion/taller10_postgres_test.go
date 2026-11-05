package fidelizacion

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Usa un esquema temporal propio; nunca migra ni elimina las tablas public.
func postgresTaller10(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("LEALCAF_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("configure LEALCAF_TEST_POSTGRES_DSN con una base de pruebas desechable")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("no se pudo abrir PostgreSQL de pruebas")
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	esquema := fmt.Sprintf("taller10_%d", time.Now().UnixNano())
	if err := db.Exec("CREATE SCHEMA " + esquema).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Exec("DROP SCHEMA " + esquema + " CASCADE").Error; err != nil {
			t.Error(err)
		}
	})
	aislada, err := gorm.Open(postgres.Open(dsn+" search_path="+esquema), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("use un DSN de pruebas en formato clave=valor")
	}
	conexiones, err := aislada.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conexiones.Close() })
	if err := Migrar(aislada); err != nil {
		t.Fatal(err)
	}
	return aislada
}

func repositorioPrueba(t *testing.T, db *gorm.DB, nombre string) Repositorio {
	t.Helper()
	base := &RepositorioGORM{DB: db}
	if nombre == "gorm" {
		return base
	}
	conexiones, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	return NuevoRepositorioSqlc(base, conexiones)
}

func TestTaller10TransaccionesPostgres(t *testing.T) {
	for _, adaptador := range []string{"gorm", "sqlc"} {
		t.Run(adaptador, func(t *testing.T) {
			db := postgresTaller10(t)
			cliente := Cliente{Cedula: "0900000010", Nombre: "Prueba taller 10", Correo: "taller10@example.invalid", Activo: true}
			if err := db.Create(&cliente).Error; err != nil {
				t.Fatal(err)
			}
			recompensa := Recompensa{Nombre: "Cafe de prueba", PuntosNecesarios: 40, Activa: true}
			if err := db.Create(&recompensa).Error; err != nil {
				t.Fatal(err)
			}
			fecha := time.Now()
			vence := fecha.AddDate(0, 6, 0)
			credito := MovimientoPuntos{ClienteID: cliente.ID, Tipo: TipoAcumulacion, Puntos: 100, Fecha: fecha, FechaVencimiento: &vence}
			if err := db.Create(&credito).Error; err != nil {
				t.Fatal(err)
			}
			servicio := &Servicio{Repo: repositorioPrueba(t, db, adaptador)}
			canje, err := servicio.SolicitarCanje(cliente.ID, EntradaCanje{RecompensaID: recompensa.ID, SucursalID: 1})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := servicio.ObtenerCanje(9999999, RolAdministrador, 0); !errors.Is(err, ErrNoEncontrado) {
				t.Fatalf("se perdió la traducción de no encontrado: %v", err)
			}
			// La primera escritura es válida; SOLO la segunda falla por un CHECK real.
			if err := db.Exec("ALTER TABLE movimiento_puntos ADD CONSTRAINT taller10_fallo_canje CHECK (tipo <> 'canje')").Error; err != nil {
				t.Fatal(err)
			}
			_, err = servicio.ActualizarEstadoCanje(canje.ID, EstadoConfirmado)
			var estadoSQL interface{ SQLState() string }
			if !errors.As(err, &estadoSQL) || estadoSQL.SQLState() != "23514" {
				t.Fatalf("no se forzó el CHECK en la segunda escritura: %v", err)
			}
			var guardado Canje
			if err := db.First(&guardado, canje.ID).Error; err != nil {
				t.Fatal(err)
			}
			var descuentos int64
			if err := db.Model(&MovimientoPuntos{}).Where("canje_id = ?", canje.ID).Count(&descuentos).Error; err != nil {
				t.Fatal(err)
			}
			saldo, err := servicio.Repo.Saldo(cliente.ID)
			if err != nil || guardado.Estado != EstadoSolicitado || descuentos != 0 || saldo != 100 {
				t.Fatalf("rollback incompleto: estado=%s descuentos=%d saldo=%d err=%v", guardado.Estado, descuentos, saldo, err)
			}
			t.Logf("adaptador=%s fallo forzado SQLSTATE=23514: estado=%s movimientos_canje=%d saldo=%d; rollback completo", adaptador, guardado.Estado, descuentos, saldo)
			if err := db.Exec("ALTER TABLE movimiento_puntos DROP CONSTRAINT taller10_fallo_canje").Error; err != nil {
				t.Fatal(err)
			}
			if _, err := servicio.ActualizarEstadoCanje(canje.ID, EstadoConfirmado); err != nil {
				t.Fatal(err)
			}
			if err := db.Model(&MovimientoPuntos{}).Where("canje_id = ?", canje.ID).Count(&descuentos).Error; err != nil {
				t.Fatal(err)
			}
			saldo, err = servicio.Repo.Saldo(cliente.ID)
			if err != nil || descuentos != 1 || saldo != 60 {
				t.Fatalf("commit incorrecto: descuentos=%d saldo=%d err=%v", descuentos, saldo, err)
			}
			t.Logf("adaptador=%s reintento válido: movimientos_canje=%d saldo=%d; commit completo", adaptador, descuentos, saldo)

			// RegistrarCompra también escribe dos filas y conserva la misma garantía.
			if err := db.Exec("ALTER TABLE movimiento_puntos ADD CONSTRAINT taller10_fallo_compra CHECK (tipo <> 'acumulacion') NOT VALID").Error; err != nil {
				t.Fatal(err)
			}
			entrada := EntradaCompra{ClienteID: cliente.ID, SucursalID: 1, Total: 50, NumeroFactura: "TALLER10-ROLLBACK"}
			_, err = servicio.RegistrarCompra(entrada)
			if !errors.As(err, &estadoSQL) || estadoSQL.SQLState() != "23514" {
				t.Fatalf("no se forzó fallo del crédito: %v", err)
			}
			var compras int64
			if err := db.Model(&Compra{}).Where("numero_factura = ?", entrada.NumeroFactura).Count(&compras).Error; err != nil {
				t.Fatal(err)
			}
			saldo, err = servicio.Repo.Saldo(cliente.ID)
			if err != nil || compras != 0 || saldo != 60 {
				t.Fatalf("compra parcial: compras=%d saldo=%d err=%v", compras, saldo, err)
			}
			t.Logf("adaptador=%s compra fallo SQLSTATE=23514: compras=%d saldo=%d; rollback completo", adaptador, compras, saldo)
			if err := db.Exec("ALTER TABLE movimiento_puntos DROP CONSTRAINT taller10_fallo_compra").Error; err != nil {
				t.Fatal(err)
			}
			resultado, err := servicio.RegistrarCompra(entrada)
			if err != nil || resultado.PuntosGenerados != 50 {
				t.Fatalf("reintento compra: %+v %v", resultado, err)
			}
			saldo, err = servicio.Repo.Saldo(cliente.ID)
			if err != nil || saldo != 110 {
				t.Fatalf("compra sin crédito: %d %v", saldo, err)
			}
			t.Logf("adaptador=%s reintento compra: puntos_generados=50 saldo=%d; commit completo", adaptador, saldo)
		})
	}
}
