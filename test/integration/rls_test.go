//go:build integration

// Test de aislamiento RLS contra la DB real del lab (E03).
// Correr con el lab arriba y la DB provisionada:
//
//	go test -tags integration ./test/integration/ -v
//
// Usa el rol de app (NOBYPASSRLS): el aislamiento lo impone la base, no el código.
package integration

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"support-service/src/support/application/dto"
	"support-service/src/support/application/usecase"
	"support-service/src/support/domain/model"
	"support-service/src/support/domain/repository"
	"support-service/src/support/domain/valueobject"
	"support-service/src/support/infrastructure/persistence"
	"support-service/test/fake"
)

func dsn() string {
	if v := os.Getenv("TEST_DATABASE_DSN"); v != "" {
		return v
	}
	// Sin password embebida: viene de TEST_DATABASE_DSN o de PGPASSWORD/.pgpass.
	// Sin credenciales, el Ping falla y los tests se saltean.
	return "host=localhost port=5432 user=support_service_app dbname=support_service sslmode=disable"
}

func openDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("postgres", dsn())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("sin DB accesible (%v) — saltando integración", err)
	}
	return db
}

// tenantConn fija una conexión y setea app.tenant_id en ella (igual que el middleware real).
func tenantConn(t *testing.T, db *sql.DB, tenant uuid.UUID) *sql.Conn {
	t.Helper()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("conn: %v", err)
	}
	if _, err := conn.ExecContext(ctx, "SELECT set_config('app.tenant_id', $1, false)", tenant.String()); err != nil {
		t.Fatalf("set app.tenant_id: %v", err)
	}
	return conn
}

func nuevoTicket(t *testing.T, tenant uuid.UUID) *model.Ticket {
	t.Helper()
	sol, _ := valueobject.NewSolicitante("Cliente Test", "+5491100000000")
	tk, err := model.NewTicket(tenant, valueobject.ChannelWhatsApp, sol, "no me anda", valueobject.PriorityMedia)
	if err != nil {
		t.Fatalf("new ticket: %v", err)
	}
	return tk
}

func TestRLS_AislamientoEntreTenants(t *testing.T) {
	db := openDB(t)
	defer db.Close()
	ctx := context.Background()
	tenantA, tenantB := uuid.New(), uuid.New()

	connA := tenantConn(t, db, tenantA)
	defer connA.Close()
	repoA := persistence.NewPgTicketRepository(connA)

	tk := nuevoTicket(t, tenantA)
	if err := repoA.Save(ctx, tk); err != nil {
		t.Fatalf("save bajo tenant A: %v", err)
	}
	t.Cleanup(func() { _, _ = connA.ExecContext(ctx, "DELETE FROM tickets WHERE id=$1", tk.ID()) })

	// A ve su ticket.
	if _, err := repoA.FindByID(ctx, tk.ID()); err != nil {
		t.Fatalf("tenant A debe ver su ticket: %v", err)
	}

	// B NO ve el ticket de A (lo filtra la base, no el código).
	connB := tenantConn(t, db, tenantB)
	defer connB.Close()
	repoB := persistence.NewPgTicketRepository(connB)

	if _, err := repoB.FindByID(ctx, tk.ID()); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("tenant B no debe ver el ticket de A — got err=%v", err)
	}
	listB, err := repoB.Find(ctx, repository.Criteria{Limit: 50})
	if err != nil {
		t.Fatalf("list B: %v", err)
	}
	for _, x := range listB {
		if x.ID() == tk.ID() {
			t.Fatal("el ticket de A apareció en el listado de B")
		}
	}

	// A lo encuentra en su listado.
	listA, err := repoA.Find(ctx, repository.Criteria{Limit: 50})
	if err != nil {
		t.Fatalf("list A: %v", err)
	}
	found := false
	for _, x := range listA {
		if x.ID() == tk.ID() {
			found = true
		}
	}
	if !found {
		t.Fatal("tenant A debe ver su ticket en el listado")
	}
}

