package usecase

import (
	"context"

	"github.com/google/uuid"

	"support-service/src/support/application/dto"
	"support-service/src/support/domain/model"
	"support-service/src/support/domain/repository"
)

// ListTickets — lista tickets del tenant con filtros opcionales (estado, asignado).
type ListTickets struct {
	repo repository.TicketRepository
}

func NewListTickets(repo repository.TicketRepository) *ListTickets {
	return &ListTickets{repo: repo}
}

func (uc *ListTickets) Execute(ctx context.Context, q dto.ListTicketsQuery) ([]dto.TicketResponse, error) {
	crit := repository.Criteria{
		Limit:  normalizeLimit(q.Limit),
		Offset: q.Offset,
	}
	if q.Estado != "" {
		st := model.Status(q.Estado)
		if !st.Valid() {
			return nil, ErrFiltroEstadoInvalido
		}
		crit.Estado = &st
	}
	if q.AsignadoA != "" {
		id, err := uuid.Parse(q.AsignadoA)
		if err != nil {
			return nil, err
		}
		crit.AsignadoA = &id
	}

	tickets, err := uc.repo.Find(ctx, crit)
	if err != nil {
		return nil, err
	}
	out := make([]dto.TicketResponse, 0, len(tickets))
	for _, t := range tickets {
		out = append(out, dto.FromTicket(t))
	}
	return out, nil
}
