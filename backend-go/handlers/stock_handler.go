package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"stock-analysis-api/models"
	"stock-analysis-api/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// StockHandler handles stock-related HTTP requests
type StockHandler struct {
	yahooService    *services.YahooFinanceService
	analysisService *services.AnalysisService
	db              *mongo.Database
}

// NewStockHandler creates a new instance of StockHandler
func NewStockHandler(yahooService *services.YahooFinanceService, analysisService *services.AnalysisService, db *mongo.Database) *StockHandler {
	return &StockHandler{
		yahooService:    yahooService,
		analysisService: analysisService,
		db:              db,
	}
}

// Popular stocks list
var popularStocks = []models.PopularStock{
	{Symbol: "TCS", Name: "Tata Consultancy Services", Exchange: "NSE"},
	{Symbol: "RELIANCE", Name: "Reliance Industries", Exchange: "NSE"},
	{Symbol: "INFY", Name: "Infosys", Exchange: "NSE"},
	{Symbol: "HDFCBANK", Name: "HDFC Bank", Exchange: "NSE"},
	{Symbol: "ICICIBANK", Name: "ICICI Bank", Exchange: "NSE"},
	{Symbol: "HINDUNILVR", Name: "Hindustan Unilever", Exchange: "NSE"},
	{Symbol: "ITC", Name: "ITC Limited", Exchange: "NSE"},
	{Symbol: "SBIN", Name: "State Bank of India", Exchange: "NSE"},
	{Symbol: "BHARTIARTL", Name: "Bharti Airtel", Exchange: "NSE"},
	{Symbol: "WIPRO", Name: "Wipro", Exchange: "NSE"},
	{Symbol: "LT", Name: "Larsen & Toubro", Exchange: "NSE"},
	{Symbol: "ASIANPAINT", Name: "Asian Paints", Exchange: "NSE"},
	{Symbol: "AXISBANK", Name: "Axis Bank", Exchange: "NSE"},
	{Symbol: "MARUTI", Name: "Maruti Suzuki", Exchange: "NSE"},
	{Symbol: "SUNPHARMA", Name: "Sun Pharma", Exchange: "NSE"},
}

// Root handles the root endpoint
func (sh *StockHandler) Root(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Message: "Stock Analysis API",
		Version: "1.0.0",
	})
}

// GetPopularStocks returns the list of popular stocks
func (sh *StockHandler) GetPopularStocks(c *gin.Context) {
	c.JSON(http.StatusOK, popularStocks)
}

// AnalyzeStock analyzes a stock and provides buy/sell/hold recommendation
func (sh *StockHandler) AnalyzeStock(c *gin.Context) {
	var request models.StockAnalysisRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Detail: "Invalid request format: " + err.Error(),
		})
		return
	}

	// Set default exchange if not provided
	if request.Exchange == "" {
		request.Exchange = "NSE"
	}

	symbol := strings.ToUpper(strings.TrimSpace(request.Symbol))
	exchange := strings.ToUpper(request.Exchange)

	log.Printf("Analyzing stock: %s on %s", symbol, exchange)

	// Fetch stock data using multiple providers with fallback
	fundamentals, technical, err := sh.analysisService.GetStockData(symbol)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Detail: "Stock '" + symbol + "' not found or data unavailable: " + err.Error(),
		})
		return
	}

	// Generate analysis
	analysis := sh.analysisService.AnalyzeStock(fundamentals, technical)

	// Store analysis in database
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		analysisDoc := models.AnalysisDocument{
			StockAnalysisResponse: analysis,
			CreatedAt:             time.Now().UTC(),
		}

		collection := sh.db.Collection("stock_analyses")
		_, err := collection.InsertOne(ctx, analysisDoc)
		if err != nil {
			log.Printf("Error storing analysis: %v", err)
		}
	}()

	c.JSON(http.StatusOK, analysis)
}

// GetAnalysisHistory returns historical analyses for a stock
func (sh *StockHandler) GetAnalysisHistory(c *gin.Context) {
	symbol := strings.ToUpper(c.Param("symbol"))
	limitStr := c.DefaultQuery("limit", "10")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := sh.db.Collection("stock_analyses")

	// Create filter and options
	filter := bson.M{"symbol": symbol}
	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetLimit(int64(limit)).
		SetProjection(bson.M{"_id": 0}) // Exclude _id field

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		log.Printf("Error fetching history: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Detail: "Error fetching analysis history",
		})
		return
	}
	defer cursor.Close(ctx)

	var analyses []models.StockAnalysisResponse
	if err := cursor.All(ctx, &analyses); err != nil {
		log.Printf("Error decoding history: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Detail: "Error processing analysis history",
		})
		return
	}

	// Handle nil slice
	if analyses == nil {
		analyses = []models.StockAnalysisResponse{}
	}

	response := models.AnalysisHistory{
		Symbol:   symbol,
		Count:    len(analyses),
		Analyses: analyses,
	}

	c.JSON(http.StatusOK, response)
}
