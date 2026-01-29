package context

import (
	"fmt"
	"foundation"
	"sync"
)

/* Execution represents a single executable operation with undo capability. */
type Execution[T any] struct {
	Execute func(ctx T) error
	Undo    func(ctx T) error
}

/* LoggerFunc is a function type for logging messages.
Clients can inject their own logging implementation (e.g., echo logger). */
type LoggerFunc func(message string)

/* UndoPolicy specifies how undo/redo operations handle failures. */
type UndoPolicy int

const (
	/* UndoPolicyStrict moves the cursor only on successful operations.
	If undo/redo fails, the cursor remains unchanged and an error is returned. */
	UndoPolicyStrict UndoPolicy = iota

	/* UndoPolicyBestEffort moves the cursor even when operations fail.
	This is an explicit opt-in for best-effort semantics. */
	UndoPolicyBestEffort
)

/* ExecutionHistory tracks executed operations for a context.

Invariants (enforced by structure):
- past: Always contains only executed operations (never undone)
- future: Always contains only undone operations (redo stack, never executed)
- batch: Always contains only unexecuted operations (queued, never executed)
- past and future are mutually exclusive
- batch is independent of past/future */
type ExecutionHistory[T any] struct {
	lock        sync.Mutex
	past        []Execution[T]  // executed & undoable (never contains undone ops)
	future      []Execution[T]  // redo stack (never contains executed ops)
	batch       []Execution[T]  // queued, never executed
	undoPolicy  UndoPolicy
	nextBatchID uint64
	logger      LoggerFunc
}

/* Batch represents a group of executions that can be committed or aborted together.
Batch handles are single-use and become invalid after commit or abort. */
type Batch[T any] struct {
	manager   *ExecutionHistory[T]
	batchID   uint64
	committed bool
}

/*
ExecutionHistoryCreate creates a new execution history manager for tracking and undoing operations.

The manager uses strict undo policy by default, meaning undo/redo operations only move the cursor
on success. This prevents state corruption and maintains causal reversibility.

Use cases:
- Implementing undo/redo functionality in applications
- Tracking operation history for audit purposes
- Managing stateful batch operations with commit/abort

Time complexity: O(1) - simple struct initialization
Space complexity: O(1) - fixed size struct allocation

Prerequisites:
- None - returns a ready-to-use manager

Edge cases:
- Manager starts with empty past, future, and batch
- No active batch initially
- Thread-safe for concurrent access

The manager maintains three separate slices to prevent mixing of different execution states.
This structural separation eliminates edge cases and makes state transitions obvious.
*/
func ExecutionHistoryCreate[T any]() *ExecutionHistory[T] {
	return ExecutionHistoryCreateWithPolicy[T](UndoPolicyStrict)
}

/*
ExecutionHistoryCreateWithPolicy creates a new execution history manager with the specified undo policy.

Use cases:
- Creating a manager with best-effort undo semantics (explicit opt-in)
- Custom undo/redo behavior requirements

Time complexity: O(1) - simple struct initialization
Space complexity: O(1) - fixed size struct allocation

Prerequisites:
- policy must be a valid UndoPolicy value

Edge cases:
- Manager starts with empty past, future, and batch
- No active batch initially
- Thread-safe for concurrent access

Best-effort policy moves the cursor even when undo/redo fails, which can lead to state corruption.
Use only when you explicitly need best-effort semantics and understand the implications.
*/
func ExecutionHistoryCreateWithPolicy[T any](policy UndoPolicy) *ExecutionHistory[T] {
	return &ExecutionHistory[T]{
		past:        make([]Execution[T], 0),
		future:      make([]Execution[T], 0),
		batch:       make([]Execution[T], 0),
		undoPolicy:  policy,
		nextBatchID: 1,
		logger:      nil,
	}
}

