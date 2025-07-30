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

func (h *TeamHandler) AddManager(c *gin.Context) {
	teamID := c.Param("teamId")
	var req struct {
		ManagerID string `json:"managerId" binding:"required"`
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

	// Check if user is already a manager
	if h.isTeamManager(req.ManagerID, teamID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is already a manager"})
		return
	}

	teamManager := models.TeamManager{
		TeamID: teamID,
		UserID: req.ManagerID,
	}

	if err := h.DB.Create(&teamManager).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add manager"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Manager added successfully"})
}

func (h *TeamHandler) RemoveManager(c *gin.Context) {
	teamID := c.Param("teamId")
	managerID := c.Param("managerId")
	userID := c.GetString("userID")

	if !h.isTeamManager(userID, teamID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to manage this team"})
		return
	}

	// Prevent removing the last manager
	var managerCount int64
	h.DB.Model(&models.TeamManager{}).Where("team_id = ?", teamID).Count(&managerCount)
	if managerCount <= 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove the last manager"})
		return
	}

	if err := h.DB.Where("team_id = ? AND user_id = ?", teamID, managerID).Delete(&models.TeamManager{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove manager"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Manager removed successfully"})
}

func (h *TeamHandler) isTeamManager(userID, teamID string) bool {
	var count int64
	h.DB.Model(&models.TeamManager{}).Where("team_id = ? AND user_id = ?", teamID, userID).Count(&count)
	return count > 0
}

func (h *TeamHandler) GetTeam(c *gin.Context) {
	teamID := c.Param("teamId")
	userID := c.GetString("userID")

	if !h.isTeamMember(userID, teamID) && !h.isTeamManager(userID, teamID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not a team member"})
		return
	}

	var team models.Team
	if err := h.DB.Where("id = ?", teamID).First(&team).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	// Get managers
	var managers []models.User
	h.DB.Joins("JOIN team_managers ON users.id = team_managers.user_id").
		Where("team_managers.team_id = ?", teamID).
		Find(&managers)

	// Get members
	var members []models.User
	h.DB.Joins("JOIN team_members ON users.id = team_members.user_id").
		Where("team_members.team_id = ?", teamID).
		Find(&members)

	c.JSON(http.StatusOK, gin.H{
		"team":     team,
		"managers": managers,
		"members":  members,
	})
}

func (h *TeamHandler) isTeamMember(userID, teamID string) bool {
	var count int64
	h.DB.Model(&models.TeamMember{}).Where("team_id = ? AND user_id = ?", teamID, userID).Count(&count)
	return count > 0
}

func (h *TeamHandler) GetAllTeams(c *gin.Context) {
	userRole := c.GetString("role")

	// Chỉ manager mới có thể xem tất cả teams
	if userRole != "manager" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Manager role required"})
		return
	}

	var teams []models.Team
	if err := h.DB.Find(&teams).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch teams"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"teams": teams})
}

func (h *TeamHandler) SearchTeams(c *gin.Context) {
	teamName := c.Query("name")
	userID := c.GetString("userID")

	var teams []models.Team
	query := h.DB.Model(&models.Team{})

	// Filter by name if provided
	if teamName != "" {
		query = query.Where("team_name ILIKE ?", "%"+teamName+"%")
	}

	// Only show teams where user is member or manager
	query = query.Where(`
		id IN (
			SELECT team_id FROM team_managers WHERE user_id = ?
			UNION
			SELECT team_id FROM team_members WHERE user_id = ?
		)
	`, userID, userID)

	if err := query.Find(&teams).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search teams"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"teams": teams})
}
