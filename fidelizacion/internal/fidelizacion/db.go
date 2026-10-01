package fidelizacion

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Conectar abre PostgreSQL con las variables de entorno DB_HOST, DB_PORT,
// DB_USER, DB_PASSWORD y DB_NAME (ver .env.example).
func Conectar() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=America/Guayaquil",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"),
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

// Migrar crea o actualiza las tablas a partir de los structs.
func Migrar(db *gorm.DB) error {
	return db.AutoMigrate(&Cliente{}, &Compra{}, &Recompensa{}, &Canje{}, &MovimientoPuntos{})
}

// Sembrar carga datos de ejemplo solo si la base está vacía.
func Sembrar(db *gorm.DB) error {
	var n int64
	if err := db.Model(&Cliente{}).Count(&n).Error; err != nil || n > 0 {
		return err
	}
	cliente := Cliente{Cedula: "0900000001", Nombre: "Cliente de prueba", Correo: "cliente@example.com"}
	if err := db.Create(&cliente).Error; err != nil {
		return err
	}
	recompensas := []Recompensa{
		{Nombre: "Café americano gratis", Descripcion: "Un café americano de 8 oz", PuntosNecesarios: 40, Activa: true},
		{Nombre: "Postre del día", Descripcion: "Un postre a elección del día", PuntosNecesarios: 60, Activa: true},
		{Nombre: "Descuento de $5", Descripcion: "Descuento de $5 en la siguiente compra", PuntosNecesarios: 100, Activa: true},
	}
	return db.Create(&recompensas).Error
}
