// Package repository — puertos de persistencia del dominio (solo interfaces).
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"support-service/src/support/domain/model"
)

// ErrNotFound — el ticket no existe (o no es visible bajo el tenant de la sesión, por RLS).
var ErrNotFound = errors.New("repository: ticket no encontrado")

// Criteria — filtros para listar tickets. El aislamiento por tenant lo garantiza RLS,
// no estos filtros (defensa en profundidad).
type Criteria struct {
	Estado    *model.Status
	AsignadoA *uuid.UUID
	Limit     int
	Offset    int
}

// TicketRepository — port de persistencia del agregado Ticket.
type TicketRepository interface {
	Save(ctx context.Context, t *model.Ticket) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Ticket, error)
	Find(ctx context.Context, c Criteria) ([]*model.Ticket, error)
	// AnonimizarSolicitante borra la PII (tombstone) de todos los tickets del tenant
	// cuyo solicitante tenga ese teléfono. Devuelve cuántos tickets afectó.
	AnonimizarSolicitante(ctx context.Context, telefono string) (int, error)
}
