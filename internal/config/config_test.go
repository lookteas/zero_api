package config

import "testing"

func TestDefaultAwarenessCycleStartDate(t *testing.T) {
	var c Config
	if c.AwarenessCycle.StartDate != "" {
		t.Fatalf("zero-value config should not set awareness cycle start date, got %q", c.AwarenessCycle.StartDate)
	}
}
