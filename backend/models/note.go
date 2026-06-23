package models

// Note corresponds to the 'notes' table in the database
type Note struct {
	NoteID        uint        `gorm:"primaryKey;column:note_id" json:"note_id"`
	SellerID   uint        `gorm:"not null" json:"seller_id"`
	Title         string      `gorm:"not null" json:"title"`
	Description   string      `gorm:"column:description" json:"desc"`
	Price         float64     `gorm:"not null;default:0.00" json:"price"`
	OverallRating float64     `gorm:"default:0.00" json:"overall_rating"`
	RatingCount   int64       `gorm:"-" json:"rating_count"`
	Media         []NoteMedia `gorm:"foreignKey:NoteID" json:"media,omitempty"`
	Tags          []Tag       `gorm:"many2many:note_tags;foreignKey:NoteID;joinForeignKey:NoteID;References:TagID;joinReferences:TagID" json:"tags,omitempty"`
	SellerName    string      `gorm:"-" json:"seller_name"`
	Status        string      `gorm:"default:DRAFT" json:"status"`
	IsClassic     bool        `gorm:"-" json:"is_classic"`
}
