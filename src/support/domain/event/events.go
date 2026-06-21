// Package event — eventos de dominio del agregado Ticket. Por ahora se publican
// in-process (ver application/port.EventPublisher); a futuro pueden ir a un bus.
package event

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent — contrato común de todo evento de dominio.
type DomainEvent interface {
	Name() string
	OccurredAt() time.Time
}

type base struct{ at time.Time }

func (b base) OccurredAt() time.Time { return b.at }

// TicketCreado — un ticket nuevo entró al sistema (origen: canal).
type TicketCreado struct {
	base
	TicketID uuid.UUID
	TenantID uuid.UUID
}

func NewTicketCreado(ticketID, tenantID uuid.UUID) TicketCreado {
	return TicketCreado{base: base{at: time.Now().UTC()}, TicketID: ticketID, TenantID: tenantID}
}

func (TicketCreado) Name() string { return "ticket.creado" }

// TicketAsignado — un ticket fue asignado a un operador.
type TicketAsignado struct {
	base
	TicketID   uuid.UUID
	OperadorID uuid.UUID
}

func NewTicketAsignado(ticketID, operadorID uuid.UUID) TicketAsignado {
	return TicketAsignado{base: base{at: time.Now().UTC()}, TicketID: ticketID, OperadorID: operadorID}
}

func (TicketAsignado) Name() string { return "ticket.asignado" }

// TicketTransicionado — el ticket cambió de estado.
type TicketTransicionado struct {
	base
	TicketID uuid.UUID
	De       string
	A        string
}

func NewTicketTransicionado(ticketID uuid.UUID, de, a string) TicketTransicionado {
	return TicketTransicionado{base: base{at: time.Now().UTC()}, TicketID: ticketID, De: de, A: a}
}

func (TicketTransicionado) Name() string { return "ticket.transicionado" }
