package integration

import (
	"context"
	"database/sql"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/nootey/observv/internal/database"
	"github.com/nootey/observv/internal/domains/terminal"
)

type TestContainer struct {
	Container *postgres.PostgresContainer
	DB        *sql.DB
	Queries   *database.Queries
	Services  *Services
	Metrics   *sdkmetric.ManualReader
}

type Services struct {
	TerminalService *terminal.TerminalService
	// Add more services here
}

type ServiceIntegrationSuite struct {
	suite.Suite
	TC  *TestContainer
	Ctx context.Context
}

func (s *ServiceIntegrationSuite) SetupSuite() {
	s.Ctx = context.Background()

	// Start PostgreSQL container
	container, err := postgres.Run(s.Ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	s.Require().NoError(err, "failed to start container")

	// Connect to container
	connStr, err := container.ConnectionString(s.Ctx, "sslmode=disable")
	s.Require().NoError(err, "failed to get connection string")

	db, err := sql.Open("postgres", connStr)
	s.Require().NoError(err, "failed to connect to database")

	// Run migrations
	err = goose.SetDialect("postgres")
	s.Require().NoError(err, "failed to set goose dialect")

	migrationsPath := filepath.Join("..", "..", "..", "migrations")
	err = goose.Up(db, migrationsPath)
	s.Require().NoError(err, "migrations failed")

	// Create queries
	queries := database.New(db)

	// Delta temporality: each collection holds only the changes since the last one
	metricReader := sdkmetric.NewManualReader(sdkmetric.WithTemporalitySelector(deltaTemporality))
	meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(metricReader))

	// Initialize services
	terminalService := terminal.NewTerminalServiceWithMeter(queries, meterProvider)

	s.TC = &TestContainer{
		Container: container,
		DB:        db,
		Queries:   queries,
		Services: &Services{
			TerminalService: terminalService,
		},
		Metrics: metricReader,
	}
}

func (s *ServiceIntegrationSuite) SetupTest() {
	_, err := s.TC.DB.Exec(`TRUNCATE TABLE terminals RESTART IDENTITY CASCADE`)
	s.Require().NoError(err, "failed to truncate tables")

	// Drop metric changes from earlier tests
	s.CollectMetrics()
}

func (s *ServiceIntegrationSuite) TearDownSuite() {
	if s.TC.DB != nil {
		_ = s.TC.DB.Close()
	}

	if s.TC.Container != nil {
		if err := s.TC.Container.Terminate(s.Ctx); err != nil {
			s.T().Logf("container cleanup warning: %s", err)
		}
	}
}

type MetricData metricdata.ResourceMetrics

// CollectMetrics returns the metric changes since the last collection.
func (s *ServiceIntegrationSuite) CollectMetrics() MetricData {
	var rm metricdata.ResourceMetrics
	s.Require().NoError(s.TC.Metrics.Collect(s.Ctx, &rm), "failed to collect metrics")
	return MetricData(rm)
}

// Value sums the int64 points of a metric that carry all of attrs.
func (m MetricData) Value(name string, attrs ...attribute.KeyValue) int64 {
	var total int64
	for _, scope := range m.ScopeMetrics {
		for _, metric := range scope.Metrics {
			if metric.Name != name {
				continue
			}
			switch data := metric.Data.(type) {
			case metricdata.Sum[int64]:
				total += sumMatching(data.DataPoints, attrs)
			case metricdata.Gauge[int64]:
				total += sumMatching(data.DataPoints, attrs)
			}
		}
	}
	return total
}

func sumMatching(points []metricdata.DataPoint[int64], attrs []attribute.KeyValue) int64 {
	var total int64
	for _, p := range points {
		if hasAll(p.Attributes, attrs) {
			total += p.Value
		}
	}
	return total
}

func hasAll(set attribute.Set, attrs []attribute.KeyValue) bool {
	for _, want := range attrs {
		got, ok := set.Value(want.Key)
		if !ok || got != want.Value {
			return false
		}
	}
	return true
}

func deltaTemporality(sdkmetric.InstrumentKind) metricdata.Temporality {
	return metricdata.DeltaTemporality
}
