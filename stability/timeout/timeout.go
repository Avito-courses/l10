package main

import (
	"context"
	"errors"
	"time"
)

var ErrTimeout = errors.New("operation timeout")

func ExecuteWithTimeout(timeout time.Duration, fn func() error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return ExecuteWithContext(ctx, func(ctx context.Context) error {
		return fn()
	})
}

func ExecuteWithContext(ctx context.Context, fn func(context.Context) error) error {
	errChan := make(chan error, 1)

	go func() {
		errChan <- fn(ctx)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ErrTimeout
	}
}

func ExecuteWithTimeoutAndResult[T any](timeout time.Duration, fn func() (T, error)) (T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	type result struct {
		value T
		err   error
	}

	resultChan := make(chan result, 1)

	go func() {
		val, err := fn()
		resultChan <- result{value: val, err: err}
	}()

	select {
	case res := <-resultChan:
		return res.value, res.err
	case <-ctx.Done():
		var zero T
		return zero, ErrTimeout
	}
}

func ExecuteWithContextAndResult[T any](ctx context.Context, fn func(context.Context) (T, error)) (T, error) {
	type result struct {
		value T
		err   error
	}

	resultChan := make(chan result, 1)

	go func() {
		val, err := fn(ctx)
		resultChan <- result{value: val, err: err}
	}()

	select {
	case res := <-resultChan:
		return res.value, res.err
	case <-ctx.Done():
		var zero T
		return zero, ErrTimeout
	}
}

type TimeoutWrapper struct {
	timeout time.Duration
}

func NewTimeoutWrapper(timeout time.Duration) *TimeoutWrapper {
	return &TimeoutWrapper{timeout: timeout}
}

func (tw *TimeoutWrapper) Execute(fn func() error) error {
	return ExecuteWithTimeout(tw.timeout, fn)
}

func (tw *TimeoutWrapper) ExecuteWithContext(ctx context.Context, fn func(context.Context) error) error {

	timeoutCtx, cancel := context.WithTimeout(ctx, tw.timeout)
	defer cancel()

	return ExecuteWithContext(timeoutCtx, fn)
}

type TimeoutConfig struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type MultiStageTimeout struct {
	stages map[string]time.Duration
}

func NewMultiStageTimeout() *MultiStageTimeout {
	return &MultiStageTimeout{
		stages: make(map[string]time.Duration),
	}
}

func (mt *MultiStageTimeout) AddStage(name string, timeout time.Duration) *MultiStageTimeout {
	mt.stages[name] = timeout
	return mt
}

func (mt *MultiStageTimeout) ExecuteStage(stageName string, fn func() error) error {
	timeout, exists := mt.stages[stageName]
	if !exists {
		return errors.New("unknown stage: " + stageName)
	}

	return ExecuteWithTimeout(timeout, fn)
}

type AdaptiveTimeout struct {
	minTimeout     time.Duration
	maxTimeout     time.Duration
	currentTimeout time.Duration
	successCount   int
	failureCount   int
	adjustFactor   float64
}

func NewAdaptiveTimeout(min, max, initial time.Duration) *AdaptiveTimeout {
	return &AdaptiveTimeout{
		minTimeout:     min,
		maxTimeout:     max,
		currentTimeout: initial,
		adjustFactor:   1.2,
	}
}

func (at *AdaptiveTimeout) Execute(fn func() error) error {
	start := time.Now()
	err := ExecuteWithTimeout(at.currentTimeout, fn)
	elapsed := time.Since(start)

	if err == ErrTimeout {
		at.failureCount++

		at.currentTimeout = time.Duration(float64(at.currentTimeout) * at.adjustFactor)
		if at.currentTimeout > at.maxTimeout {
			at.currentTimeout = at.maxTimeout
		}
	} else if err == nil {
		at.successCount++

		if elapsed < at.currentTimeout/2 {
			at.currentTimeout = time.Duration(float64(at.currentTimeout) / at.adjustFactor)
			if at.currentTimeout < at.minTimeout {
				at.currentTimeout = at.minTimeout
			}
		}
	}

	return err
}

func (at *AdaptiveTimeout) GetCurrentTimeout() time.Duration {
	return at.currentTimeout
}

func (at *AdaptiveTimeout) GetStats() (success, failures int, current time.Duration) {
	return at.successCount, at.failureCount, at.currentTimeout
}
