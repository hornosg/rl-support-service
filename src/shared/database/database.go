// Package database — conexión a Postgres y sesión de tenant fail-closed (RLS).
package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

const connKey = "tenant_db_conn"

// Connect abre el pool contra lab-postgres usando el rol de app (sin DDL, RULE-09).
func Connect() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		env("DATABASE_HOST", "lab-postgres"), env("DATABASE_PORT", "5432"),
		env("DATABASE_USER", "support_service_app"), os.Getenv("DATABASE_PASSWORD"),
		env("DATABASE_NAME", "support_service"), env("DATABASE_SSL_MODE", "disable"),
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return db, db.Ping()
}

// TenantSession fija UNA conexión por request y setea app.tenant_id en ESA conexión,
// de modo que las policies RLS (002_rls.sql) apliquen aunque el handler olvide filtrar.
// El tenant sale del header X-Tenant-ID, ya validado contra el JWT por go-shared.
func TenantSession(db *sql.DB, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenant := c.GetHeader("X-Tenant-ID")
		role := c.GetHeader("X-User-Role")
		if tenant == "" && role != "system_admin" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing X-Tenant-ID"})
			return
		}

		conn, err := db.Conn(c.Request.Context())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "db unavailable"})
			return
		}
		defer conn.Close()

		// set_config(..., false) = a nivel de sesión de ESTA conexión fijada.
		if _, err := conn.ExecContext(c.Request.Context(),
			"SELECT set_config('app.tenant_id', $1, false)", tenant); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "tenant session failed"})
			return
		}

		// Break-glass del owner: sólo system_admin activa app.role; queda auditado.
		if role == "system_admin" {
			if _, err := conn.ExecContext(c.Request.Context(),
				"SELECT set_config('app.role', 'system_admin', false)"); err == nil {
				log.Warn("break_glass_access",
					zap.String("event", "system_admin_cross_tenant"),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("tenant_hint", tenant),
				)
			}
		}

		c.Set(connKey, conn)
		c.Next()
	}
}

// Conn devuelve la conexión fijada del request. Los handlers la usan para sus queries.
func Conn(c *gin.Context) *sql.Conn {
	if v, ok := c.Get(connKey); ok {
		if conn, ok := v.(*sql.Conn); ok {
			return conn
		}
	}
	return nil
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

var _ = context.Background
