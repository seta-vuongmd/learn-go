package models

import (
    "time"
)

type User struct {
    ID           string    `json:"userId" gorm:"primaryKey"`
    Username     string    `json:"username" gorm:"not null"`
    Email        string    `json:"email" gorm:"uniqueIndex;not null"`
    PasswordHash string    `json:"-" gorm:"not null"`
    Role         string    `json:"role" gorm:"not null;check:role IN ('manager','member')"`
    CreatedAt    time.Time `json:"createdAt"`
    UpdatedAt    time.Time `json:"updatedAt"`
}

func (User) TableName() string {
    return "users"
}