package database

import (
    "log"
    "user-team-asset-management/internal/models"
    
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func Connect(databaseURL string) *gorm.DB {
    db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    
    // Auto migrate
    err = db.AutoMigrate(
        &models.User{},
        &models.Team{},
        &models.TeamManager{},
        &models.TeamMember{},
        &models.Folder{},
        &models.Note{},
        &models.FolderShare{},
        &models.NoteShare{},
    )
    if err != nil {
        log.Fatal("Failed to migrate database:", err)
    }
    
    return db
}