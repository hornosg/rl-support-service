package model_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"support-service/src/support/domain/model"
	"support-service/src/support/domain/valueobject"
	"support-service/test/mother"
)

func TestTicket_CicloDeVidaLineal(t *testing.T) {
	tk := mother.AbiertoTicket()
	if tk.Estado() != model.StatusAbierto {
		t.Fatalf("nuevo ticket debe estar abierto, got %s", tk.Estado())
	}

	op := uuid.New()
	if err := tk.Asignar(op); err != nil {
		t.Fatalf("asignar: %v", err)
	}
	if tk.Estado() != model.StatusAsignado {
		t.Fatalf("tras asignar debe estar asignado, got %s", tk.Estado())
	}
	if tk.AsignadoA() == nil || *tk.AsignadoA() != op {
		t.Fatal("asignadoA no quedó seteado al operador")
	}

	if err := tk.Tomar(); err != nil {
		t.Fatalf("tomar: %v", err)
	}
	if tk.Estado() != model.StatusEnCurso {
		t.Fatalf("tras tomar debe estar en_curso, got %s", tk.Estado())
	}

	if err := tk.Resolver(); err != nil {
		t.Fatalf("resolver: %v", err)
	}
	if tk.Estado() != model.StatusResuelto {
		t.Fatalf("tras resolver debe estar resuelto, got %s", tk.Estado())
	}

	if err := tk.Cerrar(); err != nil {
		t.Fatalf("cerrar: %v", err)
	}
	if tk.Estado() != model.StatusCerrado {
		t.Fatalf("tras cerrar debe estar cerrado, got %s", tk.Estado())
	}
}

func TestTicket_TransicionesInvalidas(t *testing.T) {
	cases := map[string]func() error{
		"tomar sin asignar":   func() error { return mother.AbiertoTicket().Tomar() },
		"resolver desde abierto": func() error { return mother.AbiertoTicket().Resolver() },
		"cerrar desde abierto": func() error { return mother.AbiertoTicket().Cerrar() },
		"asignar dos veces": func() error {
			tk := mother.AsignadoTicket(uuid.New())
			return tk.Asignar(uuid.New())
		},
		"cerrar desde en_curso": func() error { return mother.EnCursoTicket().Cerrar() },
	}
	for name, fn := range cases {
		t.Run(name, func(t *testing.T) {
			if err := fn(); !errors.Is(err, model.ErrTransicionInvalida) {
				t.Fatalf("esperaba ErrTransicionInvalida, got %v", err)
			}
		})
	}
}

func TestNewTicket_AsuntoVacio(t *testing.T) {
	_, err := model.NewTicket(uuid.New(), valueobject.ChannelWhatsApp, mother.ValidSolicitante(), "   ", valueobject.PriorityMedia)
	if !errors.Is(err, model.ErrAsuntoVacio) {
		t.Fatalf("esperaba ErrAsuntoVacio, got %v", err)
	}
}

func TestTicket_EmiteEventos(t *testing.T) {
	tk := mother.AbiertoTicket()
	_ = tk.Asignar(uuid.New())
	events := tk.PullEvents()
	if len(events) < 3 { // creado + asignado + transicionado
		t.Fatalf("esperaba >=3 eventos, got %d", len(events))
	}
	// PullEvents debe vaciar el buffer.
	if again := tk.PullEvents(); len(again) != 0 {
		t.Fatalf("PullEvents debe vaciar, got %d", len(again))
	}
}
