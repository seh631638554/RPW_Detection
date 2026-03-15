package models

import "time"

type Device struct {
	ID         uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	DeviceCode string    `gorm:"column:device_code;size:64;uniqueIndex;not null" json:"device_code"`
	Name       string    `gorm:"column:name;size:128;not null" json:"name"`
	ParkID     uint      `gorm:"column:park_id;not null;index" json:"park_id"`
	TreeID     *uint     `gorm:"column:tree_id;index" json:"tree_id,omitempty"`
	Status     int       `gorm:"column:status;not null;default:1" json:"status"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`

	Park Park  `gorm:"foreignKey:ParkID;references:ID" json:"park,omitempty"`
	Tree *Tree `gorm:"foreignKey:TreeID;references:ID" json:"tree,omitempty"`
}

func (Device) TableName() string {
	return "devices"
}
