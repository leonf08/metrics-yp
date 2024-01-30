package errorhandling

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrRetriable is a retriable error.
var ErrRetriable = errors.New("retriable error")

// ErrRetryFailed is returned when all retries failed.
var ErrRetryFailed = errors.New("all retries failed")

const (
	attempts   = 3
	difference = 2 * time.Second
)

func isRetriable(err error) bool {
	return errors.Is(err, ErrRetriable)
}

// Retry retries the function f until success, retry limit is reached or context is done.
func Retry(ctx context.Context, f func() error) error {
	var (
		try   int
		delay = time.Second
		err   error
	)

	for try <= attempts {
		err = f()
		if err == nil {
			return nil
		}

		if !isRetriable(err) {
			return err
		}

		if try == attempts {
			break
		}

		ticker := time.NewTicker(delay)

		select {
		case <-ctx.Done():
			return err
		case <-ticker.C:
			delay += difference
			try++
		}
	}

	return fmt.Errorf("%w: %s", ErrRetryFailed, err)
}
