package model

import "errors"

// Errores de dominio del agregado Ticket. Los use cases los propagan; el borde HTTP
// los mapea a Problem Details (RFC 7807).
var (
	ErrAsuntoVacio        = errors.New("ticket: asunto requerido")
	ErrTransicionInvalida = errors.New("ticket: transición de estado inválida")
	ErrSinOperador        = errors.New("ticket: no se puede tomar un ticket sin operador asignado")
	ErrAccionDesconocida  = errors.New("ticket: acción de transición desconocida")
)
