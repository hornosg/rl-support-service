// Package messaging — implementaciones de publicación de eventos.
package messaging

import (
	"context"

	"go.uber.org/zap"

	"support-service/src/support/application/port"
	"support-service/src/support/domain/event"
)

// LogPublisher — publicación in-process: por ahora solo loggea los eventos de dominio.
// No loggea PII (RULE-03): el evento solo lleva ids y nombres de evento.
type LogPublisher struct {
	log *zap.Logger
}

func NewLogPublisher(log *zap.Logger) *LogPublisher {
	return &LogPublisher{log: log}
}

var _ port.EventPublisher = (*LogPublisher)(nil)

func (p *LogPublisher) Publish(_ context.Context, events ...event.DomainEvent) {
	for _, e := range events {
		p.log.Info("domain_event",
			zap.String("event", e.Name()),
			zap.Time("occurred_at", e.OccurredAt()),
		)
	}
}
