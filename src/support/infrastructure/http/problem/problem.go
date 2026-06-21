// Package problem — respuestas de error en formato Problem Details (RFC 7807),
// la forma de error del contrato (D-02). Content-Type: application/problem+json.
package problem

import "github.com/gin-gonic/gin"

type Problem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
}

// Write emite un problema y corta la cadena del handler.
func Write(c *gin.Context, status int, title, detail string) {
	c.Header("Content-Type", "application/problem+json")
	c.JSON(status, Problem{
		Type:   "about:blank",
		Title:  title,
		Status: status,
		Detail: detail,
	})
	c.Abort()
}
