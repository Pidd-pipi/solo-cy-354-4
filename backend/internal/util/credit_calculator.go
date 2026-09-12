package util

import "github.com/lp/campus-market/internal/constants"

// CreditDelta returns the credit score delta for a review rating.
func CreditDelta(rating string) int {
	switch rating {
	case constants.ReviewRatingGood:
		return 5
	case constants.ReviewRatingMedium:
		return 0
	case constants.ReviewRatingBad:
		return -10
	default:
		return 0
	}
}

// ClampCredit keeps the credit score inside [0, 300].
func ClampCredit(score int) int {
	if score < 0 {
		return 0
	}
	if score > 300 {
		return 300
	}
	return score
}