/*
ExecutionHistoryExecute executes an operation and adds it to history.

If a batch is active, the execution is queued in the batch (not executed).
If no batch is active, the execution runs immediately and is added to past.
When executing outside a batch, any future (redo stack) is cleared.

Use cases:
- Executing operations that need undo capability
- Building up a batch of operations before committing
- Recording operations for audit trails

Time complexity: O(1) - append to slice (amortized)
Space complexity: O(1) - single execution added

Prerequisites:
- manager must be a valid ExecutionHistory instance
- exec.Execute must be a valid function
- exec.Undo can be nil for non-reversible operations

Edge cases:
- If execution fails, operation is not added to history
- If in batch mode, execution is deferred until commit
- Executing outside batch clears future (redo stack)
- Thread-safe - uses mutex for concurrent access

The function maintains the invariant that past contains only executed operations.
*/
func ExecutionHistoryExecute[T any](manager *ExecutionHistory[T], ctx T, exec Execution[T]) error {
	if exec.Execute == nil {
		return fmt.Errorf("execution function cannot be nil")
	}

	var err error
	foundation.WithLock(&manager.lock, func() {
		if len(manager.batch) > 0 {
			// In batch mode - queue the execution (do not execute)
			manager.batch = append(manager.batch, exec)
		} else {
			// Not in batch mode - execute immediately
			// Clear future (redo stack) when executing new operation
			manager.future = make([]Execution[T], 0)

			// Execute the operation
			err = exec.Execute(ctx)
			if err != nil {
				return
			}

			// Add to past (executed operations)
			manager.past = append(manager.past, exec)
		}
	})

	return err
}

/*
ExecutionHistoryUndo undoes the most recent execution.

With strict policy (default): Cursor only moves if undo succeeds.
With best-effort policy: Cursor moves even if undo fails.

Use cases:
- Implementing undo functionality in user interfaces
- Reverting operations that have side effects
- Rolling back state changes

Time complexity: O(1) - single undo operation
Space complexity: O(1) - no additional allocations

Prerequisites:
- manager must be a valid ExecutionHistory instance
- At least one execution must be in past
- Execution must have a non-nil Undo function

Edge cases:
- Returns error if no operations to undo (past is empty)
- Returns error if undo function is nil
- With strict policy: cursor unchanged on failure
- With best-effort policy: cursor moves on failure (state corruption possible)
- Thread-safe - uses mutex for concurrent access

The function maintains the invariant that past contains only executed operations
and future contains only undone operations.
*/
func ExecutionHistoryUndo[T any](manager *ExecutionHistory[T], ctx T) error {
	var err error
	var exec Execution[T]

	foundation.WithLock(&manager.lock, func() {
		if len(manager.past) == 0 {
			err = fmt.Errorf("no operations to undo")
			return
		}

		exec = manager.past[len(manager.past)-1]
		if exec.Undo == nil {
			err = fmt.Errorf("undo function is nil for execution")
			return
		}
	})

	if err != nil {
		return err
	}

	// Execute undo outside of lock
	undoErr := exec.Undo(ctx)

	foundation.WithLock(&manager.lock, func() {
		if undoErr != nil {
			if manager.undoPolicy == UndoPolicyStrict {
				// Strict: don't move cursor on failure
				err = fmt.Errorf("undo failed: %w", undoErr)
				return
			}
			// Best-effort: move cursor even on failure
			if manager.logger != nil {
				manager.logger(fmt.Sprintf("undo failed but continuing (best-effort): %v", undoErr))
			}
		}

		// Move from past to future
		manager.past = manager.past[:len(manager.past)-1]
		manager.future = append(manager.future, exec)

		if undoErr != nil {
			err = fmt.Errorf("undo failed (best-effort): %w", undoErr)
		}
	})

	return err
}

