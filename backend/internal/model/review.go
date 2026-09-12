package model

import "time"

// Review is a credit rating left after a completed trade.
type Review struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	TradeID    uint      `gorm:"uniqueIndex:uniq_review_trade_reviewer,priority:1;not null" json:"trade_id"`
	ReviewerID uint      `gorm:"uniqueIndex:uniq_review_trade_reviewer,priority:2;not null" json:"reviewer_id"`
	RevieweeID uint      `gorm:"index;not null" json:"reviewee_id"`
	Rating     string    `gorm:"size:16;not null" json:"rating"`
	Content    string    `gorm:"type:text" json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}
