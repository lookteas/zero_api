package logic

import (
	"context"
	"testing"
)

func TestCurrentUserIDFallback(t *testing.T) {
	if got := currentUserID(context.Background()); got != defaultCurrentUserID {
		t.Fatalf("expected default user id %d, got %d", defaultCurrentUserID, got)
	}
}

func TestCurrentUserIDFromContext(t *testing.T) {
	ctx := WithCurrentUserID(context.Background(), 42)

	if got := currentUserID(ctx); got != 42 {
		t.Fatalf("expected user id 42, got %d", got)
	}
}

func TestParseCurrentUserID(t *testing.T) {
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
			if got := ParseCurrentUserID(tc.input); got != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, got)
			}
		})
	}
}
