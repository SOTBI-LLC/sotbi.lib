package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConnectAcceptsLiteralDSN(t *testing.T) {
	dsn := "postgres://postgres:postgres@localhost:55004/energy?sslmode=disable"

	require.Equal(t, dsn, formatDSN(dsn, "energy"))
}
