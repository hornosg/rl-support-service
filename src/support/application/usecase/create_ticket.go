package usecase

import (
	"context"

	"github.com/google/uuid"

	"support-service/src/support/application/dto"
	"support-service/src/support/application/port"
	"support-service/src/support/domain/model"
	"support-service/src/support/domain/repository"
	"support-service/src/support/domain/valueobject"
)

// CreateTicket — alta de un ticket (origen: canal). El tenant viene de la sesión.
type CreateTicket struct {
	repo   repository.TicketRepository
	events port.EventPublisher
}

func NewCreateTicket(repo repository.TicketRepository, events port.EventPublisher) *CreateTicket {
	return &CreateTicket{repo: repo, events: events}
}

func (uc *CreateTicket) Execute(ctx context.Context, tenantID uuid.UUID, req dto.CreateTicketRequest) (dto.TicketResponse, error) {
	canal, err := valueobject.NewChannel(req.Canal)
	if err != nil {
		return dto.TicketResponse{}, err
	}
	sol, err := valueobject.NewSolicitante(req.SolicitanteNombre, req.SolicitanteTelefono)
	if err != nil {
		return dto.TicketResponse{}, err
	}
	prio, err := valueobject.NewPriority(req.Prioridad)
	if err != nil {
		return dto.TicketResponse{}, err
	}

	t, err := model.NewTicket(tenantID, canal, sol, req.Asunto, prio)
	if err != nil {
		return dto.TicketResponse{}, err
	}

	if err := uc.repo.Save(ctx, t); err != nil {
		return dto.TicketResponse{}, err
	}
	uc.events.Publish(ctx, t.PullEvents()...)
	return dto.FromTicket(t), nil
}
