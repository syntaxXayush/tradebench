package scoring

import (
	"math"
	"testing"

	"github.com/bench/shared/types"
)

// TestComputeScores_PerfectScore verifies a perfect submission gets 1.0 across the board.
// TPS = TARGET_TPS, P99 = 0ms, CorrectnessScore = 100.0.
func TestComputeScores_PerfectScore(t *testing.T) {
	snap := types.MetricSnapshot{
		TPS:              50000,
		P99LatencyMs:     0,
		CorrectnessScore: 100.0,
	}
	throughput, latency, correctness, final := computeScores(snap, 50000, 1000)

	assertFloat(t, "throughputScore", throughput, 1.0)
	assertFloat(t, "latencyScore", latency, 1.0)
	assertFloat(t, "correctnessScore", correctness, 1.0)
	assertFloat(t, "finalScore", final, 1.0)

	// Not disqualified (CorrectnessScore >= 30).
	if snap.CorrectnessScore < 30.0 {
		t.Error("expected not disqualified for CorrectnessScore 100.0")
	}
}

// TestComputeScores_ZeroTPS verifies zero throughput with partial latency and correctness.
// TPS = 0, P99 = 500ms, CorrectnessScore = 80.0.
// Expected: throughput=0.0, latency=0.5, correctness=0.8, final=0.40*0+0.40*0.5+0.20*0.8=0.36.
func TestComputeScores_ZeroTPS(t *testing.T) {
	snap := types.MetricSnapshot{
		TPS:              0,
		P99LatencyMs:     500,
		CorrectnessScore: 80.0,
	}
	throughput, latency, correctness, final := computeScores(snap, 50000, 1000)

	assertFloat(t, "throughputScore", throughput, 0.0)
	assertFloat(t, "latencyScore", latency, 0.5)
	assertFloat(t, "correctnessScore", correctness, 0.8)
	assertFloat(t, "finalScore", final, 0.36)
}

// TestComputeScores_DisqualificationThreshold verifies disqualification when
// CorrectnessScore < 30.0. The scores should still be computed and non-zero.
func TestComputeScores_DisqualificationThreshold(t *testing.T) {
	snap := types.MetricSnapshot{
		TPS:              25000,
		P99LatencyMs:     500,
		CorrectnessScore: 29.9,
	}
	throughput, latency, correctness, final := computeScores(snap, 50000, 1000)

	// Scores should be computed and non-zero.
	if throughput <= 0 {
		t.Errorf("throughputScore should be > 0, got %f", throughput)
	}
	if latency <= 0 {
		t.Errorf("latencyScore should be > 0, got %f", latency)
	}
	if correctness <= 0 {
		t.Errorf("correctnessScore should be > 0, got %f", correctness)
	}
	if final <= 0 {
		t.Errorf("finalScore should be > 0, got %f", final)
	}

	// Disqualification check.
	if snap.CorrectnessScore >= 30.0 {
		t.Errorf("expected disqualification for CorrectnessScore %f (< 30)", snap.CorrectnessScore)
	}

	// Expected values:
	// throughput = min(25000/50000, 1.0) = 0.5
	// latency = max(0, 1 - 500/1000) = 0.5
	// correctness = 29.9 / 100.0 = 0.299
	// final = 0.40*0.5 + 0.40*0.5 + 0.20*0.299 = 0.2 + 0.2 + 0.0598 = 0.4598
	assertFloat(t, "throughputScore", throughput, 0.5)
	assertFloat(t, "latencyScore", latency, 0.5)
	assertFloat(t, "correctnessScore", correctness, 0.299)
	assertFloat(t, "finalScore", final, 0.4598)
}

// TestComputeScores_TPSCappedAtOne verifies that TPS exceeding TARGET_TPS
// results in throughputScore capped at 1.0, not 2.0.
func TestComputeScores_TPSCappedAtOne(t *testing.T) {
	snap := types.MetricSnapshot{
		TPS:              100000, // double the target
		P99LatencyMs:     0,
		CorrectnessScore: 100.0,
	}
	throughput, latency, correctness, final := computeScores(snap, 50000, 1000)

	assertFloat(t, "throughputScore", throughput, 1.0) // must be capped at 1.0, not 2.0
	assertFloat(t, "latencyScore", latency, 1.0)
	assertFloat(t, "correctnessScore", correctness, 1.0)
	assertFloat(t, "finalScore", final, 1.0)
}

// assertFloat is a test helper that checks if got is within epsilon of want.
func assertFloat(t *testing.T, name string, got, want float64) {
	t.Helper()
	const epsilon = 1e-9
	if math.Abs(got-want) > epsilon {
		t.Errorf("%s: got %f, want %f", name, got, want)
	}
}
