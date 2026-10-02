package canjes

import "gorm.io/gorm"

// Sembrar carga datos de ejemplo solo si la tabla está vacía.
func Sembrar(db *gorm.DB) {
	var total int64
	db.Model(&Cliente{}).Count(&total)
	if total > 0 {
		return
	}

	datos := []Cliente{
		{
			Nombre: "Miguel Fuentes",
			Canjes: []Canje{
				{Recompensa: "Café Americano", Estado: "solicitado"},
			},
		},
		{
			Nombre: "Luis Anchundia",
			Canjes: []Canje{
				{Recompensa: "Frappuccino", Estado: "aprobado"},
				{Recompensa: "Muffin de Arándanos", Estado: "entregado"},
			},
		},
	}

	db.Create(&datos) // Guarda los clientes con sus respectivos canjes
}
