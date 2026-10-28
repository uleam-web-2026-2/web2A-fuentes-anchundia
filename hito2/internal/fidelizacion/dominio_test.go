package fidelizacion

import "testing"

// La máquina de estados se prueba sin servidor ni base de datos.
func TestValidarTransicionDelCanje(t *testing.T) {
	for _, caso := range []struct {
		actual, destino string
		permitida       bool
	}{
		{EstadoSolicitado, EstadoConfirmado, true},
		{EstadoSolicitado, EstadoCancelado, true},
		{EstadoSolicitado, EstadoSolicitado, false},
		{EstadoConfirmado, EstadoCancelado, false},
		{EstadoConfirmado, EstadoSolicitado, false},
		{EstadoConfirmado, EstadoConfirmado, false},
		{EstadoCancelado, EstadoConfirmado, false},
		{EstadoCancelado, EstadoSolicitado, false},
		{EstadoCancelado, EstadoCancelado, false},
		{EstadoSolicitado, "desconocido", false},
		{"desconocido", EstadoConfirmado, false},
	} {
		t.Run(caso.actual+"_a_"+caso.destino, func(t *testing.T) {
			err := ValidarTransicion(caso.actual, caso.destino)
			if (err == nil) != caso.permitida {
				t.Fatalf("permitida=%t, error=%v", caso.permitida, err)
			}
		})
	}
}
