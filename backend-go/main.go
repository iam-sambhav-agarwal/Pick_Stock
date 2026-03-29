package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"stock-analysis-api/handlers"
	"stock-analysis-api/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// MongoDB connection
	mongoURL := getEnv("MONGO_URL", "mongodb://localhost:27017")
	dbName := getEnv("DB_NAME", "stock_analysis")

	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURL))
	if err != nil {
		log.Fatal("Failed to create MongoDB client:", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	// Test MongoDB connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}
	log.Println("Successfully connected to MongoDB")

	db := client.Database(dbName)

	// Initialize services
	yahooService := services.NewYahooFinanceService()
	analysisService := services.NewAnalysisService()
	stockHandler := handlers.NewStockHandler(yahooService, analysisService, db)

	// Setup Gin router
	router := gin.Default()

	// CORS middleware
	corsOrigins := getEnv("CORS_ORIGINS", "*")
	config := cors.DefaultConfig()
	if corsOrigins == "*" {
		config.AllowAllOrigins = true
	} else {
		// Split multiple origins by comma and trim whitespace
		origins := []string{}
		for _, origin := range strings.Split(corsOrigins, ",") {
			origins = append(origins, strings.TrimSpace(origin))
		}
		config.AllowOrigins = origins
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "Access-Control-Allow-Origin"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour
	router.Use(cors.New(config))

	// API routes with /api prefix
	api := router.Group("/api")
	{
		api.GET("/", stockHandler.Root)
		api.GET("/stocks/popular", stockHandler.GetPopularStocks)
		api.POST("/analyze", stockHandler.AnalyzeStock)
		api.GET("/history/:symbol", stockHandler.GetAnalysisHistory)
	}

	// Serve static files (React frontend)
	staticDir := filepath.Join(".", "..", "frontend", "build")
	if _, err := os.Stat(staticDir); err == nil {
		log.Printf("Frontend build found at: %s", staticDir)
		log.Println("Serving static files")

		// Serve static assets
		router.Use(static.Serve("/static", static.LocalFile(filepath.Join(staticDir, "static"), false)))

		// Serve specific files
		router.StaticFile("/favicon.ico", filepath.Join(staticDir, "favicon.ico"))
		router.StaticFile("/manifest.json", filepath.Join(staticDir, "manifest.json"))
		router.StaticFile("/logo192.png", filepath.Join(staticDir, "logo192.png"))
		router.StaticFile("/logo512.png", filepath.Join(staticDir, "logo512.png"))

		// Catch-all route for React Router (must be last)
		router.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path

			// Skip API routes
			if len(path) >= 4 && path[:4] == "/api" {
				c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
				return
			}

			// Try to serve specific file if it exists
			filePath := filepath.Join(staticDir, path)
			if _, err := os.Stat(filePath); err == nil {
				c.File(filePath)
				return
			}

			// Otherwise serve index.html for React Router
			indexPath := filepath.Join(staticDir, "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				c.File(indexPath)
			} else {
				c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			}
		})
	} else {
		log.Printf("Frontend build not found at: %s", staticDir)
		log.Println("Only API endpoints will work")
		log.Println("To build frontend: cd ../frontend && npm install && npm run build")

		// Root route when no frontend - provide helpful information
		router.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message":         "Stock Analysis API - Frontend Build Required",
				"status":          "Backend Ready ✅",
				"frontend_status": "Not Built ⚠️",
				"instructions": map[string]interface{}{
					"build_frontend": []string{
						"cd ../frontend",
						"npm install",
						"npm run build",
						"cd ../backend-go && ./stock-analysis-api",
					},
					"api_endpoints": map[string]string{
						"popular_stocks": "/api/stocks/popular",
						"analyze_stock":  "POST /api/analyze",
						"stock_history":  "/api/history/{symbol}",
						"health_check":   "/health",
					},
				},
				"integration_guide": "/FRONTEND_BACKEND_INTEGRATION.md",
			})
		})

		// Add helpful endpoints for development
		router.GET("/integration", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"title": "Integration Status",
				"backend": map[string]interface{}{
					"status":  "Running ✅",
					"port":    getEnv("PORT", "8001"),
					"mongodb": "Connected ✅",
					"cors":    "Enabled ✅",
				},
				"frontend": map[string]interface{}{
					"status":     "Not Built ⚠️",
					"build_path": staticDir,
					"required_steps": []string{
						"Install Node.js (https://nodejs.org/)",
						"cd ../frontend",
						"npm install",
						"npm run build",
						"Restart this server",
					},
				},
				"endpoints_ready": []string{
					"GET /api/stocks/popular",
					"POST /api/analyze",
					"GET /api/history/{symbol}",
					"GET /health",
				},
			})
		})
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Get port from environment
	portStr := getEnv("PORT", "8001")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Printf("Invalid PORT value: %s, using default 8001", portStr)
		port = 8001
	}

	log.Printf("Starting server on port %d", port)
	if err := router.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
