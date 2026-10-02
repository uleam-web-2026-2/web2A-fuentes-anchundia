// main.go
package main

import (
	"log"
	"net/http"

	"web2A-fuentes-anchundia/internal/middleware"
	"web2A-fuentes-anchundia/internal/respuesta"
	"web2A-fuentes-anchundia/internal/tickets"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	// Los middleware se registran ANTES de las rutas
	r.Use(middleware.Registro)
	r.Use(middleware.Recuperacion)

	// Sobrescribir 404 y 405 en formato JSON unificado
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		respuesta.Error(w, http.StatusNotFound, "ruta_inexistente", "la ruta no existe")
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		respuesta.Error(w, http.StatusMethodNotAllowed, "metodo_no_permitido", "el método no está permitido en esta ruta")
	})

	almacen := tickets.NuevoAlmacen()

	// Registrar rutas del taller
	r.Get("/tickets", almacen.Listar)
	r.Post("/tickets", almacen.Crear)
	r.Get("/tickets/{id}", almacen.Obtener)
	r.Get("/explotar", tickets.Explotar) // Ruta de prueba para pánicos

	log.Println("mesa de ayuda en :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
