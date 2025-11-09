package hash

import "math/rand"

func randomBytes(n int, seed int64) []byte {
	r := rand.New(rand.NewSource(seed))
	data := make([]byte, n)
	_, _ = r.Read(data)
	return data
}
