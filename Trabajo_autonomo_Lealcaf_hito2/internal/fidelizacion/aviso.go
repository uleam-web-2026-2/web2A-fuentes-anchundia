package fidelizacion

// Aviso es un hecho confirmado del negocio; no contiene credenciales.
type Aviso struct {
	Tipo   string `json:"tipo"`
	ID     uint   `json:"id"`
	Estado string `json:"estado,omitempty"`
}

// Avisador separa el caso de uso del transporte. La identidad es ClienteID.
type Avisador interface {
	Enviar(clienteID uint, aviso Aviso)
}

// AvisarCon se configura una vez, antes de atender peticiones.
func (s *Servicio) AvisarCon(avisador Avisador) { s.avisos = avisador }