func TestLifecycle_PersisteTransiciones(t *testing.T) {
	db := openDB(t)
	defer db.Close()
	ctx := context.Background()
	tenant := uuid.New()

	conn := tenantConn(t, db, tenant)
	defer conn.Close()
	repo := persistence.NewPgTicketRepository(conn)
	pub := &fake.Publisher{}

	resp, err := usecase.NewCreateTicket(repo, pub).Execute(ctx, tenant, dto.CreateTicketRequest{
		Canal:               "whatsapp",
		SolicitanteNombre:   "Ana",
		SolicitanteTelefono: "+5491100000000",
		Asunto:              "no me llega el código",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	id := uuid.MustParse(resp.ID)
	t.Cleanup(func() { _, _ = conn.ExecContext(ctx, "DELETE FROM tickets WHERE id=$1", id) })

	op := uuid.New()
	if _, err := usecase.NewAssignTicket(repo, pub).Execute(ctx, id, op); err != nil {
		t.Fatalf("asignar: %v", err)
	}
	for _, accion := range []string{usecase.AccionTomar, usecase.AccionResolver, usecase.AccionCerrar} {
		if _, err := usecase.NewTransitionTicket(repo, pub).Execute(ctx, id, accion); err != nil {
			t.Fatalf("transicionar %s: %v", accion, err)
		}
	}

	// Releer del repo: el estado final y el operador deben haber persistido (upsert/update).
	got, err := usecase.NewGetTicket(repo).Execute(ctx, id)
	if err != nil {
		t.Fatalf("get final: %v", err)
	}
	if got.Estado != "cerrado" {
		t.Fatalf("estado persistido = %s, esperaba cerrado", got.Estado)
	}
	if got.AsignadoA == nil || *got.AsignadoA != op.String() {
		t.Fatal("asignado_a no persistió")
	}
}

func TestBorrarPII_AnonimizaSoloDelTenant(t *testing.T) {
	db := openDB(t)
	defer db.Close()
	ctx := context.Background()
	tenantA, tenantB := uuid.New(), uuid.New()
	tel := "PII-" + uuid.NewString() // teléfono único para no colisionar con otros datos

	connA := tenantConn(t, db, tenantA)
	defer connA.Close()
	repoA := persistence.NewPgTicketRepository(connA)

	var idsA []uuid.UUID
	for i := 0; i < 2; i++ {
		sol, _ := valueobject.NewSolicitante("Ana", tel)
		tk, _ := model.NewTicket(tenantA, valueobject.ChannelWhatsApp, sol, "x", valueobject.PriorityMedia)
		if err := repoA.Save(ctx, tk); err != nil {
			t.Fatalf("save A: %v", err)
		}
		idsA = append(idsA, tk.ID())
	}

	connB := tenantConn(t, db, tenantB)
	defer connB.Close()
	repoB := persistence.NewPgTicketRepository(connB)
	solB, _ := valueobject.NewSolicitante("Otro", tel)
	tkB, _ := model.NewTicket(tenantB, valueobject.ChannelWhatsApp, solB, "y", valueobject.PriorityMedia)
	if err := repoB.Save(ctx, tkB); err != nil {
		t.Fatalf("save B: %v", err)
	}
	t.Cleanup(func() {
		for _, id := range idsA {
			_, _ = connA.ExecContext(ctx, "DELETE FROM tickets WHERE id=$1", id)
		}
		_, _ = connB.ExecContext(ctx, "DELETE FROM tickets WHERE id=$1", tkB.ID())
	})

	// Borrar PII bajo tenant A.
	n, err := repoA.AnonimizarSolicitante(ctx, tel)
	if err != nil {
		t.Fatalf("anonimizar: %v", err)
	}
	if n != 2 {
		t.Fatalf("esperaba 2 tickets afectados en tenant A, got %d", n)
	}

	// A: sus tickets quedaron anónimos.
	for _, id := range idsA {
		got, err := repoA.FindByID(ctx, id)
		if err != nil {
			t.Fatalf("find A: %v", err)
		}
		if !got.Solicitante().EsAnonimo() {
			t.Fatal("el ticket de A debió quedar anonimizado")
		}
	}

	// B: su ticket con el MISMO teléfono NO fue tocado (RLS acotó el UPDATE al tenant A).
	gotB, err := repoB.FindByID(ctx, tkB.ID())
	if err != nil {
		t.Fatalf("find B: %v", err)
	}
	if gotB.Solicitante().EsAnonimo() {
		t.Fatal("RLS: el ticket de tenant B NO debió anonimizarse")
	}
}

func TestRLS_InsertCrossTenantRechazado(t *testing.T) {
	db := openDB(t)
	defer db.Close()
	ctx := context.Background()
	tenantA, tenantB := uuid.New(), uuid.New()

	connA := tenantConn(t, db, tenantA)
	defer connA.Close()
	repoA := persistence.NewPgTicketRepository(connA)

	// Bajo sesión A, intentar persistir un ticket de tenant B → WITH CHECK lo rechaza.
	tkB := nuevoTicket(t, tenantB)
	if err := repoA.Save(ctx, tkB); err == nil {
		_, _ = connA.ExecContext(ctx, "DELETE FROM tickets WHERE id=$1", tkB.ID())
		t.Fatal("RLS (WITH CHECK) debió rechazar el insert cross-tenant")
	}
}
