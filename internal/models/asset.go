package models

import "time"

type Folder struct {
    ID        string    `json:"folderId" gorm:"primaryKey"`
    Name      string    `json:"name" gorm:"not null"`
    OwnerID   string    `json:"ownerId" gorm:"not null"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
    
    Owner  User   `json:"owner" gorm:"foreignKey:OwnerID"`
    Notes  []Note `json:"notes" gorm:"foreignKey:FolderID"`
    Shares []FolderShare `json:"shares" gorm:"foreignKey:FolderID"`
}

type Note struct {
    ID        string    `json:"noteId" gorm:"primaryKey"`
    Title     string    `json:"title" gorm:"not null"`
    Body      string    `json:"body"`
    FolderID  string    `json:"folderId" gorm:"not null"`
    OwnerID   string    `json:"ownerId" gorm:"not null"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
    
    Owner  User        `json:"owner" gorm:"foreignKey:OwnerID"`
    Folder Folder      `json:"folder" gorm:"foreignKey:FolderID"`
    Shares []NoteShare `json:"shares" gorm:"foreignKey:NoteID"`
}

type FolderShare struct {
    FolderID string `json:"folderId" gorm:"primaryKey"`
    UserID   string `json:"userId" gorm:"primaryKey"`
    Access   string `json:"access" gorm:"not null;check:access IN ('read','write')"`
    CreatedAt time.Time `json:"createdAt"`
}

type NoteShare struct {
    NoteID   string `json:"noteId" gorm:"primaryKey"`
    UserID   string `json:"userId" gorm:"primaryKey"`
    Access   string `json:"access" gorm:"not null;check:access IN ('read','write')"`
    CreatedAt time.Time `json:"createdAt"`
}