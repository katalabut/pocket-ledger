package domain

import "testing"

func TestValidateSplits_Empty(t *testing.T) {
	if err := ValidateSplits(1000, nil); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateSplits_Match(t *testing.T) {
	splits := []Split{
		{AmountMinor: 600},
		{AmountMinor: 400},
	}
	if err := ValidateSplits(1000, splits); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateSplits_Mismatch(t *testing.T) {
	splits := []Split{
		{AmountMinor: 600},
		{AmountMinor: 300},
	}
	if err := ValidateSplits(1000, splits); err != ErrSplitMismatch {
		t.Fatalf("expected ErrSplitMismatch, got %v", err)
	}
}

func TestValidateSplits_NegativeAmounts(t *testing.T) {
	splits := []Split{
		{AmountMinor: -600},
		{AmountMinor: -400},
	}
	if err := ValidateSplits(-1000, splits); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidAccountType(t *testing.T) {
	if !ValidAccountType("card") {
		t.Fatal("card should be valid")
	}
	if ValidAccountType("invalid") {
		t.Fatal("invalid should not be valid")
	}
}
