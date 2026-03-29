package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"stock-analysis-api/models"
)

// AlphaVantageService handles fetching real stock data from Alpha Vantage API
type AlphaVantageService struct {
	client *http.Client
	apiKey string
}

// NewAlphaVantageService creates a new instance of AlphaVantageService
func NewAlphaVantageService(apiKey string) *AlphaVantageService {
	return &AlphaVantageService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey: apiKey,
	}
}

// AlphaVantageQuote represents the quote response from Alpha Vantage
type AlphaVantageQuote struct {
	GlobalQuote struct {
		Symbol           string `json:"01. symbol"`
		Open             string `json:"02. open"`
		High             string `json:"03. high"`
		Low              string `json:"04. low"`
		Price            string `json:"05. price"`
		Volume           string `json:"06. volume"`
		LatestTradingDay string `json:"07. latest trading day"`
		PreviousClose    string `json:"08. previous close"`
		Change           string `json:"09. change"`
		ChangePercent    string `json:"10. change percent"`
	} `json:"Global Quote"`
	ErrorMessage string `json:"Error Message,omitempty"`
	Note         string `json:"Note,omitempty"`
	Information  string `json:"Information,omitempty"`
}

// AlphaVantageOverview represents company overview from Alpha Vantage
type AlphaVantageOverview struct {
	Symbol                     string `json:"Symbol"`
	Name                       string `json:"Name"`
	Exchange                   string `json:"Exchange"`
	Currency                   string `json:"Currency"`
	Country                    string `json:"Country"`
	Sector                     string `json:"Sector"`
	Industry                   string `json:"Industry"`
	MarketCapitalization       string `json:"MarketCapitalization"`
	BookValue                  string `json:"BookValue"`
	DividendYield              string `json:"DividendYield"`
	EPS                        string `json:"EPS"`
	RevenuePerShareTTM         string `json:"RevenuePerShareTTM"`
	ProfitMargin               string `json:"ProfitMargin"`
	OperatingMarginTTM         string `json:"OperatingMarginTTM"`
	ReturnOnAssetsTTM          string `json:"ReturnOnAssetsTTM"`
	ReturnOnEquityTTM          string `json:"ReturnOnEquityTTM"`
	RevenueTTM                 string `json:"RevenueTTM"`
	GrossProfitTTM             string `json:"GrossProfitTTM"`
	DilutedEPSTTM              string `json:"DilutedEPSTTM"`
	QuarterlyEarningsGrowthYOY string `json:"QuarterlyEarningsGrowthYOY"`
	QuarterlyRevenueGrowthYOY  string `json:"QuarterlyRevenueGrowthYOY"`
	AnalystTargetPrice         string `json:"AnalystTargetPrice"`
	TrailingPE                 string `json:"TrailingPE"`
	ForwardPE                  string `json:"ForwardPE"`
	PriceToSalesRatioTTM       string `json:"PriceToSalesRatioTTM"`
	PriceToBookRatio           string `json:"PriceToBookRatio"`
	EVToRevenue                string `json:"EVToRevenue"`
	EVToEBITDA                 string `json:"EVToEBITDA"`
	Beta                       string `json:"Beta"`
	ErrorMessage               string `json:"Error Message,omitempty"`
	Note                       string `json:"Note,omitempty"`
	Information                string `json:"Information,omitempty"`
}

// GetStockData fetches real-time stock data from Alpha Vantage
func (av *AlphaVantageService) GetStockData(symbol string) (models.FundamentalsData, models.TechnicalData, error) {
	if av.apiKey == "" {
		return models.FundamentalsData{}, models.TechnicalData{}, fmt.Errorf("Alpha Vantage API key not provided")
	}

	// Get quote data
	quote, err := av.getQuote(symbol)
	if err != nil {
		return models.FundamentalsData{}, models.TechnicalData{}, err
	}

	// Get company overview
	overview, err := av.getOverview(symbol)
	if err != nil {
		log.Printf("Warning: Could not get overview for %s: %v", symbol, err)
		// Continue with quote data only
	} else {
		log.Printf("Successfully got overview for %s: Name=%s, MarketCap=%s, PE=%s", symbol, overview.Name, overview.MarketCapitalization, overview.TrailingPE)
	}

	// Parse and convert data
	fundamentals := av.buildFundamentals(symbol, quote, overview)
	technical := av.buildTechnical(symbol, quote, overview)

	return fundamentals, technical, nil
}

