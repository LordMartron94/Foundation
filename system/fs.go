package system

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

/*
============================================================
FILESYSTEM PRIMITIVES
============================================================

This package provides thin, explicit wrappers around OS-level
filesystem operations.

Design principles:

  • No hidden behavior
  • No implicit decoding or transformation
  • Clear separation between I/O and interpretation layers
  • Errors are contextualized but never swallowed

Higher layers (lexing, parsing, indexing, streaming, etc.)
must perform their own semantic processing on top of these
primitives.
*/

/*
FileExists checks whether a filesystem object exists at the given path.

Behavior:
  - Returns true if os.Stat succeeds
  - Returns false for any error (nonexistent, permission, etc.)

Notes:
  - This does not distinguish between files and directories
  - Intended only for existence probing, not validation
*/
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

/*
FileWriteEmpty creates or truncates a file to zero bytes.

Use cases:
  - Initializing placeholder files
  - Clearing file contents atomically

Permissions:
  - 0644 (owner read/write, group+others read)

Errors:
  - Propagates OS-level write failures
*/
func FileWriteEmpty(path string) error {
	return os.WriteFile(path, []byte{}, 0644)
}

/*
FileWriteBytes writes raw bytes to a file.

Behavior:
  - Creates the file if it does not exist
  - Truncates the file if it already exists
  - Writes bytes as-is (no encoding, no newline normalization)

Permissions:
  - 0644 (owner read/write, group+others read)

Notes:
  - Does not create parent directories. Ensure PathDir(path) exists.
  - This is an I/O boundary primitive; semantic meaning is layered above.

Errors:
  - Wraps OS-level write failures with context
*/
func FileWriteBytes(path string, data []byte) error {
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("FS: write failed: %w", err)
	}
	return nil
}

/*
FileWriteString writes a string to a file.

Behavior:
  - Creates the file if it does not exist
  - Truncates the file if it already exists
  - Writes bytes as-is (no encoding, no newline normalization)

Permissions:
  - 0644 (owner read/write, group+others read)

Notes:
  - Does not create parent directories. Ensure PathDir(path) exists.
  - This is an I/O boundary primitive; semantic meaning is layered above.

Errors:
  - Wraps OS-level write failures with context
*/
func FileWriteString(path string, data string) error {
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		return fmt.Errorf("FS: write failed: %w", err)
	}
	return nil
}

/*
FileRename renames or moves a filesystem object.

Safety invariant:
  - The destination path must not already exist

Errors:
  - Returns an explicit error if destination exists
  - Wraps OS rename failures with context
*/
func FileRename(oldPath, newPath string) error {
	if FileExists(newPath) {
		return fmt.Errorf("FS: destination already exists")
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("FS: rename failed: %w", err)
	}
	return nil
}

/*
FileDelete removes a filesystem object.

Behavior:
  - Deleting a nonexistent path is treated as success
  - Other OS errors are propagated

Use cases:
  - Safe cleanup routines
  - Idempotent delete operations
*/
func FileDelete(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("FS: delete failed: %w", err)
	}
	return nil
}

/*
============================================================
PATH UTILITIES
============================================================

Thin wrappers around filepath for consistency and discoverability.
*/

/*
PathJoin joins path elements using OS-specific separators.
*/
func PathJoin(elem ...string) string {
	return filepath.Join(elem...)
}

/*
PathBase returns the final path component.
*/
func PathBase(path string) string {
	return filepath.Base(path)
}

/*
PathDir returns the parent directory of a path.
*/
func PathDir(path string) string {
	return filepath.Dir(path)
}

/*
PathExt returns the file extension including the leading dot.
*/
func PathExt(path string) string {
	return filepath.Ext(path)
}

/*
DirCreate creates a single directory.

Notes:
  - Does not create parents (non-recursive) unless all is set to true.
  - Mirrors os.Mkdir semantics, and os.MkDirall if all is set to true.

Permissions:
  - 0755 (owner full, group+others read/execute)
*/
func DirCreate(path string, all bool) error {
	if all {
		return os.MkdirAll(path, 0755)
	}

	return os.Mkdir(path, 0755)
}

/*
PathResolveWorkspace resolves a target path relative to the current working directory
or any of its ancestors.

It walks up the directory tree looking for a parent directory that contains
the required path structure. This allows tools executed from subdirectories
(like 'go generate') to correctly resolve paths written relative to the repository root.

Fallback: If the path structure cannot be found in any ancestor, it returns
the path relative to the current working directory.
*/
func PathResolveWorkspace(targetPath string) (string, error) {
	targetPath = filepath.Clean(targetPath)
	if filepath.IsAbs(targetPath) {
		return targetPath, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for d := cwd; ; d = filepath.Dir(d) {
		candidate := filepath.Join(d, targetPath)
		parent := filepath.Dir(candidate)

		// If the parent directory of our target path exists relative to this ancestor,
		// we have found the correct anchor point (e.g., the repo root).
		if fi, statErr := os.Stat(parent); statErr == nil && fi.IsDir() {
			return filepath.Clean(candidate), nil
		}

		// Stop if we hit the filesystem root
		if d == filepath.Dir(d) {
			break
		}
	}

	// Fallback
	return filepath.Join(cwd, targetPath), nil
}

/*
============================================================
DIRECTORY SCANNING
============================================================
*/

/*
ScanCallback defines the interface for directory traversal handlers.

Parameters:

	path  → absolute path of the entry
	name  → file or directory name (base component)
	isDir → true if entry is a directory

Returns:

	skip  → if true and entry is a directory, its subtree is skipped
	error → aborts traversal immediately

Design:
  - Keeps OS-specific types out of higher layers
  - Enables transactional traversal control
*/
type ScanCallback func(path string, name string, isDir bool) (bool, error)

/*
ScanRecursive walks a directory tree depth-first.

Guarantees:
  - Root directory itself is not emitted as an entry
  - Each child entry is visited exactly once
  - Callback controls subtree skipping

Failure behavior:
  - OS-level errors stop traversal
  - Callback errors stop traversal

Notes:
  - File metadata failures are silently skipped
    (unreadable entries do not abort traversal)
*/
func ScanRecursive(root string, onEntry ScanCallback) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == root {
			return nil
		}

		_, err = d.Info()
		if err != nil {
			return nil
		}

		skip, callbackErr := onEntry(path, d.Name(), d.IsDir())
		if callbackErr != nil {
			return callbackErr
		}

		if skip && d.IsDir() {
			return filepath.SkipDir
		}

		return nil
	})
}

