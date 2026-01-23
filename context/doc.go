// Package context provides utility functions for managing execution history and context operations.
//
// The package provides a generic execution history manager that tracks operations with undo/redo
// capability and supports stateful batch operations (commit/abort) for any context type.
//
// Design Philosophy:
//
// The execution manager enforces strict invariants through structural separation:
// - past: Contains only executed operations (never undone)
// - future: Contains only undone operations (redo stack, never executed)
// - batch: Contains only unexecuted operations (queued, never executed)
//
// This separation eliminates edge cases and makes state transitions obvious and verifiable.
//
// Key Features:
// - Execution tracking: Record operations with execute and undo functions
// - Strict undo/redo: Cursor only moves on successful operations (prevents state corruption)
// - Batch operations: Pure deferral mechanism (queue operations, execute on commit)
// - Invariant enforcement: Panics on misuse (fail-fast, prevents silent corruption)
// - Thread-safe: All operations are protected by mutexes
// - Generic: Works with any context type T
//
// Undo/Redo Semantics:
//
// By default, the manager uses strict undo policy: cursor only moves on successful operations.
// This maintains causal reversibility and prevents state corruption. Best-effort policy
// is available as an explicit opt-in for cases where you need it.
//
// Batch Semantics:
//
// Batch is a pure deferral mechanism, not a transaction system:
// - Operations are queued, not executed, until commit
// - Commit executes all operations sequentially, stops on first failure
// - Abort simply drops queued operations (no undo needed - nothing was executed)
// - Batch handles are single-use (panic on reuse)
//
// Example Usage:
//
//	// Create manager for a specific context type (strict undo by default)
//	manager := context.ExecutionHistoryCreate[MyContext]()
//
//	// Single execution
//	exec := context.Execution[MyContext]{
//		Execute: func(ctx MyContext) error {
//			// Perform operation
//			return nil
//		},
//		Undo: func(ctx MyContext) error {
//			// Reverse operation
//			return nil
//		},
//	}
//	context.ExecutionHistoryExecute(manager, myCtx, exec)
//
//	// Undo the last operation (cursor only moves if undo succeeds)
//	context.ExecutionHistoryUndo(manager, myCtx)
//
//	// Batch operations (pure deferral)
//	batch := context.ExecutionHistoryBatchStart(manager)
//	context.ExecutionHistoryExecute(manager, myCtx, exec1)  // Queued, not executed
//	context.ExecutionHistoryExecute(manager, myCtx, exec2)  // Queued, not executed
//	context.ExecutionHistoryBatchCommit(batch, myCtx)        // Executes all, or
//	context.ExecutionHistoryBatchAbort(batch, myCtx)        // Drops all
//
//	// Inject logger for best-effort undo failures
//	context.ExecutionHistorySetLogger(manager, func(msg string) {
//		echo.On(mySystemUUID).Error(msg)
//	})
//
// The execution manager follows C-style function patterns (no methods on structs) and
// uses explicit context passing, consistent with the foundation library's design philosophy.
// Illegal states cause panics (fail-fast) rather than silent corruption.
package context
