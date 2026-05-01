//go:build unix && !linux

package benchhost

import "golang.org/x/sys/unix"

/*
BenchhostReadLoadAvg uses getloadavg on BSD-like Unix.

Returns ok false when the syscall is unavailable or fails.
*/
func BenchhostReadLoadAvg() (load1, load5, load15 float64, ok bool) {
	var buf [3]float64
	n, err := unix.Getloadavg(buf[:])
	if err != nil || n != 3 {
		return 0, 0, 0, false
	}
	return buf[0], buf[1], buf[2], true
}
