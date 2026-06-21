// Package mother — Object Mothers: builders de datos de test del dominio de tickets.
package mother

import (
	"github.com/google/uuid"

	"support-service/src/support/domain/model"
	"support-service/src/support/domain/valueobject"
)

// ValidSolicitante — solicitante de prueba válido.
func ValidSolicitante() valueobject.Solicitante {
	s, _ := valueobject.NewSolicitante("Cliente Test", "+5491100000000")
	return s
}

// AbiertoTicket — ticket recién creado (estado abierto).
func AbiertoTicket() *model.Ticket {
	t, _ := model.NewTicket(
		uuid.New(),
		valueobject.ChannelWhatsApp,
		ValidSolicitante(),
		"No me funciona la app",
		valueobject.PriorityMedia,
	)
	return t
}

// AsignadoTicket — ticket asignado a un operador.
func AsignadoTicket(operador uuid.UUID) *model.Ticket {
	t := AbiertoTicket()
	_ = t.Asignar(operador)
	return t
}

// EnCursoTicket — ticket asignado y tomado.
func EnCursoTicket() *model.Ticket {
	t := AsignadoTicket(uuid.New())
	_ = t.Tomar()
	return t
}
