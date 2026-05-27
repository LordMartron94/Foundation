//go:build unix && !linux

package benchhost

import (
	"encoding/binary"

	"golang.org/x/sys/unix"
)

/*
BenchhostReadLoadAvg returns the 1/5/15 minute load averages on BSD-like Unix targets (darwin,
FreeBSD, OpenBSD, NetBSD, DragonflyBSD) by reading the kernel's `vm.loadavg` sysctl.

[Context]
The `vm.loadavg` sysctl marshals the kernel's `struct loadavg`:

	struct loadavg {
	    fixpt_t ldavg[3];   // three fixpt_t (uint32) load averages
	    long    fscale;     // scaling divisor; size matches C long
	};

`long` is 4 bytes on 32-bit ABIs and 8 bytes on 64-bit ABIs, and ldavg is followed by alignment
padding before fscale on 64-bit targets. We size-detect the trailing scalar to stay portable
without per-OS or per-ARCH variants.

[Returns]
load1, load5, load15 - load averages as floating-point values.
ok                   - false when the sysctl is unavailable, returns an unexpected layout, or the

	reported scale is zero.
*/
func BenchhostReadLoadAvg() (load1, load5, load15 float64, ok bool) {
	buf, err := unix.SysctlRaw("vm.loadavg")
	if err != nil || len(buf) < 16 {
		return 0, 0, 0, false
	}
	ldavg := [3]uint32{
		binary.NativeEndian.Uint32(buf[0:4]),
		binary.NativeEndian.Uint32(buf[4:8]),
		binary.NativeEndian.Uint32(buf[8:12]),
	}

	var fscale uint64
	switch {
	case len(buf) >= 20:
		fscale = binary.NativeEndian.Uint64(buf[len(buf)-8:])
	case len(buf) >= 16:
		fscale = uint64(binary.NativeEndian.Uint32(buf[len(buf)-4:]))
	default:
		return 0, 0, 0, false
	}
	if fscale == 0 {
		return 0, 0, 0, false
	}

	scale := float64(fscale)
	return float64(ldavg[0]) / scale, float64(ldavg[1]) / scale, float64(ldavg[2]) / scale, true
}
