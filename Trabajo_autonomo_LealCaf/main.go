package main

import (
	"log"

	"fidelizacion/internal/fidelizacion"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load() // lee .env si existe

	db, err := fidelizacion.Conectar()
	if err != nil {
		log.Fatalf("no se pudo conectar a la base de datos: %v", err)
	}
	if err := fidelizacion.Migrar(db); err != nil {
		log.Fatalf("no se pudo migrar: %v", err)
	}
	if err := fidelizacion.Sembrar(db); err != nil {
		log.Fatalf("no se pudieron cargar los datos de ejemplo: %v", err)
	}

	r := gin.Default()
	fidelizacion.ConfigurarRutas(r, db)

	log.Println("Servidor corriendo en el puerto 8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
