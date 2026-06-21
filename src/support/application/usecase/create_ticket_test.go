package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"support-service/src/support/application/dto"
	"support-service/src/support/application/usecase"
	"support-service/src/support/domain/valueobject"
	"support-service/test/fake"
	"support-service/test/mother"
)

func TestCreateTicket_HappyPath(t *testing.T) {
	repo := fake.NewTicketRepo()
	pub := &fake.Publisher{}
	uc := usecase.NewCreateTicket(repo, pub)

	resp, err := uc.Execute(context.Background(), uuid.New(), dto.CreateTicketRequest{
		Canal:               "whatsapp",
		SolicitanteNombre:   "Ana",
		SolicitanteTelefono: "+5491100000000",
		Asunto:              "No me llega el código",
		Prioridad:           "", // → media por default
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if resp.Estado != "abierto" {
		t.Fatalf("estado inicial debe ser abierto, got %s", resp.Estado)
	}
	if resp.Prioridad != "media" {
		t.Fatalf("prioridad default debe ser media, got %s", resp.Prioridad)
	}
	if pub.Count == 0 {
		t.Fatal("esperaba al menos un evento publicado")
	}
}

func TestCreateTicket_CanalInvalido(t *testing.T) {
	uc := usecase.NewCreateTicket(fake.NewTicketRepo(), &fake.Publisher{})
	_, err := uc.Execute(context.Background(), uuid.New(), dto.CreateTicketRequest{
		Canal:               "telegram",
		SolicitanteNombre:   "Ana",
		SolicitanteTelefono: "+549",
		Asunto:              "x",
	})
	if !errors.Is(err, valueobject.ErrCanalInvalido) {
		t.Fatalf("esperaba ErrCanalInvalido, got %v", err)
	}
}

func TestTransitionTicket_TomarDesdeAsignado(t *testing.T) {
	repo := fake.NewTicketRepo()
	pub := &fake.Publisher{}
	tk := mother.AsignadoTicket(uuid.New())
	if err := repo.Save(context.Background(), tk); err != nil {
		t.Fatalf("seed: %v", err)
	}

	resp, err := usecase.NewTransitionTicket(repo, pub).Execute(context.Background(), tk.ID(), usecase.AccionTomar)
	if err != nil {
		t.Fatalf("transicionar: %v", err)
	}
	if resp.Estado != "en_curso" {
		t.Fatalf("esperaba en_curso, got %s", resp.Estado)
	}
}

func TestBorrarPIISolicitante(t *testing.T) {
	repo := fake.NewTicketRepo()
	ctx := context.Background()
	tenant := uuid.New()

	// Dos tickets del mismo solicitante (mismo teléfono) + uno de otro.
	uc := usecase.NewCreateTicket(repo, &fake.Publisher{})
	for i := 0; i < 2; i++ {
		if _, err := uc.Execute(ctx, tenant, dto.CreateTicketRequest{
			Canal: "whatsapp", SolicitanteNombre: "Ana", SolicitanteTelefono: "+5491111111111", Asunto: "x",
		}); err != nil {
			t.Fatalf("create: %v", err)
		}
	}
	if _, err := uc.Execute(ctx, tenant, dto.CreateTicketRequest{
		Canal: "whatsapp", SolicitanteNombre: "Beto", SolicitanteTelefono: "+5492222222222", Asunto: "y",
	}); err != nil {
		t.Fatalf("create: %v", err)
	}

	resp, err := usecase.NewBorrarPIISolicitante(repo).Execute(ctx, dto.BorrarPIIRequest{Telefono: "+5491111111111"})
	if err != nil {
		t.Fatalf("borrar pii: %v", err)
	}
	if resp.Anonimizados != 2 {
		t.Fatalf("esperaba 2 anonimizados, got %d", resp.Anonimizados)
	}
}

func TestGetTicket_NotFound(t *testing.T) {
	_, err := usecase.NewGetTicket(fake.NewTicketRepo()).Execute(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("esperaba error not found")
	}
}
