package logic

import (
	"context"
	"testing"
)

func TestCurrentAdminIDFallback(t *testing.T) {
	if got := currentAdminID(context.Background()); got != 0 {
		t.Fatalf("expected default admin id 0, got %d", got)
	}
}

func TestCurrentAdminIDFromContext(t *testing.T) {
	ctx := WithCurrentAdminID(context.Background(), 99)

	if got := currentAdminID(ctx); got != 99 {
		t.Fatalf("expected admin id 99, got %d", got)
	}
}

func TestParseCurrentAdminID(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  uint64
	}{
		{name: "valid", input: "7", want: 7},
		{name: "trimmed", input: " 8 ", want: 8},
		{name: "zero", input: "0", want: 0},
		{name: "invalid", input: "abc", want: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ParseCurrentAdminID(tc.input); got != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, got)
			}
		})
	}
}
