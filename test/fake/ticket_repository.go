// Package fake — dobles de prueba in-memory (sin DB) para tests de use cases.
package fake

import (
	"context"

	"github.com/google/uuid"

	"support-service/src/support/domain/event"
	"support-service/src/support/domain/model"
	"support-service/src/support/domain/repository"
)

// TicketRepo — repositorio in-memory. SaveErr permite forzar fallos de persistencia.
type TicketRepo struct {
	store   map[uuid.UUID]*model.Ticket
	SaveErr error
}

func NewTicketRepo() *TicketRepo {
	return &TicketRepo{store: make(map[uuid.UUID]*model.Ticket)}
}

var _ repository.TicketRepository = (*TicketRepo)(nil)

func (r *TicketRepo) Save(_ context.Context, t *model.Ticket) error {
	if r.SaveErr != nil {
		return r.SaveErr
	}
	r.store[t.ID()] = t
	return nil
}

func (r *TicketRepo) FindByID(_ context.Context, id uuid.UUID) (*model.Ticket, error) {
	t, ok := r.store[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return t, nil
}

func (r *TicketRepo) Find(_ context.Context, c repository.Criteria) ([]*model.Ticket, error) {
	out := make([]*model.Ticket, 0)
	for _, t := range r.store {
		if c.Estado != nil && t.Estado() != *c.Estado {
			continue
		}
		if c.AsignadoA != nil && (t.AsignadoA() == nil || *t.AsignadoA() != *c.AsignadoA) {
			continue
		}
		out = append(out, t)
	}
	return out, nil
}

// Publisher — EventPublisher que solo cuenta los eventos publicados.
type Publisher struct {
	Count int
}

func (p *Publisher) Publish(_ context.Context, events ...event.DomainEvent) {
	p.Count += len(events)
}
