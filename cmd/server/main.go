package main

import (
	"log"
	"user-team-asset-management/internal/config"
	"user-team-asset-management/internal/database"
	"user-team-asset-management/internal/graphql"
	"user-team-asset-management/internal/handlers"
	"user-team-asset-management/internal/logger"
	"user-team-asset-management/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/graphql-go/handler"
)

func main() {
	cfg := config.Load()
	db := database.Connect(cfg.DatabaseURL)

	// GraphQL setup
	resolver := &graphql.Resolver{DB: db, JWTSecret: cfg.JWTSecret}
	schema, err := resolver.CreateSchema()
	if err != nil {
		log.Fatal("Failed to create GraphQL schema:", err)
	}

	graphqlHandler := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	// REST API setup
	teamHandler := &handlers.TeamHandler{DB: db}
	assetHandler := &handlers.AssetHandler{DB: db}
	userHandler := &handlers.UserHandler{DB: db}
	importHandler := &handlers.ImportHandler{DB: db, JWTSecret: cfg.JWTSecret}

	r := gin.Default()

	// Add logging middleware
	r.Use(logger.GinLogger())
	r.Use(logger.Recovery())

	// GraphQL endpoint
	r.POST("/graphql", gin.WrapH(graphqlHandler))
	r.GET("/graphql", gin.WrapH(graphqlHandler))

	// Protected REST API routes
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		// User routes
		api.GET("/profile", userHandler.GetProfile)
		api.GET("/my-teams", userHandler.GetUserTeams)
		api.GET("/my-folders", assetHandler.GetUserFolders)

		// Team routes
		api.GET("/teams", teamHandler.SearchTeams) // NEW: Search teams
		api.GET("/teams/:teamId", teamHandler.GetTeam)
		api.GET("/teams/:teamId/assets", assetHandler.GetTeamAssets)

		// Asset routes
		api.GET("/folders/:folderId", assetHandler.GetFolder)
		api.PUT("/folders/:folderId", assetHandler.UpdateFolder)
		api.DELETE("/folders/:folderId", assetHandler.DeleteFolder)
		api.GET("/notes/:noteId", assetHandler.GetNote)
		api.PUT("/notes/:noteId", assetHandler.UpdateNote)
		api.DELETE("/notes/:noteId", assetHandler.DeleteNote)

		// Sharing routes
		api.DELETE("/folders/:folderId/share/:userId", assetHandler.RevokeFolderShare)
		api.POST("/notes/:noteId/share", assetHandler.ShareNote)
		api.DELETE("/notes/:noteId/share/:userId", assetHandler.RevokeNoteShare)

		// Manager-only routes
		api.GET("/users/:userId/assets", assetHandler.GetUserAssets)
		api.POST("/import-users", importHandler.ImportUsers)

		// Team management (managers only)
		teams := api.Group("/teams")
		teams.Use(middleware.RequireManager())
		{
			teams.POST("", teamHandler.CreateTeam)
			teams.POST("/:teamId/members", teamHandler.AddMember)
			teams.DELETE("/:teamId/members/:memberId", teamHandler.RemoveMember)
			teams.POST("/:teamId/managers", teamHandler.AddManager)
			teams.DELETE("/:teamId/managers/:managerId", teamHandler.RemoveManager)
			teams.GET("/all", teamHandler.GetAllTeams) // NEW: Get all teams (manager only)
		}

		// Asset management
		api.POST("/folders", assetHandler.CreateFolder)
		api.POST("/folders/:folderId/notes", assetHandler.CreateNote)
		api.POST("/folders/:folderId/share", assetHandler.ShareFolder)
	}

	log.Printf("Server starting on port %s", cfg.Port)
	r.Run(":" + cfg.Port)
}
