package main

import "testing"

func TestStripDatabaseFromDSN(t *testing.T) {
	input := "root:root123@tcp(127.0.0.1:3309)/zero_app?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"
	want := "root:root123@tcp(127.0.0.1:3309)/"

	if got := stripDatabaseFromDSN(input); got != want {
		t.Fatalf("stripDatabaseFromDSN() = %q, want %q", got, want)
	}
}
