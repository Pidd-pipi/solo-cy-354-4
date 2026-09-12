// Package model defines the GORM entity structures for campus-market.
package model

import "time"

// User represents a campus student or admin.
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Phone        string    `gorm:"size:20;uniqueIndex;not null" json:"phone"`
	PasswordHash string    `gorm:"size:100;not null" json:"-"`
	Nickname     string    `gorm:"size:32;not null" json:"nickname"`
	Avatar       string    `gorm:"size:255" json:"avatar"`
	Role         string    `gorm:"size:16;not null;default:student" json:"role"`
	Campus       string    `gorm:"size:64" json:"campus"`
	CreditScore  int       `gorm:"not null;default:100" json:"credit_score"`
	CreatedAt    time.Time `json:"created_at"`
}