/*
ExecutionHistoryRedo redoes the next execution in the redo stack.

With strict policy (default): Cursor only moves if redo succeeds.
With best-effort policy: Cursor moves even if redo fails.

Use cases:
- Implementing redo functionality after undo operations
- Re-applying operations that were undone
- Restoring state after undo

Time complexity: O(1) - single redo operation
Space complexity: O(1) - no additional allocations

Prerequisites:
- manager must be a valid ExecutionHistory instance
- At least one execution must be in future
- Execution must have a non-nil Execute function

Edge cases:
- Returns error if no operations to redo (future is empty)
- Returns error if execute function is nil
- With strict policy: cursor unchanged on failure
- With best-effort policy: cursor moves on failure (state corruption possible)
- Thread-safe - uses mutex for concurrent access

The function maintains the invariant that past contains only executed operations
and future contains only undone operations.
*/
func ExecutionHistoryRedo[T any](manager *ExecutionHistory[T], ctx T) error {
	var err error
	var exec Execution[T]

	foundation.WithLock(&manager.lock, func() {
		if len(manager.future) == 0 {
			err = fmt.Errorf("no operations to redo")
			return
		}

		exec = manager.future[len(manager.future)-1]
		if exec.Execute == nil {
			err = fmt.Errorf("execute function is nil for execution")
			return
		}
	})

	if err != nil {
		return err
	}

	// Execute redo outside of lock
	redoErr := exec.Execute(ctx)

	foundation.WithLock(&manager.lock, func() {
		if redoErr != nil {
			if manager.undoPolicy == UndoPolicyStrict {
				// Strict: don't move cursor on failure
				err = fmt.Errorf("redo failed: %w", redoErr)
				return
			}
			// Best-effort: move cursor even on failure
			if manager.logger != nil {
				manager.logger(fmt.Sprintf("redo failed but continuing (best-effort): %v", redoErr))
			}
		}

		// Move from future to past
		manager.future = manager.future[:len(manager.future)-1]
		manager.past = append(manager.past, exec)

		if redoErr != nil {
			err = fmt.Errorf("redo failed (best-effort): %w", redoErr)
		}
	})

	return err
}

/*
ExecutionHistoryBatchStart starts a new batch operation.

All subsequent executions will be queued in the batch until the batch is committed or aborted.
Only one batch can be active at a time. Starting a new batch while one is active will panic.

Use cases:
- Grouping multiple operations into a single deferred execution unit
- Batching operations for performance
- Deferring execution until validation is complete

Time complexity: O(1) - simple state update
Space complexity: O(1) - returns a batch handle

Prerequisites:
- manager must be a valid ExecutionHistory instance
- No batch should already be active

Edge cases:
- Panics if batch already active (prevents silent overwrite)
- Returns a Batch handle that must be used for commit or abort
- Batch handle is single-use (becomes invalid after commit/abort)
- Thread-safe - uses mutex for concurrent access

Batch is a pure deferral mechanism, not a transaction system.
Operations are queued, not executed, until commit.
*/
func ExecutionHistoryBatchStart[T any](manager *ExecutionHistory[T]) *Batch[T] {
	var batch *Batch[T]
	foundation.WithLock(&manager.lock, func() {
		if len(manager.batch) > 0 {
			panic("cannot start batch: batch already active")
		}

		batchID := manager.nextBatchID
		manager.nextBatchID++

		batch = &Batch[T]{
			manager:   manager,
			batchID:   batchID,
			committed: false,
		}
	})
	return batch
}

/*
ExecutionHistoryBatchHasActiveBatch checks if a batch has an active batch that can be committed or aborted.

This checks both that the batch itself is not committed and that the manager has an active batch.

Use cases:
- Checking if a batch can be committed or aborted
- Validating batch state before operations
- Determining if frame operations are no-ops

Time complexity: O(1) - simple state checks
Space complexity: O(1) - no allocations

Prerequisites:
- batch must be a valid Batch instance (can be nil)

Edge cases:
- Returns false if batch is nil
- Returns false if batch is already committed
- Returns false if manager has no active batch (empty batch list)
- Returns true if batch is valid and manager has queued operations
- Thread-safe - uses mutex for concurrent access
*/
func ExecutionHistoryBatchHasActiveBatch[T any](batch *Batch[T]) bool {
	if batch == nil {
		return false
	}
	
	var hasActiveBatch bool
	foundation.WithLock(&batch.manager.lock, func() {
		hasActiveBatch = !batch.committed && len(batch.manager.batch) > 0
	})
	return hasActiveBatch
}

