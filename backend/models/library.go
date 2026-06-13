package models

import "time"

// NoteLicense corresponds to the 'note_licenses' table (Owned Notes)
type NoteLicense struct {
	LicenseID         uint      `gorm:"primaryKey;column:license_id" json:"license_id"`
	NoteID            uint      `gorm:"not null" json:"note_id"`
	UserID            uint      `gorm:"not null" json:"user_id"`
	TransactionItemID uint      `gorm:"not null" json:"transaction_item_id"`
	AcquiredDate      time.Time `gorm:"autoCreateTime" json:"acquired_date"`
	Status            string    `gorm:"default:ACTIVE" json:"status"`

	// Preload relationship to fetch the actual Note details
	Note Note `gorm:"foreignKey:NoteID;references:NoteID" json:"note"`
}

// WishList corresponds to the 'wish_lists' table
type WishList struct {
	WishlistID uint      `gorm:"primaryKey;column:wishlist_id" json:"wishlist_id"`
	UserID     uint      `gorm:"not null" json:"user_id"`
	NoteID     uint      `gorm:"not null" json:"note_id"`
	AddedAt    time.Time `gorm:"autoCreateTime" json:"added_at"`

	// Preload relationship to fetch the actual Note details
	Note Note `gorm:"foreignKey:NoteID;references:NoteID" json:"note"`
}
