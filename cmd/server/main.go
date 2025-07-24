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
    
    r := gin.Default()
    
    // GraphQL endpoint
    r.POST("/graphql", gin.WrapH(graphqlHandler))
    r.GET("/graphql", gin.WrapH(graphqlHandler))
    
    // Protected REST API routes
    api := r.Group("/api")
    api.Use(middleware.AuthMiddleware(cfg.JWTSecret))
    {
        // Team management (managers only)
        teams := api.Group("/teams")
        teams.Use(middleware.RequireManager())
        {
            teams.POST("", teamHandler.CreateTeam)
            teams.POST("/:teamId/members", teamHandler.AddMember)
            teams.DELETE("/:teamId/members/:memberId", teamHandler.RemoveMember)
            teams.GET("/:teamId/assets", assetHandler.GetTeamAssets)
        }
        
        // Asset management
        api.POST("/folders", assetHandler.CreateFolder)
        api.POST("/folders/:folderId/notes", assetHandler.CreateNote)
        api.POST("/folders/:folderId/share", assetHandler.ShareFolder)
    }
    
    log.Printf("Server starting on port %s", cfg.Port)
    r.Run(":" + cfg.Port)
}