package scan

import (
	"math"
	"testing"
)

func TestCatromWeights(t *testing.T) {
	var weights [4]float64
	count := 0
	for x := 0.0; x < 1.0; x += 0.0001 {
		catromWeights(x, weights[:])
		sum := weights[0] + weights[1] + weights[2] + weights[3]
		if math.Abs(sum-1.0) > 0.0001 {
			t.Errorf("bad weight sum %f, %v", sum, weights)
		}
		count++
	}
	t.Logf("%d catrom weights ok", count)
}
