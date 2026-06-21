// Package persistence — adapters de persistencia (database/sql + lib/pq, raw SQL).
package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // driver postgres

	"support-service/src/support/domain/model"
	"support-service/src/support/domain/repository"
	"support-service/src/support/domain/valueobject"
)

// Executor — mínimo común entre *sql.DB y *sql.Conn. El repo recibe la conexión
// FIJADA del request (con app.tenant_id seteado), no el pool, para que RLS aplique.
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type scanner interface{ Scan(dest ...any) error }

const selectColumns = `
SELECT id, tenant_id, canal, solicitante_nombre, solicitante_telefono,
       asunto, prioridad, estado, asignado_a, created_at, updated_at
FROM tickets`

// PgTicketRepository implementa repository.TicketRepository.
type PgTicketRepository struct {
	exec Executor
}

func NewPgTicketRepository(exec Executor) *PgTicketRepository {
	return &PgTicketRepository{exec: exec}
}

var _ repository.TicketRepository = (*PgTicketRepository)(nil)

// Save hace upsert: inserta el ticket nuevo o actualiza los campos mutables.
// tenant_id es inmutable; RLS (WITH CHECK) rechaza filas de otro tenant.
func (r *PgTicketRepository) Save(ctx context.Context, t *model.Ticket) error {
	var asignado any
	if t.AsignadoA() != nil {
		asignado = *t.AsignadoA()
	}
	_, err := r.exec.ExecContext(ctx, `
INSERT INTO tickets (id, tenant_id, canal, solicitante_nombre, solicitante_telefono,
                     asunto, prioridad, estado, asignado_a, created_at, updated_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
ON CONFLICT (id) DO UPDATE SET
    estado     = EXCLUDED.estado,
    asignado_a = EXCLUDED.asignado_a,
    prioridad  = EXCLUDED.prioridad,
    updated_at = EXCLUDED.updated_at`,
		t.ID(), t.TenantID(), t.Canal().String(), t.Solicitante().Nombre(), t.Solicitante().Telefono(),
		t.Asunto(), t.Prioridad().String(), t.Estado().String(), asignado, t.CreatedAt(), t.UpdatedAt(),
	)
	return err
}

func (r *PgTicketRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Ticket, error) {
	row := r.exec.QueryRowContext(ctx, selectColumns+` WHERE id = $1`, id)
	t, err := scanTicket(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	return t, err
}

func (r *PgTicketRepository) Find(ctx context.Context, c repository.Criteria) ([]*model.Ticket, error) {
	q := selectColumns + ` WHERE 1=1`
	args := []any{}
	n := 0
	if c.Estado != nil {
		n++
		q += fmt.Sprintf(` AND estado = $%d`, n)
		args = append(args, c.Estado.String())
	}
	if c.AsignadoA != nil {
		n++
		q += fmt.Sprintf(` AND asignado_a = $%d`, n)
		args = append(args, *c.AsignadoA)
	}
	q += ` ORDER BY created_at DESC`
	n++
	q += fmt.Sprintf(` LIMIT $%d`, n)
	args = append(args, c.Limit)
	n++
	q += fmt.Sprintf(` OFFSET $%d`, n)
	args = append(args, c.Offset)

	rows, err := r.exec.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*model.Ticket, 0)
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func scanTicket(s scanner) (*model.Ticket, error) {
	var (
		id, tenantID                                  uuid.UUID
		canal, nombre, telefono, asunto, prio, estado string
		asignado                                      uuid.NullUUID
		createdAt, updatedAt                          time.Time
	)
	if err := s.Scan(&id, &tenantID, &canal, &nombre, &telefono, &asunto, &prio, &estado, &asignado, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	sol, err := valueobject.NewSolicitante(nombre, telefono)
	if err != nil {
		return nil, err
	}
	ch, err := valueobject.NewChannel(canal)
	if err != nil {
		return nil, err
	}
	p, err := valueobject.NewPriority(prio)
	if err != nil {
		return nil, err
	}
	var asignadoPtr *uuid.UUID
	if asignado.Valid {
		v := asignado.UUID
		asignadoPtr = &v
	}
	return model.Rehydrate(id, tenantID, ch, sol, asunto, p, model.Status(estado), asignadoPtr, createdAt, updatedAt), nil
}
