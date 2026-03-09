// apps/backend/src/cli/tools/context.go
package tools

import (
	"context"

	"github.com/lucas/oraculo/apps/backend/src/db"
)

type contextKey string

const dbKey contextKey = "db"

func withDB(ctx context.Context, database *db.DB) context.Context {
	return context.WithValue(ctx, dbKey, database)
}

func dbFromContext(ctx context.Context) *db.DB {
	return ctx.Value(dbKey).(*db.DB)
}
