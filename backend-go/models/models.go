package models

import "time"

// StockAnalysisRequest represents the request body for stock analysis
type StockAnalysisRequest struct {
	Symbol   string `json:"symbol" binding:"required,min=1,max=20"`
	Exchange string `json:"exchange"`
}

// StockAnalysisResponse represents the response for stock analysis
type StockAnalysisResponse struct {
	Symbol           string             `json:"symbol"`
	CompanyName      string             `json:"company_name"`
	CurrentPrice     float64            `json:"current_price"`
	Recommendation   string             `json:"recommendation"`
	Confidence       string             `json:"confidence"`
	Reason           string             `json:"reason"`
	TargetBuyPrice   float64            `json:"target_buy_price"`
	TargetSellPrice  float64            `json:"target_sell_price"`
	StopLoss         float64            `json:"stop_loss"`
	SupportLevels    map[string]float64 `json:"support_levels"`
	ResistanceLevels map[string]float64 `json:"resistance_levels"`
	Fundamentals     FundamentalsData   `json:"fundamentals"`
	Technical        TechnicalData      `json:"technical"`
	Timestamp        string             `json:"timestamp"`
}

// PopularStock represents a popular stock item
type PopularStock struct {
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Exchange string `json:"exchange"`
}

// FundamentalsData represents fundamental stock data
type FundamentalsData struct {
	Symbol        string  `json:"symbol"`
	CompanyName   string  `json:"company_name"`
	MarketCap     int64   `json:"market_cap"`
	PERatio       float64 `json:"pe_ratio"`
	PBRatio       float64 `json:"pb_ratio"`
	DividendYield float64 `json:"dividend_yield"`
	ROE           float64 `json:"roe"`
	CurrentPrice  float64 `json:"current_price"`
	Revenue       int64   `json:"revenue"`
	Profit        int64   `json:"profit"`
	DebtToEquity  float64 `json:"debt_to_equity"`
	Timestamp     string  `json:"timestamp"`
	Error         string  `json:"error,omitempty"`
}

// TechnicalData represents technical stock data
type TechnicalData struct {
	Symbol        string  `json:"symbol"`
	Price         float64 `json:"price"`
	ChangePercent float64 `json:"change_percent"`
	Volume        string  `json:"volume"`
	MarketCap     string  `json:"market_cap"`
	Timestamp     string  `json:"timestamp"`
	Error         string  `json:"error,omitempty"`
}

// AnalysisHistory represents historical analysis data
type AnalysisHistory struct {
	Symbol   string                  `json:"symbol"`
	Count    int                     `json:"count"`
	Analyses []StockAnalysisResponse `json:"analyses"`
}

// SupportResistance represents support and resistance levels
type SupportResistance struct {
	Support    map[string]float64 `json:"support"`
	Resistance map[string]float64 `json:"resistance"`
	Pivot      float64            `json:"pivot"`
}

// Recommendation represents analysis recommendation
type Recommendation struct {
	Action     string `json:"action"`
	Confidence string `json:"confidence"`
	Reason     string `json:"reason"`
	Score      int    `json:"score"`
}

// PriceTargets represents price target calculations
type PriceTargets struct {
	BuyPrice  float64 `json:"buy_price"`
	SellPrice float64 `json:"sell_price"`
	StopLoss  float64 `json:"stop_loss"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Detail string `json:"detail"`
}

// APIResponse represents a basic API response
type APIResponse struct {
	Message string `json:"message"`
	Version string `json:"version,omitempty"`
}

// MongoDB document for storing analysis
type AnalysisDocument struct {
	StockAnalysisResponse `bson:",inline"`
	CreatedAt             time.Time `bson:"created_at"`
}
