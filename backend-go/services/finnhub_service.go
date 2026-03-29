package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"stock-analysis-api/models"
)

// FinnhubService handles fetching real stock data from Finnhub API
type FinnhubService struct {
	client *http.Client
	apiKey string
}

// NewFinnhubService creates a new instance of FinnhubService
func NewFinnhubService(apiKey string) *FinnhubService {
	return &FinnhubService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey: apiKey,
	}
}

// FinnhubQuote represents the quote response from Finnhub
type FinnhubQuote struct {
	C  float64 `json:"c"`  // Current price
	D  float64 `json:"d"`  // Change
	DP float64 `json:"dp"` // Percent change
	H  float64 `json:"h"`  // High price of the day
	L  float64 `json:"l"`  // Low price of the day
	O  float64 `json:"o"`  // Open price of the day
	PC float64 `json:"pc"` // Previous close price
	T  int64   `json:"t"`  // Timestamp
}

// FinnhubProfile represents company profile from Finnhub
type FinnhubProfile struct {
	Country              string  `json:"country"`
	Currency             string  `json:"currency"`
	Exchange             string  `json:"exchange"`
	FinnhubIndustry      string  `json:"finnhubIndustry"`
	IPO                  string  `json:"ipo"`
	Logo                 string  `json:"logo"`
	MarketCapitalization float64 `json:"marketCapitalization"`
	Name                 string  `json:"name"`
	Phone                string  `json:"phone"`
	ShareOutstanding     float64 `json:"shareOutstanding"`
	Ticker               string  `json:"ticker"`
	Weburl               string  `json:"weburl"`
}

// FinnhubMetrics represents basic financials from Finnhub
type FinnhubMetrics struct {
	Metric struct {
		Beta                         float64 `json:"beta"`
		DividendYieldIndicatedAnnual float64 `json:"dividendYieldIndicatedAnnual"`
		EPS                          float64 `json:"eps"`
		EPSEstimateNextQuarter       float64 `json:"epsEstimateNextQuarter"`
		EPSEstimateNextYear          float64 `json:"epsEstimateNextYear"`
		EPSEstimateQuarter           float64 `json:"epsEstimateQuarter"`
		EPSEstimateYear              float64 `json:"epsEstimateYear"`
		PBRatio                      float64 `json:"pbRatio"`
		PEBasicExclExtraTTM          float64 `json:"peBasicExclExtraTTM"`
		PERatio                      float64 `json:"peRatio"`
		ROE                          float64 `json:"roeTTM"`
		ROA                          float64 `json:"roaTTM"`
	} `json:"metric"`
}

// GetStockData fetches real-time stock data from Finnhub
func (fh *FinnhubService) GetStockData(symbol string) (models.FundamentalsData, models.TechnicalData, error) {
	if fh.apiKey == "" {
		return models.FundamentalsData{}, models.TechnicalData{}, fmt.Errorf("Finnhub API key not provided")
	}

	// Get quote data
	quote, err := fh.getQuote(symbol)
	if err != nil {
		return models.FundamentalsData{}, models.TechnicalData{}, err
	}

	// Get company profile
	profile, err := fh.getProfile(symbol)
	if err != nil {
		log.Printf("Warning: Could not get profile for %s: %v", symbol, err)
	}

	// Get basic financials
	metrics, err := fh.getMetrics(symbol)
	if err != nil {
		log.Printf("Warning: Could not get metrics for %s: %v", symbol, err)
	}

	// Build response
	fundamentals := fh.buildFundamentals(symbol, quote, profile, metrics)
	technical := fh.buildTechnical(symbol, quote, profile)

	return fundamentals, technical, nil
}

