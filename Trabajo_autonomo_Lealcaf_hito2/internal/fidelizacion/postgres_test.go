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

// Esta prueba es opcional: necesita una base desechable autorizada explícitamente.
// Usa un esquema propio; nunca migra ni elimina las tablas del esquema public.
func TestPostgresConfirmacionesConcurrentes(t *testing.T) {
	for _, adaptador := range []string{"gorm", "sqlc"} {
		t.Run(adaptador, func(t *testing.T) { probarConfirmacionesConcurrentes(t, adaptador) })
	}
}

func probarConfirmacionesConcurrentes(t *testing.T, adaptador string) {
	dsn := os.Getenv("LEALCAF_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("configure LEALCAF_TEST_POSTGRES_DSN con una base PostgreSQL de pruebas")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("no se pudo conectar a PostgreSQL de pruebas")
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pool.Close() })
	schema := fmt.Sprintf("lealcaf_test_%d", time.Now().UnixNano())
	if err := db.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Exec("DROP SCHEMA " + schema + " CASCADE").Error; err != nil {
			t.Errorf("no se pudo limpiar el esquema de pruebas: %v", err)
		}
	})
	// El search_path debe aplicarse a todas las conexiones, incluidas las concurrentes.
	config := postgres.Config{DSN: dsn + " search_path=" + schema}
	isolated, err := gorm.Open(postgres.New(config), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("no se pudo abrir el esquema aislado; use un DSN en formato clave=valor")
	}
	isolatedPool, err := isolated.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = isolatedPool.Close() })
	t.Setenv("DEMO_DATA", "true")
	t.Setenv("DEMO_CLIENT_PASSWORD", "cliente-2026")
	t.Setenv("DEMO_ADMIN_PASSWORD", "admin-2026")
	if err := PrepararBase(isolated); err != nil {
		t.Fatal(err)
	}
	before := conteosDemo(t, isolated)
	if err := PrepararBase(isolated); err != nil {
		t.Fatal(err)
	}
	after := conteosDemo(t, isolated)
	for i := range before {
		if before[i] != after[i] {
			t.Fatalf("la preparación duplicó registros: %v -> %v", before, after)
		}
	}
	// Un cliente aparte tiene saldo para un canje, pero nunca para ambos.
	client := Cliente{Cedula: "0900000099", Nombre: "Concurrencia", Correo: "concurrencia@example.invalid"}
	if err := isolated.Create(&client).Error; err != nil {
		t.Fatal(err)
	}
	reward := Recompensa{Nombre: "Prueba concurrente", PuntosNecesarios: 40}
	if err := isolated.Create(&reward).Error; err != nil {
		t.Fatal(err)
	}
	expires := time.Now().AddDate(0, 6, 0)
	credit := MovimientoPuntos{ClienteID: client.ID, Tipo: TipoAcumulacion, Puntos: 60, Fecha: time.Now(), FechaVencimiento: &expires}
	if err := isolated.Create(&credit).Error; err != nil {
		t.Fatal(err)
	}
	service := &Servicio{Repo: repositorioPrueba(t, isolated, adaptador)}
	first, err := service.SolicitarCanje(client.ID, EntradaCanje{RecompensaID: reward.ID, SucursalID: 1})
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.SolicitarCanje(client.ID, EntradaCanje{RecompensaID: reward.ID, SucursalID: 2})
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	for _, id := range []uint{first.ID, second.ID} {
		go func(id uint) {
			<-start
			_, err := service.ActualizarEstadoCanje(id, EstadoConfirmado)
			results <- err
		}(id)
	}
	close(start)
	success, rejected := 0, 0
	for i := 0; i < 2; i++ {
		err := <-results
		switch {
		case err == nil:
			success++
		case errors.Is(err, ErrPuntosInsuficientes):
			rejected++
		default:
			t.Fatalf("confirmación concurrente falló: %v", err)
		}
	}
	balance, err := service.Repo.Saldo(client.ID)
	if err != nil || success != 1 || rejected != 1 || balance != 20 {
		t.Fatalf("saldo gastado dos veces: éxitos=%d rechazos=%d saldo=%d err=%v", success, rejected, balance, err)
	}
	var debits int64
	if err := isolated.Model(&MovimientoPuntos{}).Where("cliente_id = ? AND tipo = ?", client.ID, TipoCanje).Count(&debits).Error; err != nil || debits != 1 {
		t.Fatalf("se esperaba un único descuento, hay %d: %v", debits, err)
	}
}
