package correctness

import (
	"context"

	"github.com/bench/shared/types"
)

// Validator compares ActualFill vs ExpectedFill using the C++ reference engine.
// Day 5: will invoke the compiled ref-engine binary and compare outputs.
// Today: returns 100.0 (fully correct) for all inputs as a stub.
type Validator struct {
	refEnginePath string
}

// NewValidator creates a new Validator with the path to the compiled C++ reference binary.
func NewValidator(refEnginePath string) *Validator {
	return &Validator{
		refEnginePath: refEnginePath,
	}
}

// Validate runs correctness validation on the given BotEvents.
// Returns a score on the 0–100 scale representing the percentage of correct fills.
func (v *Validator) Validate(ctx context.Context, events []types.BotEvent) (float64, error) {
	// TODO(Day 5): invoke ref-engine binary, compare ActualFill vs ExpectedFill, return correctness score 0–100
	return 100.0, nil
}
