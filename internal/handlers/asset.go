package handlers

import (
	"net/http"
	"user-team-asset-management/internal/models"
	"user-team-asset-management/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AssetHandler struct {
	DB *gorm.DB
}

func (h *AssetHandler) CreateFolder(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")
	folder := models.Folder{
		ID:      utils.GenerateID(),
		Name:    req.Name,
		OwnerID: userID,
	}

	if err := h.DB.Create(&folder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create folder"})
		return
	}

	c.JSON(http.StatusCreated, folder)
}

func (h *AssetHandler) CreateNote(c *gin.Context) {
	folderID := c.Param("folderId")
	var req struct {
		Title string `json:"title" binding:"required"`
		Body  string `json:"body"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")

	if !h.canWriteToFolder(userID, folderID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "No write access to this folder"})
		return
	}

	note := models.Note{
		ID:       utils.GenerateID(),
		Title:    req.Title,
		Body:     req.Body,
		FolderID: folderID,
		OwnerID:  userID,
	}

	if err := h.DB.Create(&note).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create note"})
		return
	}

	c.JSON(http.StatusCreated, note)
}

func (h *AssetHandler) ShareFolder(c *gin.Context) {
	folderID := c.Param("folderId")
	var req struct {
		UserID string `json:"userId" binding:"required"`
		Access string `json:"access" binding:"required,oneof=read write"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("userID")
	if !h.ownsFolder(userID, folderID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only folder owner can share"})
		return
	}

	share := models.FolderShare{
		FolderID: folderID,
		UserID:   req.UserID,
		Access:   req.Access,
	}

	if err := h.DB.Create(&share).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share folder"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder shared successfully"})
}

func (h *AssetHandler) GetTeamAssets(c *gin.Context) {
	teamID := c.Param("teamId")
	userID := c.GetString("userID")

	if !h.isTeamManager(userID, teamID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Manager access required"})
		return
	}

	// Get all team members
	var memberIDs []string
	h.DB.Model(&models.TeamMember{}).Where("team_id = ?", teamID).Pluck("user_id", &memberIDs)

	// Get folders owned by team members
	var folders []models.Folder
	h.DB.Preload("Notes").Where("owner_id IN ?", memberIDs).Find(&folders)

	// Get shared folders accessible by team members
	var sharedFolders []models.Folder
	h.DB.Preload("Notes").
		Joins("JOIN folder_shares ON folders.id = folder_shares.folder_id").
		Where("folder_shares.user_id IN ?", memberIDs).
		Find(&sharedFolders)

	c.JSON(http.StatusOK, gin.H{
		"ownedFolders":  folders,
		"sharedFolders": sharedFolders,
	})
}

func (h *AssetHandler) ownsFolder(userID, folderID string) bool {
	var count int64
	h.DB.Model(&models.Folder{}).Where("id = ? AND owner_id = ?", folderID, userID).Count(&count)
	return count > 0
}

func (h *AssetHandler) canWriteToFolder(userID, folderID string) bool {
	if h.ownsFolder(userID, folderID) {
		return true
	}

	var count int64
	h.DB.Model(&models.FolderShare{}).
		Where("folder_id = ? AND user_id = ? AND access = 'write'", folderID, userID).
		Count(&count)
	return count > 0
}

func (h *AssetHandler) isTeamManager(userID, teamID string) bool {
	var count int64
	h.DB.Model(&models.TeamManager{}).Where("team_id = ? AND user_id = ?", teamID, userID).Count(&count)
	return count > 0
}

func (h *AssetHandler) GetUserFolders(c *gin.Context) {
	userID := c.GetString("userID")

	// Get owned folders
	var ownedFolders []models.Folder
	h.DB.Preload("Notes").Where("owner_id = ?", userID).Find(&ownedFolders)

	// Get shared folders
	var sharedFolders []models.Folder
	h.DB.Preload("Notes").
		Joins("JOIN folder_shares ON folders.id = folder_shares.folder_id").
		Where("folder_shares.user_id = ?", userID).
		Find(&sharedFolders)

	c.JSON(http.StatusOK, gin.H{
		"ownedFolders":  ownedFolders,
		"sharedFolders": sharedFolders,
	})
}

func (h *AssetHandler) GetFolder(c *gin.Context) {
	folderID := c.Param("folderId")
	userID := c.GetString("userID")

	if !h.canReadFolder(userID, folderID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "No access to this folder"})
		return
	}

	var folder models.Folder
	if err := h.DB.Preload("Notes").Where("id = ?", folderID).First(&folder).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}

	c.JSON(http.StatusOK, folder)
}

func (h *AssetHandler) canReadFolder(userID, folderID string) bool {
	// Check if user owns the folder
	var count int64
	h.DB.Model(&models.Folder{}).Where("id = ? AND owner_id = ?", folderID, userID).Count(&count)
	if count > 0 {
		return true
	}

	// Check if folder is shared with user
	h.DB.Model(&models.FolderShare{}).Where("folder_id = ? AND user_id = ?", folderID, userID).Count(&count)
	return count > 0
}
