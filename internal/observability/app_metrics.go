package observability

import "github.com/prometheus/client_golang/prometheus"

var (
	AppVersion = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "app_version_info",
			Help: "Application version info",
		},
		[]string{"version"},
	)
)

func ExposeAppVersion(version string) {
	AppVersion.WithLabelValues(version).Set(1)
}

func init() {
	prometheus.MustRegister(AppVersion)
}
