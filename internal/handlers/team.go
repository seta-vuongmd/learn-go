package handlers

import (
	"net/http"
	"user-team-asset-management/internal/models"
	"user-team-asset-management/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TeamHandler struct {
	DB *gorm.DB
}

type CreateTeamRequest struct {
	TeamName string `json:"teamName" binding:"required"`
	Managers []struct {
		ManagerID   string `json:"managerId"`
		ManagerName string `json:"managerName"`
	} `json:"managers"`
	Members []struct {
		MemberID   string `json:"memberId"`
		MemberName string `json:"memberName"`
	} `json:"members"`
}

func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")
	teamID := utils.GenerateID()

	team := models.Team{
		ID:       teamID,
		TeamName: req.TeamName,
	}

	if err := h.DB.Create(&team).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team"})
		return
	}

	// Add creator as manager
	h.DB.Create(&models.TeamManager{TeamID: teamID, UserID: userID})

	// Add other managers
	for _, manager := range req.Managers {
		h.DB.Create(&models.TeamManager{TeamID: teamID, UserID: manager.ManagerID})
	}

	// Add members
	for _, member := range req.Members {
		h.DB.Create(&models.TeamMember{TeamID: teamID, UserID: member.MemberID})
	}

	c.JSON(http.StatusCreated, team)
}

func (h *TeamHandler) AddMember(c *gin.Context) {
	teamID := c.Param("teamId")
	var req struct {
		MemberID string `json:"memberId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")
	if !h.isTeamManager(userID, teamID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to manage this team"})
		return
	}

	teamMember := models.TeamMember{
		TeamID: teamID,
		UserID: req.MemberID,
	}

	if err := h.DB.Create(&teamMember).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member added successfully"})
}

func (h *TeamHandler) RemoveMember(c *gin.Context) {
	teamID := c.Param("teamId")
	memberID := c.Param("memberId")
	userID := c.GetString("userID")

	if !h.isTeamManager(userID, teamID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to manage this team"})
		return
	}

	if err := h.DB.Where("team_id = ? AND user_id = ?", teamID, memberID).Delete(&models.TeamMember{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed successfully"})
}

func (h *TeamHandler) isTeamManager(userID, teamID string) bool {
	var count int64
	h.DB.Model(&models.TeamManager{}).Where("team_id = ? AND user_id = ?", teamID, userID).Count(&count)
	return count > 0
}