// getQuote fetches quote data from Alpha Vantage
func (av *AlphaVantageService) getQuote(symbol string) (AlphaVantageQuote, error) {
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=%s", symbol, av.apiKey)

	resp, err := av.client.Get(url)
	if err != nil {
		return AlphaVantageQuote{}, fmt.Errorf("failed to fetch quote: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return AlphaVantageQuote{}, fmt.Errorf("failed to read response: %v", err)
	}

	var quote AlphaVantageQuote
	if err := json.Unmarshal(body, &quote); err != nil {
		return AlphaVantageQuote{}, fmt.Errorf("failed to parse quote response: %v", err)
	}

	if quote.ErrorMessage != "" {
		return AlphaVantageQuote{}, fmt.Errorf("API error: %s", quote.ErrorMessage)
	}

	if quote.Note != "" {
		return AlphaVantageQuote{}, fmt.Errorf("API limit reached: %s", quote.Note)
	}

	if quote.Information != "" {
		return AlphaVantageQuote{}, fmt.Errorf("API rate limit: %s", quote.Information)
	}

	return quote, nil
}

// getOverview fetches company overview from Alpha Vantage
func (av *AlphaVantageService) getOverview(symbol string) (AlphaVantageOverview, error) {
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=OVERVIEW&symbol=%s&apikey=%s", symbol, av.apiKey)

	resp, err := av.client.Get(url)
	if err != nil {
		return AlphaVantageOverview{}, fmt.Errorf("failed to fetch overview: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return AlphaVantageOverview{}, fmt.Errorf("failed to read response: %v", err)
	}

	log.Printf("Raw overview response for %s: %s", symbol, string(body))

	var overview AlphaVantageOverview
	if err := json.Unmarshal(body, &overview); err != nil {
		return AlphaVantageOverview{}, fmt.Errorf("failed to parse overview response: %v", err)
	}

	if overview.ErrorMessage != "" {
		return AlphaVantageOverview{}, fmt.Errorf("API error: %s", overview.ErrorMessage)
	}

	if overview.Information != "" {
		return AlphaVantageOverview{}, fmt.Errorf("API rate limit: %s", overview.Information)
	}

	log.Printf("Parsed overview for %s: Symbol=%s, Name=%s, MarketCap=%s", symbol, overview.Symbol, overview.Name, overview.MarketCapitalization)

	return overview, nil
}

// buildFundamentals converts Alpha Vantage data to FundamentalsData
func (av *AlphaVantageService) buildFundamentals(symbol string, quote AlphaVantageQuote, overview AlphaVantageOverview) models.FundamentalsData {
	price, _ := strconv.ParseFloat(quote.GlobalQuote.Price, 64)
	marketCap, _ := strconv.ParseInt(overview.MarketCapitalization, 10, 64)
	peRatio, _ := strconv.ParseFloat(overview.TrailingPE, 64)
	pbRatio, _ := strconv.ParseFloat(overview.PriceToBookRatio, 64)
	dividendYield, _ := strconv.ParseFloat(overview.DividendYield, 64)
	roe, _ := strconv.ParseFloat(overview.ReturnOnEquityTTM, 64)
	revenue, _ := strconv.ParseInt(overview.RevenueTTM, 10, 64)

	// Calculate profit from profit margin and revenue
	profitMargin, _ := strconv.ParseFloat(overview.ProfitMargin, 64)
	profit := int64(float64(revenue) * profitMargin)

	return models.FundamentalsData{
		Symbol:        symbol,
		CompanyName:   overview.Name,
		MarketCap:     marketCap,
		PERatio:       peRatio,
		PBRatio:       pbRatio,
		DividendYield: dividendYield * 100, // Convert to percentage
		ROE:           roe * 100,           // Convert to percentage
		CurrentPrice:  price,
		Revenue:       revenue,
		Profit:        profit,
		DebtToEquity:  0, // Not available in Alpha Vantage free tier
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}
}

// buildTechnical converts Alpha Vantage data to TechnicalData
func (av *AlphaVantageService) buildTechnical(symbol string, quote AlphaVantageQuote, overview AlphaVantageOverview) models.TechnicalData {
	price, _ := strconv.ParseFloat(quote.GlobalQuote.Price, 64)
	volume, _ := strconv.ParseInt(quote.GlobalQuote.Volume, 10, 64)
	changePercentStr := quote.GlobalQuote.ChangePercent

	// Remove % sign and parse
	if len(changePercentStr) > 0 && changePercentStr[len(changePercentStr)-1] == '%' {
		changePercentStr = changePercentStr[:len(changePercentStr)-1]
	}
	changePercent, _ := strconv.ParseFloat(changePercentStr, 64)

	return models.TechnicalData{
		Symbol:        symbol,
		Price:         price,
		ChangePercent: changePercent,
		Volume:        av.formatVolume(volume),
		MarketCap:     av.formatMarketCap(overview.MarketCapitalization),
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}
}

// formatVolume formats volume for display
func (av *AlphaVantageService) formatVolume(volume int64) string {
	if volume >= 10000000 {
		return fmt.Sprintf("%.1fM", float64(volume)/1000000)
	} else if volume >= 100000 {
		return fmt.Sprintf("%.1fK", float64(volume)/1000)
	}
	return strconv.FormatInt(volume, 10)
}

// formatMarketCap formats market cap for display
func (av *AlphaVantageService) formatMarketCap(marketCapStr string) string {
	if marketCapStr == "" {
		return "N/A"
	}

	marketCap, err := strconv.ParseInt(marketCapStr, 10, 64)
	if err != nil {
		return "N/A"
	}

	if marketCap >= 1000000000000 { // Trillion
		return fmt.Sprintf("%.2fT", float64(marketCap)/1000000000000)
	} else if marketCap >= 1000000000 { // Billion
		return fmt.Sprintf("%.2fB", float64(marketCap)/1000000000)
	} else if marketCap >= 1000000 { // Million
		return fmt.Sprintf("%.2fM", float64(marketCap)/1000000)
	}
	return strconv.FormatInt(marketCap, 10)
}
