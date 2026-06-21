// Package handler — entry points HTTP (Gin). Traducen HTTP ↔ DTO e invocan use cases.
// Gin nunca cruza a application/domain.
package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"support-service/src/shared/database"
	"support-service/src/support/application/dto"
	"support-service/src/support/application/usecase"
	"support-service/src/support/domain/model"
	"support-service/src/support/domain/repository"
	"support-service/src/support/domain/valueobject"
	"support-service/src/support/infrastructure/http/problem"
	"support-service/src/support/infrastructure/messaging"
	"support-service/src/support/infrastructure/persistence"
)

// TicketHandler expone el contrato de tickets (RULE-06). El repo se construye por request
// con la conexión FIJADA (RLS) — desvío justificado del wiring en main del skill hexagonal-go.
type TicketHandler struct {
	log *zap.Logger
}

func NewTicketHandler(log *zap.Logger) *TicketHandler {
	return &TicketHandler{log: log}
}

// Register monta las rutas del contrato bajo el grupo dado (ej: /api/v1).
func (h *TicketHandler) Register(rg *gin.RouterGroup) {
	rg.POST("/tickets", h.Create)
	rg.GET("/tickets", h.List)
	rg.GET("/tickets/:id", h.Get)
	rg.POST("/tickets/:id/asignar", h.Assign)
	rg.POST("/tickets/:id/transicionar", h.Transition)
	rg.POST("/solicitantes/borrar-pii", h.BorrarPII)
}

// BorrarPII anonimiza la PII del solicitante (Ley 25.326) en los tickets del tenant.
func (h *TicketHandler) BorrarPII(c *gin.Context) {
	repo, _, ok := h.deps(c)
	if !ok {
		return
	}
	var req dto.BorrarPIIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		problem.Write(c, http.StatusBadRequest, "body inválido", err.Error())
		return
	}
	resp, err := usecase.NewBorrarPIISolicitante(repo).Execute(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *TicketHandler) Create(c *gin.Context) {
	tenant, ok := tenantID(c)
	if !ok {
		return
	}
	repo, pub, ok := h.deps(c)
	if !ok {
		return
	}
	var req dto.CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		problem.Write(c, http.StatusBadRequest, "body inválido", err.Error())
		return
	}
	resp, err := usecase.NewCreateTicket(repo, pub).Execute(c.Request.Context(), tenant, req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *TicketHandler) Get(c *gin.Context) {
	repo, _, ok := h.deps(c)
	if !ok {
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	resp, err := usecase.NewGetTicket(repo).Execute(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *TicketHandler) List(c *gin.Context) {
	repo, _, ok := h.deps(c)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	q := dto.ListTicketsQuery{
		Estado:    c.Query("estado"),
		AsignadoA: c.Query("asignado_a"),
		Limit:     limit,
		Offset:    offset,
	}
	resp, err := usecase.NewListTickets(repo).Execute(c.Request.Context(), q)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *TicketHandler) Assign(c *gin.Context) {
	repo, pub, ok := h.deps(c)
	if !ok {
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	var req dto.AssignTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		problem.Write(c, http.StatusBadRequest, "body inválido", err.Error())
		return
	}
	operador, err := uuid.Parse(req.OperadorID)
	if err != nil {
		problem.Write(c, http.StatusBadRequest, "operador inválido", err.Error())
		return
	}
	resp, err := usecase.NewAssignTicket(repo, pub).Execute(c.Request.Context(), id, operador)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *TicketHandler) Transition(c *gin.Context) {
	repo, pub, ok := h.deps(c)
	if !ok {
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	var req dto.TransitionTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		problem.Write(c, http.StatusBadRequest, "body inválido", err.Error())
		return
	}
	resp, err := usecase.NewTransitionTicket(repo, pub).Execute(c.Request.Context(), id, req.Accion)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ── helpers ──

// deps arma el repo (sobre la conexión fijada del request) y el publisher.
func (h *TicketHandler) deps(c *gin.Context) (repository.TicketRepository, *messaging.LogPublisher, bool) {
	conn := database.Conn(c)
	if conn == nil {
		problem.Write(c, http.StatusServiceUnavailable, "db no disponible", "el servicio no tiene conexión a la base")
		return nil, nil, false
	}
	return persistence.NewPgTicketRepository(conn), messaging.NewLogPublisher(h.log), true
}

func tenantID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.GetHeader("X-Tenant-ID"))
	if err != nil {
		problem.Write(c, http.StatusUnauthorized, "tenant requerido", "X-Tenant-ID ausente o inválido")
		return uuid.Nil, false
	}
	return id, true
}

func pathID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		problem.Write(c, http.StatusBadRequest, "id inválido", "el id del ticket no es un uuid")
		return uuid.Nil, false
	}
	return id, true
}

// writeError mapea errores de dominio/aplicación → Problem Details (RFC 7807).
func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		problem.Write(c, http.StatusNotFound, "ticket no encontrado", "")
	case errors.Is(err, model.ErrTransicionInvalida), errors.Is(err, model.ErrSinOperador):
		problem.Write(c, http.StatusConflict, "transición inválida", err.Error())
	case errors.Is(err, model.ErrAsuntoVacio),
		errors.Is(err, model.ErrAccionDesconocida),
		errors.Is(err, valueobject.ErrCanalInvalido),
		errors.Is(err, valueobject.ErrPrioridadInvalida),
		errors.Is(err, valueobject.ErrNombreVacio),
		errors.Is(err, valueobject.ErrTelefonoVacio),
		errors.Is(err, usecase.ErrFiltroEstadoInvalido):
		problem.Write(c, http.StatusUnprocessableEntity, "datos inválidos", err.Error())
	default:
		problem.Write(c, http.StatusInternalServerError, "error interno", "")
	}
}
