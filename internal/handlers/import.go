package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"sync"
	"user-team-asset-management/internal/models"
	"user-team-asset-management/internal/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type ImportHandler struct {
	DB        *gorm.DB
	JWTSecret string
}

type ImportResult struct {
	TotalUsers   int      `json:"totalUsers"`
	SuccessCount int      `json:"successCount"`
	FailureCount int      `json:"failureCount"`
	Errors       []string `json:"errors"`
}

type UserRow struct {
	Username string
	Email    string
	Password string
	Role     string
	RowNum   int
}

func (h *ImportHandler) ImportUsers(c *gin.Context) {
	// Check if user is manager
	role := c.GetString("role")
	if role != "manager" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Manager role required"})
		return
	}

	// Parse multipart form
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// Parse CSV
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CSV file"})
		return
	}

	if len(records) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Empty CSV file"})
		return
	}

	// Skip header row
	if len(records) > 0 {
		records = records[1:]
	}

	// Parse user data
	var userRows []UserRow
	for i, record := range records {
		if len(record) < 4 {
			continue // Skip invalid rows
		}
		userRows = append(userRows, UserRow{
			Username: record[0],
			Email:    record[1],
			Password: record[2],
			Role:     record[3],
			RowNum:   i + 2, // +2 because we skipped header and arrays are 0-indexed
		})
	}

	// Process users with goroutines
	result := h.processUsersWithWorkerPool(userRows, 5) // 5 workers

	c.JSON(http.StatusOK, result)
}

func (h *ImportHandler) processUsersWithWorkerPool(userRows []UserRow, numWorkers int) ImportResult {
	jobs := make(chan UserRow, len(userRows))
	results := make(chan ProcessResult, len(userRows))

	// Start workers
	var wg sync.WaitGroup
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go h.worker(jobs, results, &wg)
	}

	// Send jobs
	for _, userRow := range userRows {
		jobs <- userRow
	}
	close(jobs)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var errors []string
	successCount := 0
	failureCount := 0

	for result := range results {
		if result.Success {
			successCount++
		} else {
			failureCount++
			errors = append(errors, fmt.Sprintf("Row %d: %s", result.RowNum, result.Error))
		}
	}

	return ImportResult{
		TotalUsers:   len(userRows),
		SuccessCount: successCount,
		FailureCount: failureCount,
		Errors:       errors,
	}
}

type ProcessResult struct {
	Success bool
	Error   string
	RowNum  int
}

func (h *ImportHandler) worker(jobs <-chan UserRow, results chan<- ProcessResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for userRow := range jobs {
		result := h.createUserFromRow(userRow)
		results <- result
	}
}

func (h *ImportHandler) createUserFromRow(userRow UserRow) ProcessResult {
	// Validate role
	if userRow.Role != "manager" && userRow.Role != "member" {
		return ProcessResult{
			Success: false,
			Error:   "invalid role, must be 'manager' or 'member'",
			RowNum:  userRow.RowNum,
		}
	}

	// Check if email already exists
	var existingUser models.User
	if err := h.DB.Where("email = ?", userRow.Email).First(&existingUser).Error; err == nil {
		return ProcessResult{
			Success: false,
			Error:   "email already exists",
			RowNum:  userRow.RowNum,
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userRow.Password), bcrypt.DefaultCost)
	if err != nil {
		return ProcessResult{
			Success: false,
			Error:   "failed to hash password",
			RowNum:  userRow.RowNum,
		}
	}

	// Create user
	user := models.User{
		ID:           utils.GenerateID(),
		Username:     userRow.Username,
		Email:        userRow.Email,
		PasswordHash: string(hashedPassword),
		Role:         userRow.Role,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		return ProcessResult{
			Success: false,
			Error:   "failed to create user in database",
			RowNum:  userRow.RowNum,
		}
	}

	return ProcessResult{
		Success: true,
		RowNum:  userRow.RowNum,
	}
}
