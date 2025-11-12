package hash

import (
	"fmt"
	foundationtesting "foundation/testing"
	"testing"

	"golang.org/x/sys/cpu"
)

func TestHash(t *testing.T) {
	foundationtesting.EnableOkMessagesSet(true)

	data := randomBytes(1024, 42)
	hasher := XXH3HasherCreateWithSeed(42)

	cpu.X86.HasAVX2 = false
	XXH3HasherReinitialize()
	gotScalar := XXH3HasherHash64(hasher, data)

	cpu.X86.HasAVX2 = true
	XXH3HasherReinitialize()
	gotAVX2 := XXH3HasherHash64(hasher, data)

	success := fmt.Sprintf("Hash is the same: %x", gotScalar)
	err := fmt.Sprintf("Hash mismatch: scalar=%x,avx2=%x", gotScalar, gotAVX2)

	foundationtesting.Assert(gotScalar == gotAVX2, err, success, t)

	foundationtesting.EnableOkMessagesSet(false)
}
