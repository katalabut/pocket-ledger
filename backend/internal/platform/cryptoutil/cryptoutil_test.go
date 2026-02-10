package cryptoutil

import "testing"

func TestRandomDigits_Length(t *testing.T) {
	code, err := RandomDigits(6)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(code) != 6 {
		t.Fatalf("expected len 6, got %d (%q)", len(code), code)
	}
}
