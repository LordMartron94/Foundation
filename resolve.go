package foundation

import (
	"os"
	"path/filepath"
)

type SearchRoot struct {
	Name string
	Path string
}

func ResolveFile(
	relativePath string,
	roots []SearchRoot,
) (resolvedPath string, root SearchRoot, ok bool) {

	for _, r := range roots {
		p := filepath.Join(r.Path, relativePath)
		if _, err := os.Stat(p); err == nil {
			return p, r, true
		}
	}

	return "", SearchRoot{}, false
}
