package usecase

import (
	"context"

	"github.com/google/uuid"

	"support-service/src/support/application/dto"
	"support-service/src/support/application/port"
	"support-service/src/support/domain/repository"
)

// AssignTicket — asigna un ticket a un operador (abierto → asignado).
type AssignTicket struct {
	repo   repository.TicketRepository
	events port.EventPublisher
}

func NewAssignTicket(repo repository.TicketRepository, events port.EventPublisher) *AssignTicket {
	return &AssignTicket{repo: repo, events: events}
}

func (uc *AssignTicket) Execute(ctx context.Context, ticketID, operadorID uuid.UUID) (dto.TicketResponse, error) {
	t, err := uc.repo.FindByID(ctx, ticketID)
	if err != nil {
		return dto.TicketResponse{}, err
	}
	if err := t.Asignar(operadorID); err != nil {
		return dto.TicketResponse{}, err
	}
	if err := uc.repo.Save(ctx, t); err != nil {
		return dto.TicketResponse{}, err
	}
	uc.events.Publish(ctx, t.PullEvents()...)
	return dto.FromTicket(t), nil
}
