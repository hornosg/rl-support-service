package valueobject

import (
	"errors"
	"strings"
)

// Solicitante — el cliente final que origina el ticket (G-04). Value Object inmutable,
// comparado por valor. Embebe contacto (nombre + teléfono): es PII del cliente final
// (Ley 25.326) → no se loggea (RULE-03) y debe poder borrarse a pedido del tenant.
type Solicitante struct {
	nombre   string
	telefono string
}

var (
	ErrNombreVacio   = errors.New("valueobject: nombre del solicitante requerido")
	ErrTelefonoVacio = errors.New("valueobject: teléfono del solicitante requerido")
)

// Anonimizado — tombstone que reemplaza la PII tras un borrado (Ley 25.326).
const Anonimizado = "[borrado]"

// Anonimo construye un solicitante con la PII borrada. Se usa al ejercer el derecho
// de supresión: el ticket sobrevive (historial/métricas), la PII no.
func Anonimo() Solicitante {
	return Solicitante{nombre: Anonimizado, telefono: Anonimizado}
}

// EsAnonimo indica si la PII del solicitante ya fue borrada.
func (s Solicitante) EsAnonimo() bool {
	return s.nombre == Anonimizado && s.telefono == Anonimizado
}

// NewSolicitante valida y normaliza (trim) los datos de contacto.
func NewSolicitante(nombre, telefono string) (Solicitante, error) {
	nombre = strings.TrimSpace(nombre)
	telefono = strings.TrimSpace(telefono)
	if nombre == "" {
		return Solicitante{}, ErrNombreVacio
	}
	if telefono == "" {
		return Solicitante{}, ErrTelefonoVacio
	}
	return Solicitante{nombre: nombre, telefono: telefono}, nil
}

func (s Solicitante) Nombre() string   { return s.nombre }
func (s Solicitante) Telefono() string { return s.telefono }

func (s Solicitante) Equals(other Solicitante) bool {
	return s.nombre == other.nombre && s.telefono == other.telefono
}
