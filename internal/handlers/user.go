package handlers

import (
    "net/http"
    "user-team-asset-management/internal/models"
    
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type UserHandler struct {
    DB *gorm.DB
}

func (h *UserHandler) GetProfile(c *gin.Context) {
    userID := c.GetString("userID")
    
    var user models.User
    if err := h.DB.Where("id = ?", userID).First(&user).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }
    
    c.JSON(http.StatusOK, user)
}

func (h *UserHandler) GetUserTeams(c *gin.Context) {
    userID := c.GetString("userID")
    
    // Get teams where user is manager
    var managerTeams []models.Team
    h.DB.Joins("JOIN team_managers ON teams.id = team_managers.team_id").
        Where("team_managers.user_id = ?", userID).
        Find(&managerTeams)
    
    // Get teams where user is member
    var memberTeams []models.Team
    h.DB.Joins("JOIN team_members ON teams.id = team_members.team_id").
        Where("team_members.user_id = ?", userID).
        Find(&memberTeams)
    
    c.JSON(http.StatusOK, gin.H{
        "managerTeams": managerTeams,
        "memberTeams":  memberTeams,
    })
}