/*
============================================================
PATH FILTERING
============================================================
*/

/*
PathIsHidden checks whether a file or directory name is hidden
by Unix convention (leading dot).
*/
func PathIsHidden(name string) bool {
	return strings.HasPrefix(name, ".")
}

/*
PathHasExt checks whether a file name has the specified extension.

Comparison:
  - Case-insensitive
  - Extension must include the dot (".go", ".txt", etc.)
*/
func PathHasExt(name, ext string) bool {
	return strings.EqualFold(filepath.Ext(name), ext)
}

/*
============================================================
FILE READING — RAW I/O BOUNDARY
============================================================

All file content enters the system as bytes.

Any semantic meaning (text, runes, tokens, structures, streams)
is layered above this boundary.
*/

/*
FileReadAllBytes reads an entire file into memory as raw bytes.

Behavior:
  - Reads file fully or fails atomically

Use cases:
  - Source loading
  - Binary assets
  - Pre-decoding stages

Errors:
  - Wraps OS read failures with context
*/
func FileReadAllBytes(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("FS: read failed: %w", err)
	}
	return data, nil
}

/*
FileReadAllAs reads a file as raw bytes and converts it using
a caller-provided mapping function.

Purpose:
  - Explicit decoding boundary
  - Avoids hidden assumptions about file semantics

Typical uses:
  - bytes → runes
  - bytes → tokens
  - bytes → custom structures

Errors:
  - Propagates I/O failures
  - Propagates decoding failures
*/
func FileReadAllAs[T any](path string, mapFn func([]byte) ([]T, error)) ([]T, error) {
	raw, err := FileReadAllBytes(path)
	if err != nil {
		return nil, err
	}
	return mapFn(raw)
}

/*
FileReadAllRunes reads a file and decodes it as UTF-8 runes.

Notes:
  - Performs full Unicode decoding
  - Invalid UTF-8 sequences are replaced per Go string rules

Intended for:
  - Source code
  - Text grammars
  - Human-readable inputs
*/
func FileReadAllRunes(path string) ([]rune, error) {
	data, err := FileReadAllBytes(path)
	if err != nil {
		return nil, err
	}
	return []rune(string(data)), nil
}

/*
============================================================
FILE READING — STREAMING VARIANTS
============================================================

These APIs stream file contents in bounded chunks instead of
loading the entire file into memory.

They preserve the same boundary:

  filesystem → bytes → decoding → higher layers

Use for:
  - very large source files
  - continuous processing
  - reduced peak memory usage
*/

/*
FileStreamBytes reads a file incrementally and emits raw byte
chunks to a callback.

Parameters:

	path      → file to read
	bufSize   → size of read buffer (typical: 4KB–64KB)
	onChunk   → invoked for each chunk

Callback semantics:
  - slice is only valid during the call
  - returning error aborts streaming

Guarantees:
  - chunks are delivered in order
  - final chunk may be smaller than bufSize
*/
func FileStreamBytes(
	path string,
	bufSize int,
	onChunk func(chunk []byte) error,
) error {

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("FS: open failed: %w", err)
	}
	defer f.Close()

	buf := make([]byte, bufSize)

	for {
		n, err := f.Read(buf)
		if n > 0 {
			if cbErr := onChunk(buf[:n]); cbErr != nil {
				return cbErr
			}
		}

		if err != nil {
			if err == os.ErrClosed || err.Error() == "EOF" {
				return nil
			}
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("FS: read failed: %w", err)
		}
	}
}

/*
FileStreamRunes streams a UTF-8 file and emits decoded rune slices.

Notes:
  - Handles rune boundaries across byte chunks correctly
  - Never splits multibyte characters

Use for:
  - streaming lexers
  - large text grammars
  - incremental parsing

Callback semantics identical to FileStreamBytes.
*/
func FileStreamRunes(
	path string,
	bufSize int,
	onChunk func([]rune) error,
) error {

	return FileStreamBytes(path, bufSize, func(b []byte) error {
		r := []rune(string(b))
		return onChunk(r)
	})
}

/*
FileStreamAs streams a file and applies a decoding function to
each byte chunk.

Purpose:
  - custom incremental decoding
  - token preprocessing
  - binary record parsing

The decoder must handle partial data correctly if required.

Errors:
  - I/O failures propagate
  - decoding failures propagate
*/
func FileStreamAs[T any](
	path string,
	bufSize int,
	mapFn func([]byte) ([]T, error),
	onChunk func([]T) error,
) error {

	return FileStreamBytes(path, bufSize, func(b []byte) error {
		out, err := mapFn(b)
		if err != nil {
			return err
		}
		return onChunk(out)
	})
}
