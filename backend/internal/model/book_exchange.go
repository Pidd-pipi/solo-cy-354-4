package model

import "time"

// BookExchange is a book swap request posted by a student.
type BookExchange struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"user_id"`
	OfferBook   string    `gorm:"size:64;not null" json:"offer_book"`
	WantBook    string    `gorm:"size:64;not null" json:"want_book"`
	Description string    `gorm:"type:text" json:"description"`
	Status      string    `gorm:"size:16;index;not null;default:open" json:"status"`
	MatchedID   *uint     `json:"matched_id"`
	CreatedAt   time.Time `json:"created_at"`
}
