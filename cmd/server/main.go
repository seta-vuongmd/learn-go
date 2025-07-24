package main

import (
	"log"
	"user-team-asset-management/internal/config"
	"user-team-asset-management/internal/database"
	"user-team-asset-management/internal/graphql"
	"user-team-asset-management/internal/handlers"
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

	r := gin.Default()

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

		// Team management (managers only)
		teams := api.Group("/teams")
		teams.Use(middleware.RequireManager())
		{
			teams.POST("", teamHandler.CreateTeam)
			teams.POST("/:teamId/members", teamHandler.AddMember)
			teams.DELETE("/:teamId/members/:memberId", teamHandler.RemoveMember)
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
