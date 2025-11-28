package system

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func FileWriteEmpty(path string) error {
	return os.WriteFile(path, []byte{}, 0644)
}

func FileRename(oldPath, newPath string) error {
	if FileExists(newPath) {
		return fmt.Errorf("FS: destination already exists")
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("FS: rename failed: %w", err)
	}
	return nil
}

func FileDelete(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("FS: delete failed: %w", err)
	}
	return nil
}

func PathJoin(elem ...string) string {
	return filepath.Join(elem...)
}

func PathBase(path string) string {
	return filepath.Base(path)
}

func PathDir(path string) string {
	return filepath.Dir(path)
}

func PathExt(path string) string {
	return filepath.Ext(path)
}

func DirCreate(path string) error {
	return os.Mkdir(path, 0755)
}

// ScanCallback is the function signature the Logic layer must implement.
// path: absolute path
// name: file/folder name
// isDir: true if directory
// Returns: true to skip this item/directory, error if processing failed
type ScanCallback func(path string, name string, isDir bool) (bool, error)

// ScanRecursive walks the directory tree.
// It swallows OS-specific types (DirEntry) and passes simple types to the callback.
func ScanRecursive(root string, onEntry ScanCallback) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // Stop on OS error
		}

		// Don't process the root folder itself as an entry
		if path == root {
			return nil
		}

		_, err = d.Info()
		if err != nil {
			return nil // Skip files we can't read info for
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

// PathIsHidden checks if the file starts with a dot
func PathIsHidden(name string) bool {
	return strings.HasPrefix(name, ".")
}

// PathHasExt checks file extension
func PathHasExt(name, ext string) bool {
	return strings.EqualFold(filepath.Ext(name), ext)
}
