//go:build linux

package benchhost

import (
	"encoding/binary"
	"unsafe"

	"golang.org/x/sys/unix"
)

/*
BenchhostHWCounterSession holds open perf_event fds for one timed region.

Use cases:
- Measuring cycles, instructions, and cache misses around benchmark work.

Edge cases:
- Close must be called to release fds.
*/
type BenchhostHWCounterSession struct {
	fds []int
}

func perfOpenHardware(config uint64) (int, error) {
	attr := unix.PerfEventAttr{
		Type:   unix.PERF_TYPE_HARDWARE,
		Size:   uint32(unsafe.Sizeof(unix.PerfEventAttr{})),
		Config: config,
		Bits:   unix.PerfBitDisabled | unix.PerfBitExcludeKernel | unix.PerfBitExcludeHv,
	}
	return unix.PerfEventOpen(&attr, 0, -1, -1, 0)
}

func perfIOC(fd int, req uint) error {
	return unix.IoctlSetInt(fd, req, 0)
}

func perfReadCount(fd int) (uint64, error) {
	var buf [8]byte
	n, err := unix.Read(fd, buf[:])
	if err != nil {
		return 0, err
	}
	if n != 8 {
		return 0, unix.EIO
	}
	return binary.LittleEndian.Uint64(buf[:]), nil
}

/*
BenchhostHWCounterOpen tries to open three hardware counters for the calling thread.

Returns nil and a short human-readable reason on failure (EPERM, ENOENT, etc.).
*/
func BenchhostHWCounterOpen() (*BenchhostHWCounterSession, string) {
	configs := []uint64{
		unix.PERF_COUNT_HW_CPU_CYCLES,
		unix.PERF_COUNT_HW_INSTRUCTIONS,
		unix.PERF_COUNT_HW_CACHE_MISSES,
	}
	var fds []int
	for _, cfg := range configs {
		fd, err := perfOpenHardware(cfg)
		if err != nil {
			for _, old := range fds {
				_ = unix.Close(old)
			}
			return nil, "unavailable: " + err.Error()
		}
		fds = append(fds, fd)
	}
	return &BenchhostHWCounterSession{fds: fds}, ""
}

/*
BenchhostHWCounterResetEnable resets and enables all counters in the session.
*/
func BenchhostHWCounterResetEnable(s *BenchhostHWCounterSession) {
	if s == nil {
		return
	}
	for _, fd := range s.fds {
		_ = perfIOC(fd, uint(unix.PERF_EVENT_IOC_RESET))
	}
	for _, fd := range s.fds {
		_ = perfIOC(fd, uint(unix.PERF_EVENT_IOC_ENABLE))
	}
}

/*
BenchhostHWCounterDisableRead disables counters and reads delta counts.

Returns ok false if any step fails.
*/
func BenchhostHWCounterDisableRead(s *BenchhostHWCounterSession) (cycles, instructions, cacheMisses uint64, ok bool) {
	if s == nil || len(s.fds) != 3 {
		return 0, 0, 0, false
	}
	for _, fd := range s.fds {
		if err := perfIOC(fd, uint(unix.PERF_EVENT_IOC_DISABLE)); err != nil {
			return 0, 0, 0, false
		}
	}
	var vals [3]uint64
	for i, fd := range s.fds {
		v, err := perfReadCount(fd)
		if err != nil {
			return 0, 0, 0, false
		}
		vals[i] = v
	}
	return vals[0], vals[1], vals[2], true
}

/*
BenchhostHWCounterClose closes all perf fds.
*/
func BenchhostHWCounterClose(s *BenchhostHWCounterSession) {
	if s == nil {
		return
	}
	for _, fd := range s.fds {
		_ = unix.Close(fd)
	}
	s.fds = nil
}
