package models

type User struct {
	ID           uint
	Username     string
	PasswordHash string
	Status       int // 1=正常 0=禁用
}
