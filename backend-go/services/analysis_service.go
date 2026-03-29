package services

import (
	"fmt"
	"log"
	"math"
	"os"
	"stock-analysis-api/models"
	"strings"
	"time"
)

// DataProvider represents different financial data providers
type DataProvider interface {
	GetStockData(symbol string) (models.FundamentalsData, models.TechnicalData, error)
}

// AnalysisService handles stock analysis and recommendation generation
type AnalysisService struct {
	yahooService        *YahooFinanceService
	alphaVantageService *AlphaVantageService
	finnhubService      *FinnhubService
	nseIndiaService     *NSEIndiaService
}

// NewAnalysisService creates a new instance of AnalysisService
func NewAnalysisService() *AnalysisService {
	// Initialize all data providers
	alphaVantageKey := os.Getenv("ALPHA_VANTAGE_API_KEY")
	finnhubKey := os.Getenv("FINNHUB_API_KEY")

	return &AnalysisService{
		yahooService:        NewYahooFinanceService(),
		alphaVantageService: NewAlphaVantageService(alphaVantageKey),
		finnhubService:      NewFinnhubService(finnhubKey),
		nseIndiaService:     NewNSEIndiaService(),
	}
}

// GetStockData fetches stock data using multiple providers with fallback strategy
func (as *AnalysisService) GetStockData(symbol string) (models.FundamentalsData, models.TechnicalData, error) {
	var lastError error

	// For Indian stocks, prioritize NSE India service first
	isIndianStock := as.isIndianStock(symbol)
	if isIndianStock {
		// Strategy 1 for Indian stocks: Try NSE India service first
		log.Printf("Trying NSE India service for Indian stock %s", symbol)
		fundamentals, technical, err := as.nseIndiaService.GetStockData(symbol)
		if err != nil {
			// Check if it's a "stock not found" error
			if strings.Contains(strings.ToLower(err.Error()), "stock not found") || strings.Contains(strings.ToLower(err.Error()), "not a valid") {
				return models.FundamentalsData{}, models.TechnicalData{}, err
			}
			log.Printf("NSE India service failed for %s: %v", symbol, err)
		} else if fundamentals.CurrentPrice > 0 {
			log.Printf("Successfully fetched real-time data from NSE India service for %s", symbol)
			return fundamentals, technical, nil
		}

		// Strategy 2 for Indian stocks: Try Yahoo Finance as backup
		log.Printf("Trying Yahoo Finance for Indian stock %s", symbol)
		fundamentals = as.yahooService.GetStockFundamentals(symbol)
		technical = as.yahooService.GetStockTechnical(symbol, "NSE")

		if fundamentals.Error == "" && technical.Error == "" && fundamentals.CurrentPrice > 0 {
			log.Printf("Successfully fetched data from Yahoo Finance for %s", symbol)
			return fundamentals, technical, nil
		}
		log.Printf("Yahoo Finance failed for %s", symbol)
	}

	// Strategy 2: Try Alpha Vantage if API key is available
	if os.Getenv("ALPHA_VANTAGE_API_KEY") != "" {
		log.Printf("Trying Alpha Vantage for %s", symbol)
		fundamentals, technical, err := as.alphaVantageService.GetStockData(symbol)
		if err == nil {
			log.Printf("Successfully fetched data from Alpha Vantage for %s", symbol)
			return fundamentals, technical, nil
		}
		log.Printf("Alpha Vantage failed for %s: %v", symbol, err)
		lastError = err
	}

	// Strategy 3: Try Finnhub if API key is available (mainly for US stocks)
	if os.Getenv("FINNHUB_API_KEY") != "" && !isIndianStock {
		log.Printf("Trying Finnhub for %s", symbol)
		fundamentals, technical, err := as.finnhubService.GetStockData(symbol)
		if err == nil && fundamentals.CurrentPrice > 0 {
			log.Printf("Successfully fetched data from Finnhub for %s", symbol)
			return fundamentals, technical, nil
		}
		log.Printf("Finnhub failed for %s: %v", symbol, err)
		lastError = err
	}

	// Strategy 4: Try Yahoo Finance (for non-Indian stocks or as final fallback)
	if !isIndianStock {
		log.Printf("Trying Yahoo Finance for %s", symbol)
		fundamentals := as.yahooService.GetStockFundamentals(symbol)
		technical := as.yahooService.GetStockTechnical(symbol, "NSE")

		// Check if Yahoo Finance returned real data or mock data
		if fundamentals.Error == "" && technical.Error == "" {
			log.Printf("Successfully fetched data from Yahoo Finance for %s", symbol)
			return fundamentals, technical, nil
		}
	}

	// If all providers failed, check if it was due to invalid stock symbol
	if isIndianStock {
		// For Indian stocks, if NSE India service said "stock not found", don't fallback to mock data
		log.Printf("All Indian stock providers failed for %s", symbol)
		return models.FundamentalsData{}, models.TechnicalData{}, fmt.Errorf("stock not found: %s is not a valid Indian stock or data is unavailable", symbol)
	}

	// For US stocks, try mock data as final fallback
	log.Printf("All providers failed for %s, using mock data", symbol)
	fundamentals := as.yahooService.GetStockFundamentals(symbol)
	technical := as.yahooService.GetStockTechnical(symbol, "NSE")

	if fundamentals.Symbol != "" && technical.Symbol != "" {
		log.Printf("Using mock data for %s", symbol)
		return fundamentals, technical, nil
	}

	// If everything failed, return the last error
	if lastError != nil {
		return models.FundamentalsData{}, models.TechnicalData{}, lastError
	}

	return models.FundamentalsData{}, models.TechnicalData{}, fmt.Errorf("stock not found: %s is not a valid stock symbol or data is unavailable", symbol)
}

