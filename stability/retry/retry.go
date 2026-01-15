package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"
)

var ErrMaxAttemptsExceeded = errors.New("max retry attempts exceeded")

type Strategy interface {
	NextDelay(attempt int) time.Duration
}

type RetryConfig struct {
	MaxAttempts int
	Strategy    Strategy
	ShouldRetry func(error) bool
}

type RetryExecutor struct {
	config RetryConfig
}

func NewRetryExecutor(config RetryConfig) *RetryExecutor {
	if config.MaxAttempts == 0 {
		config.MaxAttempts = 3
	}
	if config.Strategy == nil {
		config.Strategy = NewExponentialBackoff(100*time.Millisecond, 5*time.Second, 2.0)
	}
	if config.ShouldRetry == nil {
		config.ShouldRetry = func(err error) bool { return err != nil }
	}
	return &RetryExecutor{config: config}
}

func (r *RetryExecutor) Execute(fn func() error) error {
	var lastErr error

	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		if !r.config.ShouldRetry(err) {
			return err
		}

		if attempt == r.config.MaxAttempts {
			break
		}

		delay := r.config.Strategy.NextDelay(attempt)
		time.Sleep(delay)
	}

	return fmt.Errorf("%w: %v", ErrMaxAttemptsExceeded, lastErr)
}

func (r *RetryExecutor) ExecuteWithContext(ctx context.Context, fn func(context.Context) error) error {
	var lastErr error

	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := fn(ctx)
		if err == nil {
			return nil
		}

		lastErr = err

		if !r.config.ShouldRetry(err) {
			return err
		}

		if attempt == r.config.MaxAttempts {
			break
		}

		delay := r.config.Strategy.NextDelay(attempt)

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return fmt.Errorf("%w: %v", ErrMaxAttemptsExceeded, lastErr)
}

func (r *RetryExecutor) ExecuteWithCallback(
	fn func() error,
	onRetry func(attempt int, err error, delay time.Duration),
) error {
	var lastErr error

	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		if !r.config.ShouldRetry(err) {
			return err
		}

		if attempt == r.config.MaxAttempts {
			break
		}

		delay := r.config.Strategy.NextDelay(attempt)

		if onRetry != nil {
			onRetry(attempt, err, delay)
		}

		time.Sleep(delay)
	}

	return fmt.Errorf("%w: %v", ErrMaxAttemptsExceeded, lastErr)
}

type ExponentialBackoff struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

func NewExponentialBackoff(initial, max time.Duration, multiplier float64) *ExponentialBackoff {
	return &ExponentialBackoff{
		InitialDelay: initial,
		MaxDelay:     max,
		Multiplier:   multiplier,
	}
}

func (e *ExponentialBackoff) NextDelay(attempt int) time.Duration {
	delay := float64(e.InitialDelay) * math.Pow(e.Multiplier, float64(attempt-1))
	if delay > float64(e.MaxDelay) {
		delay = float64(e.MaxDelay)
	}
	return time.Duration(delay)
}

type FixedDelay struct {
	Delay time.Duration
}

func NewFixedDelay(delay time.Duration) *FixedDelay {
	return &FixedDelay{Delay: delay}
}

func (f *FixedDelay) NextDelay(attempt int) time.Duration {
	return f.Delay
}

type LinearBackoff struct {
	InitialDelay time.Duration
	Increment    time.Duration
	MaxDelay     time.Duration
}

func NewLinearBackoff(initial, increment, max time.Duration) *LinearBackoff {
	return &LinearBackoff{
		InitialDelay: initial,
		Increment:    increment,
		MaxDelay:     max,
	}
}

func (l *LinearBackoff) NextDelay(attempt int) time.Duration {
	delay := l.InitialDelay + time.Duration(attempt-1)*l.Increment
	if delay > l.MaxDelay {
		delay = l.MaxDelay
	}
	return delay
}

type ExponentialBackoffWithJitter struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	JitterFactor float64
}

func NewExponentialBackoffWithJitter(initial, max time.Duration, multiplier, jitter float64) *ExponentialBackoffWithJitter {
	return &ExponentialBackoffWithJitter{
		InitialDelay: initial,
		MaxDelay:     max,
		Multiplier:   multiplier,
		JitterFactor: jitter,
	}
}

func (e *ExponentialBackoffWithJitter) NextDelay(attempt int) time.Duration {
	delay := float64(e.InitialDelay) * math.Pow(e.Multiplier, float64(attempt-1))
	if delay > float64(e.MaxDelay) {
		delay = float64(e.MaxDelay)
	}

	jitter := delay * e.JitterFactor * (rand.Float64()*2 - 1)
	delay += jitter

	if delay < 0 {
		delay = 0
	}

	return time.Duration(delay)
}

type FullJitter struct {
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Multiplier float64
}

func NewFullJitter(base, max time.Duration, multiplier float64) *FullJitter {
	return &FullJitter{
		BaseDelay:  base,
		MaxDelay:   max,
		Multiplier: multiplier,
	}
}

func (f *FullJitter) NextDelay(attempt int) time.Duration {
	maxDelay := float64(f.BaseDelay) * math.Pow(f.Multiplier, float64(attempt-1))
	if maxDelay > float64(f.MaxDelay) {
		maxDelay = float64(f.MaxDelay)
	}

	return time.Duration(rand.Float64() * maxDelay)
}
