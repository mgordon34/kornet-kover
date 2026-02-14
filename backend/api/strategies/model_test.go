package strategies

import "testing"

func TestStrategyConstants(t *testing.T) {
	if ValueComparison == "" || FunctionComparison == "" || ModifiedComparison == "" {
		t.Fatalf("comparison constants should be non-empty")
	}
	if Multiply == "" || Divide == "" || Add == "" || Subtract == "" {
		t.Fatalf("modifier operator constants should be non-empty")
	}
	if GreaterThan == "" || LessThan == "" || GreaterOrEqual == "" || LessOrEqual == "" || Equal == "" {
		t.Fatalf("comparison operator constants should be non-empty")
	}
}
