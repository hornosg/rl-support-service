package usecase

import (
	"context"

	"github.com/google/uuid"

	"support-service/src/support/application/dto"
	"support-service/src/support/application/port"
	"support-service/src/support/domain/model"
	"support-service/src/support/domain/repository"
)

// Acciones de transición soportadas (validadas también por binding `oneof` en el DTO).
const (
	AccionTomar    = "tomar"
	AccionResolver = "resolver"
	AccionCerrar   = "cerrar"
)

// TransitionTicket — avanza el ticket por su máquina de estados según la acción.
type TransitionTicket struct {
	repo   repository.TicketRepository
	events port.EventPublisher
}

func NewTransitionTicket(repo repository.TicketRepository, events port.EventPublisher) *TransitionTicket {
	return &TransitionTicket{repo: repo, events: events}
}

func (uc *TransitionTicket) Execute(ctx context.Context, ticketID uuid.UUID, accion string) (dto.TicketResponse, error) {
	t, err := uc.repo.FindByID(ctx, ticketID)
	if err != nil {
		return dto.TicketResponse{}, err
	}
	if err := apply(t, accion); err != nil {
		return dto.TicketResponse{}, err
	}
	if err := uc.repo.Save(ctx, t); err != nil {
		return dto.TicketResponse{}, err
	}
	uc.events.Publish(ctx, t.PullEvents()...)
	return dto.FromTicket(t), nil
}

func apply(t *model.Ticket, accion string) error {
	switch accion {
	case AccionTomar:
		return t.Tomar()
	case AccionResolver:
		return t.Resolver()
	case AccionCerrar:
		return t.Cerrar()
	default:
		return model.ErrAccionDesconocida
	}
}
