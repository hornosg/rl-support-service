package valueobject

import "errors"

// Priority — prioridad del ticket (G-05 informa el SLA, fuera del slice). Value Object inmutable.
type Priority string

const (
	PriorityBaja    Priority = "baja"
	PriorityMedia   Priority = "media"
	PriorityAlta    Priority = "alta"
	PriorityUrgente Priority = "urgente"
)

// PriorityDefault — prioridad asumida cuando el canal no la informa.
const PriorityDefault = PriorityMedia

var ErrPrioridadInvalida = errors.New("valueobject: prioridad inválida")

// NewPriority valida contra el conjunto cerrado. Cadena vacía → PriorityDefault.
func NewPriority(s string) (Priority, error) {
	if s == "" {
		return PriorityDefault, nil
	}
	switch Priority(s) {
	case PriorityBaja, PriorityMedia, PriorityAlta, PriorityUrgente:
		return Priority(s), nil
	default:
		return "", ErrPrioridadInvalida
	}
}

func (p Priority) String() string { return string(p) }
