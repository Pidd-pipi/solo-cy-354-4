package util

import (
	"testing"

	"github.com/lp/campus-market/internal/constants"
)

func TestCreditDelta(t *testing.T) {
	tests := []struct {
		rating string
		want   int
	}{
		{rating: constants.ReviewRatingGood, want: 5},
		{rating: constants.ReviewRatingMedium, want: 0},
		{rating: constants.ReviewRatingBad, want: -10},
		{rating: "unknown", want: 0},
	}
	for _, tt := range tests {
		if got := CreditDelta(tt.rating); got != tt.want {
			t.Fatalf("CreditDelta(%s) = %d, want %d", tt.rating, got, tt.want)
		}
	}
}

func TestClampCredit(t *testing.T) {
	tests := []struct {
		in   int
		want int
	}{
		{in: -5, want: 0},
		{in: 150, want: 150},
		{in: 500, want: 300},
	}
	for _, tt := range tests {
		if got := ClampCredit(tt.in); got != tt.want {
			t.Fatalf("ClampCredit(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}