// getQuote fetches quote data from Finnhub
func (fh *FinnhubService) getQuote(symbol string) (FinnhubQuote, error) {
	url := fmt.Sprintf("https://finnhub.io/api/v1/quote?symbol=%s&token=%s", symbol, fh.apiKey)

	resp, err := fh.client.Get(url)
	if err != nil {
		return FinnhubQuote{}, fmt.Errorf("failed to fetch quote: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return FinnhubQuote{}, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FinnhubQuote{}, fmt.Errorf("failed to read response: %v", err)
	}

	var quote FinnhubQuote
	if err := json.Unmarshal(body, &quote); err != nil {
		return FinnhubQuote{}, fmt.Errorf("failed to parse quote response: %v", err)
	}

	return quote, nil
}

// getProfile fetches company profile from Finnhub
func (fh *FinnhubService) getProfile(symbol string) (FinnhubProfile, error) {
	url := fmt.Sprintf("https://finnhub.io/api/v1/stock/profile2?symbol=%s&token=%s", symbol, fh.apiKey)

	resp, err := fh.client.Get(url)
	if err != nil {
		return FinnhubProfile{}, fmt.Errorf("failed to fetch profile: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FinnhubProfile{}, fmt.Errorf("failed to read response: %v", err)
	}

	var profile FinnhubProfile
	if err := json.Unmarshal(body, &profile); err != nil {
		return FinnhubProfile{}, fmt.Errorf("failed to parse profile response: %v", err)
	}

	return profile, nil
}

// getMetrics fetches basic financials from Finnhub
func (fh *FinnhubService) getMetrics(symbol string) (FinnhubMetrics, error) {
	url := fmt.Sprintf("https://finnhub.io/api/v1/stock/metric?symbol=%s&metric=all&token=%s", symbol, fh.apiKey)

	resp, err := fh.client.Get(url)
	if err != nil {
		return FinnhubMetrics{}, fmt.Errorf("failed to fetch metrics: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FinnhubMetrics{}, fmt.Errorf("failed to read response: %v", err)
	}

	var metrics FinnhubMetrics
	if err := json.Unmarshal(body, &metrics); err != nil {
		return FinnhubMetrics{}, fmt.Errorf("failed to parse metrics response: %v", err)
	}

	return metrics, nil
}

// buildFundamentals converts Finnhub data to FundamentalsData
func (fh *FinnhubService) buildFundamentals(symbol string, quote FinnhubQuote, profile FinnhubProfile, metrics FinnhubMetrics) models.FundamentalsData {
	// Calculate market cap in actual value (Finnhub returns in millions)
	marketCap := int64(profile.MarketCapitalization * 1000000)

	return models.FundamentalsData{
		Symbol:        symbol,
		CompanyName:   profile.Name,
		MarketCap:     marketCap,
		PERatio:       metrics.Metric.PERatio,
		PBRatio:       metrics.Metric.PBRatio,
		DividendYield: metrics.Metric.DividendYieldIndicatedAnnual,
		ROE:           metrics.Metric.ROE * 100, // Convert to percentage
		CurrentPrice:  quote.C,
		Revenue:       0, // Not available in free tier
		Profit:        0, // Not available in free tier
		DebtToEquity:  0, // Not available in free tier
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}
}

// buildTechnical converts Finnhub data to TechnicalData
func (fh *FinnhubService) buildTechnical(symbol string, quote FinnhubQuote, profile FinnhubProfile) models.TechnicalData {
	return models.TechnicalData{
		Symbol:        symbol,
		Price:         quote.C,
		ChangePercent: quote.DP,
		Volume:        "N/A", // Volume not included in basic quote
		MarketCap:     fh.formatMarketCap(int64(profile.MarketCapitalization * 1000000)),
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}
}

// formatMarketCap formats market cap for display
func (fh *FinnhubService) formatMarketCap(marketCap int64) string {
	if marketCap == 0 {
		return "N/A"
	}
	if marketCap >= 1000000000000 { // Trillion
		return fmt.Sprintf("%.2fT", float64(marketCap)/1000000000000)
	} else if marketCap >= 1000000000 { // Billion
		return fmt.Sprintf("%.2fB", float64(marketCap)/1000000000)
	} else if marketCap >= 1000000 { // Million
		return fmt.Sprintf("%.2fM", float64(marketCap)/1000000)
	}
	return fmt.Sprintf("%d", marketCap)
}
