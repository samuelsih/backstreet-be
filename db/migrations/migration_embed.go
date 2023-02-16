package migrations

import (
	_ "embed"
)

var (
	//go:embed 000001_create_sources_table.up.sql
	UpCmd string

	//go:embed 000001_create_sources_table.down.sql
	DownCmd string
)
