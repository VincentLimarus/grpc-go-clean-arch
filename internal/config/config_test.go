package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"VincentLimarus/grpc-golang/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	clearConfigEnv(t)

	cfg := config.Load()
	assert.Equal(t, ":50051", cfg.GRPCAddr)
	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, "5432", cfg.DBPort)
	assert.Equal(t, "postgres", cfg.DBUser)
	assert.Equal(t, "Vincent27", cfg.DBPassword)
	assert.Equal(t, "grpc_products", cfg.DBName)
	assert.Equal(t, "disable", cfg.DBSSLMode)
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("GRPC_ADDR", ":9090")
	t.Setenv("DB_HOST", "db.internal")
	t.Setenv("DB_PORT", "6543")
	t.Setenv("DB_USER", "shop")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "shopdb")
	t.Setenv("DB_SSLMODE", "require")

	cfg := config.Load()
	assert.Equal(t, ":9090", cfg.GRPCAddr)
	assert.Equal(t, "db.internal", cfg.DBHost)
	assert.Equal(t, "6543", cfg.DBPort)
	assert.Equal(t, "shop", cfg.DBUser)
	assert.Equal(t, "secret", cfg.DBPassword)
	assert.Equal(t, "shopdb", cfg.DBName)
	assert.Equal(t, "require", cfg.DBSSLMode)

	wantDSN := "host=db.internal port=6543 user=shop password=secret dbname=shopdb sslmode=require"
	assert.Equal(t, wantDSN, cfg.DSN())
}

func clearConfigEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"GRPC_ADDR", "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE"} {
		t.Setenv(k, "")
	}
}
