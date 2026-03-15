package models

type User struct {
	ID           uint   `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Username     string `gorm:"column:username;size:64;uniqueIndex;not null" json:"username"`
	PasswordHash string `gorm:"column:password_hash;size:255;not null" json:"-"`
	Status       int    `gorm:"column:status;not null;default:1" json:"status"`
	IsAdmin      int    `gorm:"column:is_admin;not null;default:0" json:"is_admin"`
}

func (User) TableName() string {
	return "users"
}
