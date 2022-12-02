package errormonitoring

import (
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
)

func Initialise() {
	err := sentry.Init(sentry.ClientOptions{
		// Either set your DSN here or set the SENTRY_DSN environment variable.
		Dsn: os.Getenv("SENTRY_DSN"),
		// Either set environment and release here or set the SENTRY_ENVIRONMENT
		// and SENTRY_RELEASE environment variables.
		Environment: os.Getenv("GIN_MODE"),
		// Enable printing of SDK debug messages.
		// Useful when getting started or trying to figure something out.
		Debug: false,
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		TracesSampleRate: func() float64 {
			if os.Getenv("GIN_MODE") == "production" {
				return 0.7
			}
			return 1.0
		}(),
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	// Flush buffered events before the program terminates.
	// Set the timeout to the maximum duration the program can afford to wait.
	defer sentry.Flush(2 * time.Second)
}

func ReportError(err error) {
	sentry.CaptureException(err)
}
