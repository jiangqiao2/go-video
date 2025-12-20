package observability

import (
	"log"
	"os"
	"strings"

	"github.com/grafana/pyroscope-go"
)

// StartProfiling enables continuous CPU and heap profiling when PROFILE_ENABLED is set.
// Profiles are sent to a Pyroscope/Grafana Profiles server configured via env vars.
func StartProfiling(service string) {
	if !isEnabled(os.Getenv("PROFILE_ENABLED")) {
		return
	}

	server := strings.TrimSpace(os.Getenv("PROFILE_SERVER"))
	if server == "" {
		log.Println("profiling: PROFILE_SERVER not set, disabling profiling")
		return
	}

	env := firstNonEmpty(
		os.Getenv("PROFILE_ENV"),
		os.Getenv("ENV"),
		os.Getenv("APP_ENV"),
	)
	if env == "" {
		env = "prod"
	}

	host, err := os.Hostname()
	if err != nil {
		host = ""
	}

	_, err = pyroscope.Start(pyroscope.Config{
		ApplicationName: service + ".go-video",
		ServerAddress:   server,
		Tags: map[string]string{
			"service":  service,
			"env":      env,
			"instance": host,
		},
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileAllocSpace,
		},
	})
	if err != nil {
		log.Printf("profiling: failed to start: %v\n", err)
	}
}

func isEnabled(v string) bool {
	v = strings.TrimSpace(strings.ToLower(v))
	return v == "1" || v == "true" || v == "yes" || v == "y"
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
