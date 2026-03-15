package models

import "time"

type Tree struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ParkID    uint      `gorm:"column:park_id;not null;index" json:"park_id"`
	TreeCode  string    `gorm:"column:tree_code;size:64;not null" json:"tree_code"`
	Species   string    `gorm:"column:species;size:128;default:''" json:"species"`
	Status    int       `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	Park Park `gorm:"foreignKey:ParkID;references:ID" json:"park,omitempty"`
}

func (Tree) TableName() string {
	return "trees"
}
