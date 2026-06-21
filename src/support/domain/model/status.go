package model

// Status — estado del ticket en su ciclo de vida. Máquina LINEAL del slice POC
// (la reapertura queda para H2). Transición inválida → ErrTransicionInvalida (P-07).
//
//	abierto → asignado → en_curso → resuelto → cerrado
type Status string

const (
	StatusAbierto  Status = "abierto"
	StatusAsignado Status = "asignado"
	StatusEnCurso  Status = "en_curso"
	StatusResuelto Status = "resuelto"
	StatusCerrado  Status = "cerrado"
)

func (s Status) Valid() bool {
	switch s {
	case StatusAbierto, StatusAsignado, StatusEnCurso, StatusResuelto, StatusCerrado:
		return true
	default:
		return false
	}
}

func (s Status) String() string { return string(s) }
