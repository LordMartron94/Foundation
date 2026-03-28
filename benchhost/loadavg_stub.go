//go:build !unix

package benchhost

/*
BenchhostReadLoadAvg is unavailable on this platform.
*/
func BenchhostReadLoadAvg() (load1, load5, load15 float64, ok bool) {
	return 0, 0, 0, false
}
