package models

type Tag struct {
	TagID   uint   `gorm:"primaryKey;column:tag_id" json:"tag_id"`
	TagName string `gorm:"unique;not null;column:tag_name" json:"tag_name"`
}

type NoteTag struct {
	NoteID uint `gorm:"primaryKey;column:note_id" json:"note_id"`
	TagID  uint `gorm:"primaryKey;column:tag_id" json:"tag_id"`
}

type NoteMedia struct {
	MediaID   uint   `gorm:"primaryKey;column:media_id" json:"media_id"`
	NoteID    uint   `gorm:"not null;column:note_id" json:"note_id"`
	FileURL   string `gorm:"not null;column:file_url" json:"file_url"`
	MediaType string `gorm:"default:media;column:media_type" json:"media_type"`
}
