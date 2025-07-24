package models

import "time"

type Team struct {
    ID        string    `json:"teamId" gorm:"primaryKey"`
    TeamName  string    `json:"teamName" gorm:"not null"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
    
    Managers []User `json:"managers" gorm:"many2many:team_managers;"`
    Members  []User `json:"members" gorm:"many2many:team_members;"`
}

type TeamManager struct {
    TeamID string `gorm:"primaryKey"`
    UserID string `gorm:"primaryKey"`
}

type TeamMember struct {
    TeamID string `gorm:"primaryKey"`
    UserID string `gorm:"primaryKey"`
}