package model

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"support-service/src/support/domain/event"
	"support-service/src/support/domain/valueobject"
)

// Ticket — aggregate root del dominio de soporte (G-01). Estado mutable solo a través
// de sus métodos de comportamiento; sin setters. Multi-tenant: el tenant es parte de
// la identidad del agregado y la red de seguridad RLS lo refuerza en la DB.
type Ticket struct {
	id          uuid.UUID
	tenantID    uuid.UUID
	canal       valueobject.Channel
	solicitante valueobject.Solicitante
	asunto      string
	prioridad   valueobject.Priority
	estado      Status
	asignadoA   *uuid.UUID // operador; nil hasta asignar
	createdAt   time.Time
	updatedAt   time.Time
	events      []event.DomainEvent
}

// NewTicket crea un ticket en estado abierto (origen: canal). Valida en la creación.
func NewTicket(
	tenantID uuid.UUID,
	canal valueobject.Channel,
	solicitante valueobject.Solicitante,
	asunto string,
	prioridad valueobject.Priority,
) (*Ticket, error) {
	asunto = strings.TrimSpace(asunto)
	if asunto == "" {
		return nil, ErrAsuntoVacio
	}
	now := time.Now().UTC()
	id := uuid.New()
	t := &Ticket{
		id:          id,
		tenantID:    tenantID,
		canal:       canal,
		solicitante: solicitante,
		asunto:      asunto,
		prioridad:   prioridad,
		estado:      StatusAbierto,
		createdAt:   now,
		updatedAt:   now,
	}
	t.record(event.NewTicketCreado(id, tenantID))
	return t, nil
}

// Rehydrate reconstruye un Ticket desde persistencia: sin validación ni eventos.
// Lo usa el repositorio al mapear filas → dominio.
func Rehydrate(
	id, tenantID uuid.UUID,
	canal valueobject.Channel,
	solicitante valueobject.Solicitante,
	asunto string,
	prioridad valueobject.Priority,
	estado Status,
	asignadoA *uuid.UUID,
	createdAt, updatedAt time.Time,
) *Ticket {
	return &Ticket{
		id:          id,
		tenantID:    tenantID,
		canal:       canal,
		solicitante: solicitante,
		asunto:      asunto,
		prioridad:   prioridad,
		estado:      estado,
		asignadoA:   asignadoA,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// ── Comportamiento (máquina de estados lineal) ──

// Asignar lleva el ticket de abierto → asignado, fijando el operador responsable.
func (t *Ticket) Asignar(operador uuid.UUID) error {
	if t.estado != StatusAbierto {
		return ErrTransicionInvalida
	}
	t.asignadoA = &operador
	t.transition(StatusAsignado)
	t.record(event.NewTicketAsignado(t.id, operador))
	return nil
}

// Tomar lleva el ticket de asignado → en_curso (el operador empieza a trabajarlo).
func (t *Ticket) Tomar() error {
	if t.estado != StatusAsignado {
		return ErrTransicionInvalida
	}
	if t.asignadoA == nil {
		return ErrSinOperador
	}
	t.transition(StatusEnCurso)
	return nil
}

// Resolver lleva el ticket de en_curso → resuelto.
func (t *Ticket) Resolver() error {
	if t.estado != StatusEnCurso {
		return ErrTransicionInvalida
	}
	t.transition(StatusResuelto)
	return nil
}

// Cerrar lleva el ticket de resuelto → cerrado (estado terminal del slice).
func (t *Ticket) Cerrar() error {
	if t.estado != StatusResuelto {
		return ErrTransicionInvalida
	}
	t.transition(StatusCerrado)
	return nil
}

func (t *Ticket) transition(to Status) {
	from := t.estado
	t.estado = to
	t.updatedAt = time.Now().UTC()
	t.record(event.NewTicketTransicionado(t.id, from.String(), to.String()))
}

func (t *Ticket) record(e event.DomainEvent) { t.events = append(t.events, e) }

// PullEvents devuelve y limpia los eventos acumulados (el use case los publica).
func (t *Ticket) PullEvents() []event.DomainEvent {
	ev := t.events
	t.events = nil
	return ev
}

// ── Getters (sin setters) ──

func (t *Ticket) ID() uuid.UUID                       { return t.id }
func (t *Ticket) TenantID() uuid.UUID                 { return t.tenantID }
func (t *Ticket) Canal() valueobject.Channel          { return t.canal }
func (t *Ticket) Solicitante() valueobject.Solicitante { return t.solicitante }
func (t *Ticket) Asunto() string                      { return t.asunto }
func (t *Ticket) Prioridad() valueobject.Priority     { return t.prioridad }
func (t *Ticket) Estado() Status                      { return t.estado }
func (t *Ticket) AsignadoA() *uuid.UUID               { return t.asignadoA }
func (t *Ticket) CreatedAt() time.Time                { return t.createdAt }
func (t *Ticket) UpdatedAt() time.Time                { return t.updatedAt }
