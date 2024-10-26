package models

import (
	"gorm.io/gorm"
)

type User struct {
	*gorm.Model
	Name           string
	Email          string `gorm:"unique"`
	Password       string
	UserIdentifier string `gorm:"unique"`
	BucketName     string `gorm:"unique"`
	DeleteConsent  string
	MobileNumber   uint64 `gorm:"unique"`
}
