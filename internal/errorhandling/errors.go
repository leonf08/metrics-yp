package errorhandling

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrRetriable = errors.New("retriable error")

const (
	attempts   = 3
	difference = 2 * time.Second
)

func isRetriable(err error) bool {
	return errors.Is(err, ErrRetriable)
}

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

	return fmt.Errorf("%w: all retries failed", err)
}
