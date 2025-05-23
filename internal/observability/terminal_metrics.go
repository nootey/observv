package observability

import "github.com/prometheus/client_golang/prometheus"

var (
	TerminalStatusTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "terminal_status_total",
			Help: "Number of terminals by status",
		},
		[]string{"status"},
	)

	TerminalDataSent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "terminal_data_sent_bytes_total",
			Help: "Total bytes of data sent by terminals",
		},
		[]string{"terminal_id"},
	)

	TerminalUptime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "terminal_uptime_seconds",
			Help: "Total uptime of terminals in seconds",
		},
		[]string{"terminal_id"},
	)

	TerminalToggleEvents = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "terminal_toggle_events_total",
			Help: "Number of times a terminal changed status",
		},
		[]string{"terminal_id"},
	)
)

func init() {
	prometheus.MustRegister(TerminalStatusTotal, TerminalDataSent, TerminalUptime, TerminalToggleEvents)
}