// AnalyzeStock generates comprehensive stock analysis with buy/sell/hold recommendation
func (as *AnalysisService) AnalyzeStock(fundamentals models.FundamentalsData, technical models.TechnicalData) models.StockAnalysisResponse {
	// Calculate support and resistance levels
	supportResistance := as.calculateSupportResistance(technical.Price)

	// Generate recommendation based on fundamentals and technicals
	recommendation := as.generateRecommendation(fundamentals, technical, supportResistance)

	// Calculate price targets and stop loss
	targets := as.calculateTargets(recommendation, technical.Price, supportResistance)

	return models.StockAnalysisResponse{
		Symbol:           fundamentals.Symbol,
		CompanyName:      fundamentals.CompanyName,
		CurrentPrice:     technical.Price,
		Recommendation:   recommendation.Action,
		Confidence:       recommendation.Confidence,
		Reason:           recommendation.Reason,
		TargetBuyPrice:   targets.BuyPrice,
		TargetSellPrice:  targets.SellPrice,
		StopLoss:         targets.StopLoss,
		SupportLevels:    supportResistance.Support,
		ResistanceLevels: supportResistance.Resistance,
		Fundamentals: models.FundamentalsData{
			PERatio:       fundamentals.PERatio,
			PBRatio:       fundamentals.PBRatio,
			ROE:           fundamentals.ROE,
			DividendYield: fundamentals.DividendYield,
			DebtToEquity:  fundamentals.DebtToEquity,
			MarketCap:     fundamentals.MarketCap,
		},
		Technical: models.TechnicalData{
			Price:         technical.Price,
			ChangePercent: technical.ChangePercent,
			Volume:        technical.Volume,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// calculateSupportResistance calculates support and resistance levels using price-based method
func (as *AnalysisService) calculateSupportResistance(currentPrice float64) models.SupportResistance {
	if currentPrice <= 0 {
		return models.SupportResistance{
			Support:    map[string]float64{"S1": 0, "S2": 0},
			Resistance: map[string]float64{"R1": 0, "R2": 0},
			Pivot:      0,
		}
	}

	// Simple calculation based on percentage levels
	s1 := roundToTwo(currentPrice * 0.97) // 3% below
	s2 := roundToTwo(currentPrice * 0.94) // 6% below
	r1 := roundToTwo(currentPrice * 1.03) // 3% above
	r2 := roundToTwo(currentPrice * 1.06) // 6% above
	pivot := roundToTwo((s1 + r1) / 2)

	return models.SupportResistance{
		Support:    map[string]float64{"S1": s1, "S2": s2},
		Resistance: map[string]float64{"R1": r1, "R2": r2},
		Pivot:      pivot,
	}
}

// generateRecommendation generates buy/sell/hold recommendation based on multiple factors
func (as *AnalysisService) generateRecommendation(fundamentals models.FundamentalsData, technical models.TechnicalData, srLevels models.SupportResistance) models.Recommendation {
	peRatio := fundamentals.PERatio
	roe := fundamentals.ROE
	debtToEquity := fundamentals.DebtToEquity
	price := technical.Price
	changePercent := technical.ChangePercent

	score := 0
	var reasons []string

	// PE Ratio analysis
	if peRatio > 0 && peRatio < 15 {
		score += 2
		reasons = append(reasons, "Undervalued (Low PE)")
	} else if peRatio >= 15 && peRatio < 25 {
		score += 1
		reasons = append(reasons, "Fairly valued PE")
	} else if peRatio >= 35 {
		score -= 2
		reasons = append(reasons, "Overvalued (High PE)")
	}

	// ROE analysis
	if roe > 15 {
		score += 1
		reasons = append(reasons, "Strong ROE")
	} else if roe < 10 && roe > 0 {
		score -= 1
		reasons = append(reasons, "Weak ROE")
	}

	// Debt analysis
	if debtToEquity >= 0 && debtToEquity < 0.5 {
		score += 1
		reasons = append(reasons, "Low debt")
	} else if debtToEquity > 2 {
		score -= 1
		reasons = append(reasons, "High debt")
	}

	// Price position relative to support/resistance
	s1 := srLevels.Support["S1"]
	r1 := srLevels.Resistance["R1"]

	if price > 0 {
		if price <= s1 {
			score += 1
			reasons = append(reasons, "Near support level")
		} else if price >= r1 {
			score -= 1
			reasons = append(reasons, "Near resistance level")
		}
	}

	// Recent momentum
	if changePercent < -3 {
		score += 1
		reasons = append(reasons, "Oversold")
	} else if changePercent > 5 {
		score -= 1
		reasons = append(reasons, "Overbought")
	}

	// Determine action and confidence
	var action, confidence string

	if score >= 3 {
		action = "BUY"
		confidence = "High"
	} else if score >= 1 {
		action = "BUY"
		confidence = "Medium"
	} else if score <= -3 {
		action = "SELL"
		confidence = "High"
	} else if score <= -1 {
		action = "SELL"
		confidence = "Medium"
	} else {
		action = "HOLD"
		confidence = "Medium"
	}

	reason := "Neutral signals"
	if len(reasons) > 0 {
		reason = strings.Join(reasons, ", ")
	}

	return models.Recommendation{
		Action:     action,
		Confidence: confidence,
		Reason:     reason,
		Score:      score,
	}
}

// calculateTargets calculates buy price, sell price, and stop loss
func (as *AnalysisService) calculateTargets(recommendation models.Recommendation, currentPrice float64, srLevels models.SupportResistance) models.PriceTargets {
	action := recommendation.Action
	s1 := srLevels.Support["S1"]
	s2 := srLevels.Support["S2"]
	r1 := srLevels.Resistance["R1"]

	var buyPrice, sellPrice, stopLoss float64

	switch action {
	case "BUY":
		buyPrice = roundToTwo(math.Min(currentPrice, s1))
		sellPrice = roundToTwo(r1)
		stopLoss = roundToTwo(s2)
	case "SELL":
		buyPrice = roundToTwo(s2)
		sellPrice = roundToTwo(currentPrice)
		stopLoss = roundToTwo(r1)
	default: // HOLD
		buyPrice = roundToTwo(s1)
		sellPrice = roundToTwo(r1)
		stopLoss = roundToTwo(s2)
	}

	return models.PriceTargets{
		BuyPrice:  math.Max(buyPrice, 0),
		SellPrice: math.Max(sellPrice, 0),
		StopLoss:  math.Max(stopLoss, 0),
	}
}

// roundToTwo rounds a float64 to 2 decimal places
func roundToTwo(num float64) float64 {
	return math.Round(num*100) / 100
}

// isIndianStock checks if a symbol is likely an Indian stock
func (as *AnalysisService) isIndianStock(symbol string) bool {
	// Check for US stock patterns first
	if strings.Contains(symbol, ".") || // US stocks often have periods
		len(symbol) <= 4 && strings.ToUpper(symbol) == symbol && !strings.ContainsAny(symbol, "0123456789") {
		// Check common US patterns
		usPatterns := []string{"AAPL", "MSFT", "GOOGL", "AMZN", "TSLA", "META", "NVDA", "NFLX"}
		for _, usStock := range usPatterns {
			if symbol == usStock {
				return false
			}
		}
	}

	// For ambiguous cases, treat as potentially Indian to allow NSE service to validate
	// This ensures we try NSE India service first for validation
	return true
}
