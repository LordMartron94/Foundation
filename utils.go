package foundation

import (
	"reflect"
	"runtime"
)

/*
GetFunctionName returns the name of a function.

Pass the pointer to the function in.
*/
func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

/*
Overhead represents the calculated relationship between real-world wall time
and the summed CPU/execution time of a task.
*/
type Overhead struct {
	WallNS           float64
	SummedNS         float64
	AbsoluteLostNS   float64
	LostPercentage   float64
	EffectiveWorkers float64
}

/*
ComputeOverheadNS computes overhead metrics based on wall time and summed time.
*/
func ComputeOverheadNS(wallNS, summedNS float64) Overhead {
	if wallNS <= 0 {
		return Overhead{SummedNS: summedNS}
	}

	return Overhead{
		WallNS:           wallNS,
		SummedNS:         summedNS,
		AbsoluteLostNS:   computeLostTime(wallNS, summedNS),
		LostPercentage:   computeLostPercentage(wallNS, summedNS),
		EffectiveWorkers: computeConcurrencyMultiplier(wallNS, summedNS),
	}
}

func computeLostTime(wall, summed float64) float64 {
	diff := wall - summed
	if diff < 0 {
		return 0
	}
	return diff
}

func computeLostPercentage(wall, summed float64) float64 {
	if wall <= summed {
		return 0
	}
	return ((wall - summed) / wall) * 100
}

func computeConcurrencyMultiplier(wall, summed float64) float64 {
	if summed <= wall {
		return 1.0
	}
	return summed / wall
}
