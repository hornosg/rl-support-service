package valueobject

import "errors"

// Channel — canal de origen del ticket (G-07). Es un adaptador de entrada, no el dueño
// del dominio (P-06). WhatsApp es el canal del POC; web/email quedan habilitados para H3.
type Channel string

const (
	ChannelWhatsApp Channel = "whatsapp"
	ChannelWeb      Channel = "web"
	ChannelEmail    Channel = "email"
)

var ErrCanalInvalido = errors.New("valueobject: canal inválido")

// NewChannel valida contra el conjunto cerrado de canales soportados.
func NewChannel(s string) (Channel, error) {
	switch Channel(s) {
	case ChannelWhatsApp, ChannelWeb, ChannelEmail:
		return Channel(s), nil
	default:
		return "", ErrCanalInvalido
	}
}

func (c Channel) String() string { return string(c) }