// executionHistoryValidateBatch validates that a batch is valid for the given operation.
// Panics if batch is invalid (already committed/aborted or no active batch).
func executionHistoryValidateBatch[T any](manager *ExecutionHistory[T], batch *Batch[T], operation string) {
	foundation.WithLock(&manager.lock, func() {
		if batch.committed {
			panic(fmt.Sprintf("batch %d already committed or aborted, cannot %s", batch.batchID, operation))
		}
		if len(manager.batch) == 0 {
			panic(fmt.Sprintf("no active batch for %s", operation))
		}
	})
}

/*
ExecutionHistoryBatchCommit commits all executions in the batch.

All operations queued in the batch are executed sequentially.
If any execution fails, the function stops and returns an error.
The batch remains unchanged on failure (operations were never executed, so no rollback needed).

On success, all operations are moved from batch to past.

Use cases:
- Finalizing batch operations after validation
- Applying queued operations
- Committing a group of deferred operations

Time complexity: O(n) where n is the number of executions in the batch
Space complexity: O(1) - processes executions sequentially

Prerequisites:
- batch must be a valid Batch instance from ExecutionHistoryBatchStart
- batch must not have been already committed or aborted
- All execution functions must be valid

Edge cases:
- Panics if batch already committed/aborted
- Panics if no active batch
- Empty batches are valid and return nil
- On first failure: stops execution, returns error, batch unchanged
- On success: moves all to past, clears batch, marks batch committed
- Thread-safe - uses mutex for concurrent access

Batch is pure deferral - operations are queued, not executed, until commit.
No rollback is needed on failure because nothing was executed.
*/
func ExecutionHistoryBatchCommit[T any](batch *Batch[T], ctx T) error {
	if batch == nil {
		return fmt.Errorf("batch cannot be nil")
	}

	manager := batch.manager
	executionHistoryValidateBatch(manager, batch, "commit")

	var executions []Execution[T]
	foundation.WithLock(&manager.lock, func() {
		executions = make([]Execution[T], len(manager.batch))
		copy(executions, manager.batch)
	})

	if len(executions) == 0 {
		// Empty batch - mark committed and return
		foundation.WithLock(&manager.lock, func() {
			manager.batch = make([]Execution[T], 0)
			batch.committed = true
		})
		return nil
	}

	// Execute all operations in the batch sequentially
	for i, exec := range executions {
		if exec.Execute == nil {
			return fmt.Errorf("execute function is nil for execution at batch index %d", i)
		}

		execErr := exec.Execute(ctx)
		if execErr != nil {
			// Stop on first failure - batch unchanged (nothing was executed)
			return fmt.Errorf("batch execution failed at index %d: %w", i, execErr)
		}
	}

	// Success - move all to past and clear batch
	foundation.WithLock(&manager.lock, func() {
		// Clear future (redo stack) when committing batch
		manager.future = make([]Execution[T], 0)
		manager.past = append(manager.past, executions...)
		manager.batch = make([]Execution[T], 0)
		batch.committed = true
	})

	return nil
}

/*
ExecutionHistoryBatchAbort aborts the batch and drops all queued operations.

The batch is simply cleared - no undo operations are needed because nothing was executed.
This is a pure deferral mechanism, not a transaction system.

Use cases:
- Canceling a group of operations
- Dropping queued operations on error
- Aborting deferred execution

Time complexity: O(1) - slice clear
Space complexity: O(1) - clears references

Prerequisites:
- batch must be a valid Batch instance from ExecutionHistoryBatchStart
- batch must not have been already committed or aborted

Edge cases:
- Panics if batch already committed/aborted
- Panics if no active batch
- Empty batches are valid and return nil
- Simply clears batch (no undo needed - nothing was executed)
- Thread-safe - uses mutex for concurrent access

Batch is pure deferral - abort simply drops queued operations.
No undo is needed because operations were never executed.
*/
func ExecutionHistoryBatchAbort[T any](batch *Batch[T], ctx T) error {
	if batch == nil {
		return fmt.Errorf("batch cannot be nil")
	}

	manager := batch.manager
	executionHistoryValidateBatch(manager, batch, "abort")

	foundation.WithLock(&manager.lock, func() {
		manager.batch = make([]Execution[T], 0)
		batch.committed = true
	})

	return nil
}

