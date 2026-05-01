//go:build !linux

package benchhost

/*
BenchhostHWCounterSession is a stub on non-Linux platforms.
*/
type BenchhostHWCounterSession struct{}

/*
BenchhostHWCounterOpen returns nil and an unavailable reason on non-Linux.
*/
func BenchhostHWCounterOpen() (*BenchhostHWCounterSession, string) {
	return nil, "unavailable: not linux"
}

/*
BenchhostHWCounterResetEnable is a no-op on non-Linux.
*/
func BenchhostHWCounterResetEnable(_ *BenchhostHWCounterSession) {}

/*
BenchhostHWCounterDisableRead returns ok false on non-Linux.
*/
func BenchhostHWCounterDisableRead(_ *BenchhostHWCounterSession) (cycles, instructions, cacheMisses uint64, ok bool) {
	return 0, 0, 0, false
}

/*
BenchhostHWCounterClose is a no-op on non-Linux.
*/
func BenchhostHWCounterClose(_ *BenchhostHWCounterSession) {}
