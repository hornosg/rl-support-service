// Package port — driven ports de la capa de aplicación (servicios externos al dominio).
package port

import (
	"context"

	"support-service/src/support/domain/event"
)

// EventPublisher — publica eventos de dominio. Implementación in-process en infrastructure.
type EventPublisher interface {
	Publish(ctx context.Context, events ...event.DomainEvent)
}
