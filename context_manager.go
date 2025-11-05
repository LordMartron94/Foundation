// Package foundation provides utilities that are re-used in all other libs.
// Said differently: foundation is the foundational lib.
package foundation

import "fmt"

// ContextFunc specifies what to execute during the context.
type ContextFunc[T any] func(data T) error

// ContextCleanupFunc specifies how to clean up the context.
type ContextCleanupFunc[T any] func(data T) error

// WithContext executes the given context function and cleans up using the cleanup function.
// It returns an error if something happens either during execution or cleanup.
func WithContext[T any](ctx T, ctxFn ContextFunc[T], ctxCleanupFn ContextCleanupFunc[T]) (err error) {
	defer func() {
		if cErr := ctxCleanupFn(ctx); cErr != nil && err == nil {
			err = cErr
		}
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in context: %v", r)
		}
	}()
	return ctxFn(ctx)
}
