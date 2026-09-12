package constants

// TradeStatus defines trade order state machine values shared with the frontend.
const (
	TradeStatusPending   = "pending"
	TradeStatusConfirmed = "confirmed"
	TradeStatusCompleted = "completed"
	TradeStatusCancelled = "cancelled"
)

// TradeStatuses lists all valid trade statuses in flow order.
var TradeStatuses = []string{
	TradeStatusPending, TradeStatusConfirmed, TradeStatusCompleted, TradeStatusCancelled,
}

// IsTradeStatus reports whether the given status is valid.
func IsTradeStatus(s string) bool {
	for _, v := range TradeStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// TradeStatusText returns the Chinese label of a trade status.
func TradeStatusText(s string) string {
	switch s {
	case TradeStatusPending:
		return "待确认"
	case TradeStatusConfirmed:
		return "已确认"
	case TradeStatusCompleted:
		return "已完成"
	case TradeStatusCancelled:
		return "已取消"
	default:
		return "未知"
	}
}

// ReviewRating defines review rating enum values shared with the frontend.
const (
	ReviewRatingGood   = "good"
	ReviewRatingMedium = "medium"
	ReviewRatingBad    = "bad"
)

// ReviewRatingText returns the Chinese label of a review rating.
func ReviewRatingText(r string) string {
	switch r {
	case ReviewRatingGood:
		return "好评"
	case ReviewRatingMedium:
		return "中评"
	case ReviewRatingBad:
		return "差评"
	default:
		return "未知"
	}
}
