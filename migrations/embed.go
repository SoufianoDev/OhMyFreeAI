package migrations

import "embed"

// Files contains forward and development-only rollback SQL migrations.
//
//go:embed *.sql
var Files embed.FS
