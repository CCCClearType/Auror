package models

import "time"

// ShoppingCart corresponds to the 'shopping_carts' table
type ShoppingCart struct {
	CartID  uint      `gorm:"primaryKey;column:cart_id" json:"cart_id"`
	UserID  uint      `gorm:"not null" json:"user_id"`
	NoteID  uint      `gorm:"not null" json:"note_id"`
	AddedAt time.Time `gorm:"autoCreateTime" json:"added_at"`

	// Preload relationship to fetch the actual Note
	Note Note `gorm:"foreignKey:NoteID;references:NoteID" json:"note"`
}
