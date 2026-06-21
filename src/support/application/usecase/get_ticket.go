package usecase

import (
	"context"

	"github.com/google/uuid"

	"support-service/src/support/application/dto"
	"support-service/src/support/domain/repository"
)

// GetTicket — consulta un ticket por id (acotado al tenant por RLS).
type GetTicket struct {
	repo repository.TicketRepository
}

func NewGetTicket(repo repository.TicketRepository) *GetTicket {
	return &GetTicket{repo: repo}
}

func (uc *GetTicket) Execute(ctx context.Context, ticketID uuid.UUID) (dto.TicketResponse, error) {
	t, err := uc.repo.FindByID(ctx, ticketID)
	if err != nil {
		return dto.TicketResponse{}, err
	}
	return dto.FromTicket(t), nil
}
