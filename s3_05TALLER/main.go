package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"s3_05TALLER/internal/canjes"

	"github.com/go-chi/chi/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	reset := flag.Bool("reset", false, "Borra y recrea la base de datos")
	flag.Parse()

	// 1. Conexión a PostgreSQL (Asegúrate de que el password sea el tuyo si no es 12345)
	dsn := "host=localhost user=postgres password=123 dbname=CAFETERIA port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Fallo al conectar a la base de datos: ", err)
	}

	// 2. Si usamos -reset, borramos las tablas viejas
	if *reset {
		fmt.Println("Borrando tablas antiguas...")
		db.Migrator().DropTable(&canjes.Canje{}, &canjes.Cliente{})
	}

	// 3. Crear las tablas
	fmt.Println("Migrando base de datos...")
	db.AutoMigrate(&canjes.Cliente{}, &canjes.Canje{})

	// 4. Cargar datos de prueba
	canjes.Sembrar(db)

	// 5. Configurar el enrutador
	r := chi.NewRouter()

	manejador := &canjes.Manejador{DB: db}

	// Rutas de tu cafetería
	r.Post("/canjes", manejador.Crear)
	r.Get("/canjes", manejador.Listar)
	r.Put("/canjes/{id}", manejador.Actualizar)
	r.Delete("/canjes/{id}", manejador.Borrar)

	// Ruta para solucionar el N+1 viendo clientes y sus canjes
	r.Get("/clientes", manejador.ListarClientes)

	// 6. Arrancar el servidor
	fmt.Println("Servidor corriendo en http://localhost:8080")
	http.ListenAndServe(":8080", r)
}
