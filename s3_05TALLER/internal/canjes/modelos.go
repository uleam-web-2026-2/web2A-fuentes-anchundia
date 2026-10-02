package canjes

// Cliente es la entidad del lado del "uno"
type Cliente struct {
	ID     uint
	Nombre string
	Canjes []Canje
}

// Canje es la entidad del lado de los "muchos" y tiene estados
type Canje struct {
	ID         uint
	ClienteID  uint
	Recompensa string
	Estado     string
}

// Lista de estados válidos en minúscula para que no salga del paquete
var estadosValidos = map[string]bool{
	"solicitado": true,
	"aprobado":   true,
	"entregado":  true,
}
