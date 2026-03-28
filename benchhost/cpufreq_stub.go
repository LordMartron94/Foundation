//go:build !linux

package benchhost

/*
BenchhostReadCPUMHzNominal is not implemented on this platform.
*/
func BenchhostReadCPUMHzNominal() float64 {
	return 0
}
