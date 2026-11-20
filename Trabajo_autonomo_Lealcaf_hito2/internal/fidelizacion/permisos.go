package fidelizacion

import (
	_ "embed"
	"strings"

	"gorm.io/gorm"
)

//go:embed permisos.sql
var permisosSQL string

// PrepararConPermisos solo se invoca con las credenciales de migración.
func PrepararConPermisos(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := PrepararBase(tx); err != nil {
			return err
		}
		for _, statement := range strings.Split(permisosSQL, ";") {
			if strings.TrimSpace(statement) != "" {
				if err := tx.Exec(statement).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}
