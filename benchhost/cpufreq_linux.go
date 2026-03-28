//go:build linux

package benchhost

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

/*
BenchhostReadCPUMHzNominal returns a rough CPU frequency in MHz from /proc/cpuinfo.

Use cases:
- Context for timing and HW counter interpretation.

Edge cases:
- Returns 0 when missing (common in some VMs).
*/
func BenchhostReadCPUMHzNominal() float64 {
	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return 0
	}
	defer f.Close()

	var maxMHz float64
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if !strings.HasPrefix(line, "cpu MHz") && !strings.HasPrefix(line, "model name") {
			continue
		}
		if strings.HasPrefix(line, "cpu MHz") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			v, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if err != nil {
				continue
			}
			if v > maxMHz {
				maxMHz = v
			}
		}
	}
	if maxMHz > 0 {
		return maxMHz
	}

	// cpufreq scaling_cur_freq is in kHz
	data, err := os.ReadFile("/sys/devices/system/cpu/cpu0/cpufreq/scaling_cur_freq")
	if err != nil {
		return 0
	}
	khz, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
	if err != nil || khz <= 0 {
		return 0
	}
	return khz / 1000.0
}
