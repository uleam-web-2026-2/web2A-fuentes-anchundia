package fidelizacion

import (
	"errors"
	"gorm.io/gorm"
)

// Reutiliza la clave demo del cliente, configurada en el entorno privado.
// No reemplaza cuentas ni clientes que ya tengan una configuración distinta.
func sembrarSegundoCliente(tx *gorm.DB, password string) error {
	var cliente Cliente
	err := tx.Where("correo = ?", "cliente.segundo@example.invalid").First(&cliente).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		cliente = Cliente{Cedula: "9999999998", Nombre: "Segundo cliente ficticio", Correo: "cliente.segundo@example.invalid", Activo: true}
		if err := tx.Create(&cliente).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if cliente.Cedula != "9999999998" || !cliente.Activo {
		return errors.New("el segundo cliente demo existente no coincide; no se modificó")
	}
	return asegurarCuentaDemo(tx, "cliente2", password, RolCliente, &cliente.ID)
}
