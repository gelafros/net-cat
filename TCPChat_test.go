package main

import "testing"

func TestIsPort(t *testing.T) {
	if !isPort("8989") {
		t.Fatal("expected 8989 to be valid port digits")
	}
	if isPort("localhost") || isPort("-ui") || isPort("89a9") {
		t.Fatal("expected non-digit strings to return false")
	}
}
