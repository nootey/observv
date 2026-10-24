package config

import "testing"

// Guards the defaults-as-registry contract: a key must be in setDefaults to be
// overridable by env var, and the env name is the key with dots replaced by
// underscores.
func TestLoadEnvOverrides(t *testing.T) {
	t.Setenv("DATABASE_HOST", "db")
	t.Setenv("DATABASE_NAME", "observv_test")
	t.Setenv("APP_ENVIRONMENT", "production")
	t.Setenv("SEED_TERMINALS", "15")
	t.Setenv("SIMULATOR_API_URL", "http://api:8080")
	t.Setenv("SIMULATOR_STATUS_FLIP_CHANCE", "0.5")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://alloy:4317")
	t.Setenv("OTEL_SDK_DISABLED", "true")
	t.Setenv("CHAOS_ERROR_RATE", "0.25")

	if err := Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	c := Get()

	if c.Database.Host != "db" {
		t.Errorf("Database.Host = %q, want %q", c.Database.Host, "db")
	}
	if c.Database.DBName != "observv_test" {
		t.Errorf("Database.DBName = %q, want %q", c.Database.DBName, "observv_test")
	}
	if c.App.Environment != "production" {
		t.Errorf("App.Environment = %q, want %q", c.App.Environment, "production")
	}
	if c.Seed.Terminals != 15 {
		t.Errorf("Seed.Terminals = %d, want 15", c.Seed.Terminals)
	}
	if c.Simulator.APIURL != "http://api:8080" {
		t.Errorf("Simulator.APIURL = %q, want %q", c.Simulator.APIURL, "http://api:8080")
	}
	if c.Simulator.StatusFlipChance != 0.5 {
		t.Errorf("Simulator.StatusFlipChance = %v, want 0.5", c.Simulator.StatusFlipChance)
	}
	if c.OTel.ExporterOTLPEndpoint != "http://alloy:4317" {
		t.Errorf("OTel.ExporterOTLPEndpoint = %q, want %q", c.OTel.ExporterOTLPEndpoint, "http://alloy:4317")
	}
	if !c.OTel.SDKDisabled {
		t.Error("OTel.SDKDisabled = false, want true")
	}
	if c.Chaos.ErrorRate != 0.25 {
		t.Errorf("Chaos.ErrorRate = %v, want 0.25", c.Chaos.ErrorRate)
	}
}

func TestLoadDefaults(t *testing.T) {
	if err := Load(); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	c := Get()

	if c.Server.Port != "8080" {
		t.Errorf("Server.Port = %q, want %q", c.Server.Port, "8080")
	}
	if c.Seed.Terminals != 10 {
		t.Errorf("Seed.Terminals = %d, want 10", c.Seed.Terminals)
	}
	if c.Simulator.IntervalSeconds != 5 {
		t.Errorf("Simulator.IntervalSeconds = %d, want 5", c.Simulator.IntervalSeconds)
	}
}

func TestLoadRejectsInvalidFlipChance(t *testing.T) {
	t.Setenv("SIMULATOR_STATUS_FLIP_CHANCE", "1.5")

	if err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error for flip chance above 1")
	}
}

func TestLoadRejectsSingleTerminal(t *testing.T) {
	t.Setenv("SEED_TERMINALS", "1")

	if err := Load(); err == nil {
		t.Fatal("Load() error = nil, want error: transfers need at least two terminals")
	}
}
