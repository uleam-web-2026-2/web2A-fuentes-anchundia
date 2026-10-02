package canjes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type Manejador struct {
	DB *gorm.DB
}

// Helper para leer el ID de la URL
func leerID(r *http.Request) uint {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)
	return uint(id)
}

// POST /canjes - Crea un canje (Regla de negocio aplicada)
func (m *Manejador) Crear(w http.ResponseWriter, r *http.Request) {
	var t Canje
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Decisión D1: El servidor fuerza el estado inicial por seguridad
	t.Estado = "solicitado"

	// Validación: Verificar que el cliente realmente exista en PostgreSQL
	var cliente Cliente
	if err := m.DB.First(&cliente, t.ClienteID).Error; err != nil {
		http.Error(w, "El cliente no existe", http.StatusUnprocessableEntity)
		return
	}

	m.DB.Create(&t)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

// GET /canjes - Filtro seguro sin inyección SQL
func (m *Manejador) Listar(w http.ResponseWriter, r *http.Request) {
	var lista []Canje
	estado := r.URL.Query().Get("estado")

	consulta := m.DB
	if estado != "" {
		// El uso de '?' evita la inyección SQL
		consulta = consulta.Where("estado = ?", estado)
	}

	consulta.Find(&lista)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lista)
}

// GET /clientes - Optimización del problema N+1
func (m *Manejador) ListarClientes(w http.ResponseWriter, r *http.Request) {
	var clientes []Cliente
	// Preload extrae todo haciendo solo 2 consultas en lugar de 1 por fila
	m.DB.Debug().Preload("Canjes").Find(&clientes)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clientes)
}

// PUT /canjes/{id} - Actualizar estado validando el map
func (m *Manejador) Actualizar(w http.ResponseWriter, r *http.Request) {
	var datos Canje
	if err := json.NewDecoder(r.Body).Decode(&datos); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Validación: Verificar que el estado enviado sea válido en el map
	if !estadosValidos[datos.Estado] {
		http.Error(w, "Estado no permitido", http.StatusUnprocessableEntity)
		return
	}

	var t Canje
	if err := m.DB.First(&t, leerID(r)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Canje no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "Error del servidor", http.StatusInternalServerError)
		return
	}

	t.Estado = datos.Estado
	m.DB.Save(&t)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

// DELETE /canjes/{id} - Borrar
func (m *Manejador) Borrar(w http.ResponseWriter, r *http.Request) {
	res := m.DB.Delete(&Canje{}, leerID(r))
	if res.Error == nil && res.RowsAffected == 0 {
		http.Error(w, "Canje no encontrado", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
