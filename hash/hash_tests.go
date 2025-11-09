package hash

import (
	"testing"
)

func TestHash(t *testing.T) {
	data := randomBytes(1024, 0)

	hasher := XXH3HasherCreateWithSeed(0)

	_ = XXH3HasherHash64(hasher, data)
}
