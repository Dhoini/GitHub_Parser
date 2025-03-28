package metrics

import (
	"net/http"
	"time"

	"github.com/Dhoini/GitHub_Parser/pkg/utils/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics содержит все метрики сервиса
type Metrics struct {
	logger *logger.Logger

	// Счетчики запросов к GitHub API
	APIRequests *prometheus.CounterVec

	// Гистограмма времени ответа GitHub API
	APILatency *prometheus.HistogramVec

	// Счетчики парсинга сущностей
	ParsedRepositories prometheus.Counter
	ParsedIssues       prometheus.Counter
	ParsedPullRequests prometheus.Counter
	ParsedUsers        prometheus.Counter

	// Счетчики ошибок
	Errors *prometheus.CounterVec

	// Информация о заданиях парсинга
	ParsingJobs       *prometheus.GaugeVec
	ParsingJobsTotal  prometheus.Counter
	ParsingJobsErrors prometheus.Counter

	// Информация о состоянии БД
	DBConnectionsOpen prometheus.Gauge
	DBOperations      *prometheus.CounterVec
}

// NewMetrics создает новый экземпляр метрик
func NewMetrics(logger *logger.Logger) *Metrics {
	m := &Metrics{
		logger: logger,

		APIRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "github_parser_api_requests_total",
				Help: "Total number of requests to GitHub API",
			},
			[]string{"endpoint"},
		),

		APILatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "github_parser_api_latency_seconds",
				Help:    "GitHub API request latency in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"endpoint"},
		),

		ParsedRepositories: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "github_parser_parsed_repositories_total",
				Help: "Total number of parsed repositories",
			},
		),

		ParsedIssues: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "github_parser_parsed_issues_total",
				Help: "Total number of parsed issues",
			},
		),

		ParsedPullRequests: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "github_parser_parsed_pull_requests_total",
				Help: "Total number of parsed pull requests",
			},
		),

		ParsedUsers: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "github_parser_parsed_users_total",
				Help: "Total number of parsed users",
			},
		),

		Errors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "github_parser_errors_total",
				Help: "Total number of errors",
			},
			[]string{"type"},
		),

		ParsingJobs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "github_parser_parsing_jobs",
				Help: "Current number of parsing jobs",
			},
			[]string{"status"},
		),

		ParsingJobsTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "github_parser_parsing_jobs_total",
				Help: "Total number of parsing jobs",
			},
		),

		ParsingJobsErrors: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "github_parser_parsing_jobs_errors_total",
				Help: "Total number of parsing job errors",
			},
		),

		DBConnectionsOpen: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "github_parser_db_connections_open",
				Help: "Number of open database connections",
			},
		),

		DBOperations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "github_parser_db_operations_total",
				Help: "Total number of database operations",
			},
			[]string{"operation", "entity"},
		),
	}

	// Регистрируем метрики в Prometheus
	prometheus.MustRegister(
		m.APIRequests,
		m.APILatency,
		m.ParsedRepositories,
		m.ParsedIssues,
		m.ParsedPullRequests,
		m.ParsedUsers,
		m.Errors,
		m.ParsingJobs,
		m.ParsingJobsTotal,
		m.ParsingJobsErrors,
		m.DBConnectionsOpen,
		m.DBOperations,
	)

	logger.Info("Metrics initialized")
	return m
}

// ObserveAPILatency измеряет время запроса к GitHub API
func (m *Metrics) ObserveAPILatency(endpoint string, start time.Time) {
	duration := time.Since(start).Seconds()
	m.APILatency.WithLabelValues(endpoint).Observe(duration)
}

// StartMetricsServer запускает HTTP сервер для Prometheus метрик
func StartMetricsServer(addr string, logger *logger.Logger) {
	http.Handle("/metrics", promhttp.Handler())
	logger.Info("Starting metrics server on %s", addr)

	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			logger.Error("Metrics server failed: %v", err)
		}
	}()
}
