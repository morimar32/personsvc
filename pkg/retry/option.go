package retry

import (
	"time"
)

type DbRetryOption func(*DbRetry)

func WithRetry(retry int) DbRetryOption {
	return func(o *DbRetry) {
		o.Retry = retry
	}
}

func WithDelay(delay time.Duration) DbRetryOption {
	return func(o *DbRetry) {
		o.Delay = delay
	}
}
