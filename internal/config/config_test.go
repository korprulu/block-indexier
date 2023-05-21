package config

import "testing"

func TestConfig(t *testing.T) {
	t.Parallel()

	cfg, err := Load()
	if err != nil {
		t.Error(err)
	}

	if cfg.Postgres.DB != "testing" {
		t.Errorf("Expected testing, got %s", cfg.Postgres.DB)
	}
}
