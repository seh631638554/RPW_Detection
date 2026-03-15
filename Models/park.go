package models

import "time"

type Park struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name;size:128;uniqueIndex;not null" json:"name"`
	Code      string    `gorm:"column:code;size:64;uniqueIndex;not null" json:"code"`
	Location  string    `gorm:"column:location;size:255;default:''" json:"location"`
	Status    int       `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (Park) TableName() string {
	return "parks"
}
