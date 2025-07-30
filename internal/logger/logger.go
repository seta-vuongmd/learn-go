package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type Logger struct {
	*log.Logger
}

var DefaultLogger *Logger

func init() {
	DefaultLogger = NewLogger()
}

func NewLogger() *Logger {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatal("Failed to create logs directory:", err)
	}

	// Create log file with timestamp
	logFile := filepath.Join("logs", "app-"+time.Now().Format("2006-01-02")+".log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}

	// Write to both file and stdout
	multiWriter := io.MultiWriter(os.Stdout, file)
	logger := log.New(multiWriter, "", log.LstdFlags|log.Lshortfile)

	return &Logger{Logger: logger}
}

func (l *Logger) Info(v ...interface{}) {
	l.Printf("[INFO] %v\n", v...)
}

func (l *Logger) Error(v ...interface{}) {
	l.Printf("[ERROR] %v\n", v...)
}

func (l *Logger) Warning(v ...interface{}) {
	l.Printf("[WARNING] %v\n", v...)
}

func (l *Logger) Debug(v ...interface{}) {
	l.Printf("[DEBUG] %v\n", v...)
}

// GinLogger returns a gin.HandlerFunc (middleware) that logs requests using our custom logger.
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Log request
		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()

		DefaultLogger.Printf("[GIN] %s | %3d | %13v | %15s | %-7s %s",
			start.Format("2006/01/02 - 15:04:05"),
			status,
			latency,
			clientIP,
			method,
			path,
		)
	}
}

// Recovery returns a gin.HandlerFunc that recovers from panics and logs them.
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		DefaultLogger.Error("Panic recovered:", recovered)
		c.AbortWithStatus(500)
	})
}
