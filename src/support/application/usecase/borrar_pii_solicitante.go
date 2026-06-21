package usecase

import (
	"context"

	"support-service/src/support/application/dto"
	"support-service/src/support/domain/repository"
)

// BorrarPIISolicitante — ejerce el derecho de supresión (Ley 25.326): anonimiza la PII
// del solicitante (por teléfono) en todos los tickets del tenant. El alcance al tenant
// lo garantiza RLS, no este use case.
type BorrarPIISolicitante struct {
	repo repository.TicketRepository
}

func NewBorrarPIISolicitante(repo repository.TicketRepository) *BorrarPIISolicitante {
	return &BorrarPIISolicitante{repo: repo}
}

func (uc *BorrarPIISolicitante) Execute(ctx context.Context, req dto.BorrarPIIRequest) (dto.BorrarPIIResponse, error) {
	n, err := uc.repo.AnonimizarSolicitante(ctx, req.Telefono)
	if err != nil {
		return dto.BorrarPIIResponse{}, err
	}
	return dto.BorrarPIIResponse{Anonimizados: n}, nil
}