/*
ExecutionHistoryClear clears all execution history.

Use cases:
- Resetting the execution manager
- Clearing history after a major state change
- Freeing memory when history is no longer needed

Time complexity: O(1) - slice reset
Space complexity: O(1) - clears references, memory may be reclaimed by GC

Prerequisites:
- manager must be a valid ExecutionHistory instance

Edge cases:
- Clears all past, future, and batch
- Resets nextBatchID to 1
- Thread-safe - uses mutex for concurrent access

After clearing, the manager is in the same state as a newly created manager.
*/
func ExecutionHistoryClear[T any](manager *ExecutionHistory[T]) {
	foundation.WithLock(&manager.lock, func() {
		manager.past = make([]Execution[T], 0)
		manager.future = make([]Execution[T], 0)
		manager.batch = make([]Execution[T], 0)
		manager.nextBatchID = 1
	})
}

/*
ExecutionHistoryCanUndo checks if an undo operation is possible.

Use cases:
- Enabling/disabling undo UI buttons
- Checking state before attempting undo
- Validating undo availability

Time complexity: O(1) - simple comparison
Space complexity: O(1) - no allocations

Prerequisites:
- manager must be a valid ExecutionHistory instance

Edge cases:
- Returns false if past is empty (no operations to undo)
- Returns false if last execution has nil undo function
- Thread-safe - uses mutex for concurrent access

This is a convenience function for checking undo availability without
attempting the undo operation.
*/
func ExecutionHistoryCanUndo[T any](manager *ExecutionHistory[T]) bool {
	var canUndo bool
	foundation.WithLock(&manager.lock, func() {
		canUndo = len(manager.past) > 0 && manager.past[len(manager.past)-1].Undo != nil
	})
	return canUndo
}

/*
ExecutionHistoryCanRedo checks if a redo operation is possible.

Use cases:
- Enabling/disabling redo UI buttons
- Checking state before attempting redo
- Validating redo availability

Time complexity: O(1) - simple comparison
Space complexity: O(1) - no allocations

Prerequisites:
- manager must be a valid ExecutionHistory instance

Edge cases:
- Returns false if future is empty (no operations to redo)
- Returns false if last execution has nil execute function
- Thread-safe - uses mutex for concurrent access

This is a convenience function for checking redo availability without
attempting the redo operation.
*/
func ExecutionHistoryCanRedo[T any](manager *ExecutionHistory[T]) bool {
	var canRedo bool
	foundation.WithLock(&manager.lock, func() {
		canRedo = len(manager.future) > 0 && manager.future[len(manager.future)-1].Execute != nil
	})
	return canRedo
}

/*
ExecutionHistorySetLogger sets a logger function for the execution history manager.

The logger will be called when operations fail during best-effort undo/redo,
allowing clients to inject their own logging implementation.

Use cases:
- Integrating with echo logger
- Custom logging systems
- Debugging execution failures

Time complexity: O(1) - simple assignment
Space complexity: O(1) - stores function reference

Prerequisites:
- manager must be a valid ExecutionHistory instance
- logger can be nil to disable logging

Edge cases:
- Setting logger to nil disables logging
- Thread-safe - uses mutex for concurrent access

Example usage with echo:
	manager := context.ExecutionHistoryCreate[MyContext]()
	context.ExecutionHistorySetLogger(manager, func(msg string) {
		echo.On(mySystemUUID).Error(msg)
	})
*/
func ExecutionHistorySetLogger[T any](manager *ExecutionHistory[T], logger LoggerFunc) {
	foundation.WithLock(&manager.lock, func() {
		manager.logger = logger
	})
}
