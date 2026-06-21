// Package dto — request/response de la capa de aplicación. La validación de borde vive
// en las binding tags; el dominio nunca recibe datos sin validar.
package dto

import (
	"time"

	"support-service/src/support/domain/model"
)

// CreateTicketRequest — alta de ticket desde el canal. El tenant NO viaja en el body:
// sale de la sesión de tenant (X-Tenant-ID validado contra el JWT).
type CreateTicketRequest struct {
	Canal               string `json:"canal" binding:"required"`
	SolicitanteNombre   string `json:"solicitante_nombre" binding:"required"`
	SolicitanteTelefono string `json:"solicitante_telefono" binding:"required"`
	Asunto              string `json:"asunto" binding:"required"`
	Prioridad           string `json:"prioridad"` // opcional → media
}

// AssignTicketRequest — asignación a un operador.
type AssignTicketRequest struct {
	OperadorID string `json:"operador_id" binding:"required,uuid"`
}

// TransitionTicketRequest — transición de estado por acción.
type TransitionTicketRequest struct {
	Accion string `json:"accion" binding:"required,oneof=tomar resolver cerrar"`
}

// ListTicketsQuery — filtros de listado (query params).
type ListTicketsQuery struct {
	Estado    string
	AsignadoA string
	Limit     int
	Offset    int
}

// TicketResponse — proyección de lectura del ticket.
type TicketResponse struct {
	ID                  string    `json:"id"`
	TenantID            string    `json:"tenant_id"`
	Canal               string    `json:"canal"`
	SolicitanteNombre   string    `json:"solicitante_nombre"`
	SolicitanteTelefono string    `json:"solicitante_telefono"`
	Asunto              string    `json:"asunto"`
	Prioridad           string    `json:"prioridad"`
	Estado              string    `json:"estado"`
	AsignadoA           *string   `json:"asignado_a"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// FromTicket mapea el agregado de dominio a su proyección de lectura.
func FromTicket(t *model.Ticket) TicketResponse {
	var asignado *string
	if t.AsignadoA() != nil {
		s := t.AsignadoA().String()
		asignado = &s
	}
	return TicketResponse{
		ID:                  t.ID().String(),
		TenantID:            t.TenantID().String(),
		Canal:               t.Canal().String(),
		SolicitanteNombre:   t.Solicitante().Nombre(),
		SolicitanteTelefono: t.Solicitante().Telefono(),
		Asunto:              t.Asunto(),
		Prioridad:           t.Prioridad().String(),
		Estado:              t.Estado().String(),
		AsignadoA:           asignado,
		CreatedAt:           t.CreatedAt(),
		UpdatedAt:           t.UpdatedAt(),
	}
}
