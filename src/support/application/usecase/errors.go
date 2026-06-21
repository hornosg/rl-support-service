package usecase

import "errors"

// ErrFiltroEstadoInvalido — el filtro `estado` del listado no es un estado válido.
var ErrFiltroEstadoInvalido = errors.New("usecase: filtro de estado inválido")

const (
	defaultLimit = 20
	maxLimit     = 100
)

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}
