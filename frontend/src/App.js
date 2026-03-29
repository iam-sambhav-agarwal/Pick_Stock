import { useState, useEffect } from "react";
import "@/App.css";
import axios from "axios";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Search, TrendingUp, TrendingDown, DollarSign, AlertCircle } from "lucide-react";
import { toast } from "sonner";

// Backend URL configuration with fallbacks for different environments
const BACKEND_URL = process.env.REACT_APP_BACKEND_URL || 
  (process.env.NODE_ENV === 'production' ? '' : 'http://localhost:8001');
const API = `${BACKEND_URL}/api`;

function App() {
  const [symbol, setSymbol] = useState("");
  const [loading, setLoading] = useState(false);
  const [analysis, setAnalysis] = useState(null);
  const [popularStocks, setPopularStocks] = useState([]);

  // Fetch popular stocks on mount
  useEffect(() => {
    fetchPopularStocks();
  }, []);

  const fetchPopularStocks = async () => {
    try {
      const response = await axios.get(`${API}/stocks/popular`);
      setPopularStocks(response.data);
    } catch (error) {
      console.error("Error fetching popular stocks:", error);
    }
  };

  const analyzeStock = async (stockSymbol) => {
    if (!stockSymbol || stockSymbol.trim() === "") {
      toast.error("Please enter a stock symbol");
      return;
    }

    setLoading(true);
    setAnalysis(null);

    try {
      const response = await axios.post(`${API}/analyze`, {
        symbol: stockSymbol.toUpperCase(),
        exchange: "NSE"
      });

      setAnalysis(response.data);
      toast.success(`Analysis complete for ${stockSymbol.toUpperCase()}`);
    } catch (error) {
      console.error("Error analyzing stock:", error);
      toast.error(error.response?.data?.detail || "Failed to analyze stock. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = (e) => {
    e.preventDefault();
    analyzeStock(symbol);
  };

  const handleStockSelect = (value) => {
    setSymbol(value);
    analyzeStock(value);
  };

  const getRecommendationColor = (recommendation) => {
    switch (recommendation) {
      case "BUY":
        return "text-buy";
      case "SELL":
        return "text-sell";
      case "HOLD":
        return "text-hold";
      default:
        return "text-foreground";
    }
  };

  const getRecommendationBg = (recommendation) => {
    switch (recommendation) {
      case "BUY":
        return "bg-buyBg border-buy";
      case "SELL":
        return "bg-sellBg border-sell";
      case "HOLD":
        return "bg-holdBg border-hold";
      default:
        return "bg-muted";
    }
  };

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b border-border bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <h1 className="text-3xl sm:text-4xl font-heading font-bold tracking-tight text-foreground">
            Stock Analysis
          </h1>
          <p className="mt-2 text-base text-muted-foreground">
            Get buy/sell/hold recommendations based on fundamental and technical analysis
          </p>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Search Section */}
        <Card className="p-6 shadow-sm border border-border" data-testid="search-section">
          <div className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {/* Text Input */}
              <div>
                <label className="block text-sm font-medium text-foreground mb-2">
                  Search by Symbol
                </label>
                <form onSubmit={handleSearch} className="flex gap-2">
                  <Input
                    data-testid="stock-search-input"
                    type="text"
                    placeholder="Enter stock symbol (e.g., TCS, RELIANCE)"
                    value={symbol}
                    onChange={(e) => setSymbol(e.target.value.toUpperCase())}
                    className="flex-1"
                  />
                  <Button
                    data-testid="analyze-button"
                    type="submit"
                    disabled={loading}
                    className="px-6"
                  >
                    <Search className="h-4 w-4 mr-2" />
                    {loading ? "Analyzing..." : "Analyze"}
                  </Button>
                </form>
              </div>

              {/* Dropdown Select */}
              <div>
                <label className="block text-sm font-medium text-foreground mb-2">
                  Or Select Popular Stock
                </label>
                <Select onValueChange={handleStockSelect} disabled={loading}>
                  <SelectTrigger data-testid="popular-stocks-dropdown">
                    <SelectValue placeholder="Select a popular stock" />
                  </SelectTrigger>
                  <SelectContent>
                    {popularStocks.map((stock) => (
                      <SelectItem key={stock.symbol} value={stock.symbol}>
                        {stock.symbol} - {stock.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
          </div>
        </Card>

        {/* Loading State */}
        {loading && (
          <div className="mt-8 text-center" data-testid="loading-state">
            <div className="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
            <p className="mt-4 text-muted-foreground">Analyzing stock data...</p>
          </div>
        )}

        {/* Results Section */}
        {!loading && analysis && (
          <div className="mt-8 space-y-6" data-testid="analysis-results">
            {/* Header Info */}
            <div>
              <h2 className="text-2xl font-heading font-semibold tracking-tight text-foreground">
                {analysis.company_name} ({analysis.symbol})
              </h2>
              <p className="text-sm text-muted-foreground mt-1">
                Current Price: ₹{analysis.current_price.toFixed(2)}
              </p>
            </div>

            {/* Bento Grid Layout */}
            <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-4 gap-4">
              {/* Main Recommendation Card - Spans 2 columns */}
              <Card
                className={`p-6 md:col-span-2 border-2 ${getRecommendationBg(analysis.recommendation)}`}
                data-testid="recommendation-card"
              >
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <h3 className="text-lg font-heading font-semibold tracking-tight text-foreground">
                      Recommendation
                    </h3>
                    <Badge
                      className={`px-3 py-1 text-sm font-semibold uppercase ${getRecommendationColor(analysis.recommendation)}`}
                      data-testid="recommendation-badge"
                    >
                      {analysis.recommendation}
                    </Badge>
                  </div>

                  <div className="space-y-2">
                    <p className="text-sm text-muted-foreground">Confidence: {analysis.confidence}</p>
                    <p className="text-sm text-foreground">
                      <AlertCircle className="inline h-4 w-4 mr-1" />
                      {analysis.reason}
                    </p>
                  </div>

                  <div className="grid grid-cols-3 gap-4 pt-4 border-t border-border">
                    <div>
                      <p className="text-xs text-muted-foreground uppercase tracking-wider">Buy Price</p>
                      <p className="text-lg font-semibold text-buy" data-testid="target-buy-price">
                        ₹{analysis.target_buy_price.toFixed(2)}
                      </p>
                    </div>
                    <div>
                      <p className="text-xs text-muted-foreground uppercase tracking-wider">Sell Price</p>
                      <p className="text-lg font-semibold text-sell" data-testid="target-sell-price">
                        ₹{analysis.target_sell_price.toFixed(2)}
                      </p>
                    </div>
                    <div>
                      <p className="text-xs text-muted-foreground uppercase tracking-wider">Stop Loss</p>
                      <p className="text-lg font-semibold text-destructive" data-testid="stop-loss">
                        ₹{analysis.stop_loss.toFixed(2)}
                      </p>
                    </div>
                  </div>
                </div>
              </Card>

              {/* Support Levels Card */}
              <Card className="p-6 border border-border shadow-sm" data-testid="support-card">
                <h3 className="text-base font-heading font-semibold tracking-tight text-foreground mb-4">
                  <TrendingDown className="inline h-5 w-5 mr-2 text-sell" />
                  Support Levels
                </h3>
                <div className="space-y-3">
                  <div className="flex justify-between items-center">
                    <span className="text-sm text-muted-foreground">S1</span>
                    <span className="text-sm font-medium text-foreground">
                      ₹{analysis.support_levels.S1.toFixed(2)}
                    </span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-sm text-muted-foreground">S2</span>
                    <span className="text-sm font-medium text-foreground">
                      ₹{analysis.support_levels.S2.toFixed(2)}
                    </span>
                  </div>
                </div>
              </Card>

              {/* Resistance Levels Card */}
              <Card className="p-6 border border-border shadow-sm" data-testid="resistance-card">
                <h3 className="text-base font-heading font-semibold tracking-tight text-foreground mb-4">
                  <TrendingUp className="inline h-5 w-5 mr-2 text-buy" />
                  Resistance Levels
                </h3>
                <div className="space-y-3">
                  <div className="flex justify-between items-center">
                    <span className="text-sm text-muted-foreground">R1</span>
                    <span className="text-sm font-medium text-foreground">
                      ₹{analysis.resistance_levels.R1.toFixed(2)}
                    </span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-sm text-muted-foreground">R2</span>
                    <span className="text-sm font-medium text-foreground">
                      ₹{analysis.resistance_levels.R2.toFixed(2)}
                    </span>
                  </div>
                </div>
              </Card>

              {/* Fundamentals Card - Spans 2 rows on large screens */}
              <Card className="p-6 border border-border shadow-sm md:col-span-3 lg:col-span-2 lg:row-span-2" data-testid="fundamentals-card">
                <h3 className="text-lg font-heading font-semibold tracking-tight text-foreground mb-4">
                  Fundamental Metrics
                </h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-xs text-muted-foreground uppercase tracking-wider">P/E Ratio</p>
                    <p className="text-base font-medium text-foreground mt-1">
                      {analysis.fundamentals.pe_ratio.toFixed(2)}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground uppercase tracking-wider">P/B Ratio</p>
                    <p className="text-base font-medium text-foreground mt-1">
                      {analysis.fundamentals.pb_ratio.toFixed(2)}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground uppercase tracking-wider">ROE (%)</p>
                    <p className="text-base font-medium text-foreground mt-1">
                      {analysis.fundamentals.roe.toFixed(2)}%
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground uppercase tracking-wider">Dividend Yield</p>
                    <p className="text-base font-medium text-foreground mt-1">
                      {analysis.fundamentals.dividend_yield.toFixed(2)}%
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground uppercase tracking-wider">Debt to Equity</p>
                    <p className="text-base font-medium text-foreground mt-1">
                      {analysis.fundamentals.debt_to_equity.toFixed(2)}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground uppercase tracking-wider">Market Cap</p>
                    <p className="text-base font-medium text-foreground mt-1">
                      ₹{(analysis.fundamentals.market_cap / 10000000).toFixed(2)}Cr
                    </p>
                  </div>
                </div>
              </Card>

              {/* Technical Indicators Card */}
              <Card className="p-6 border border-border shadow-sm md:col-span-3 lg:col-span-2" data-testid="technical-card">
                <h3 className="text-lg font-heading font-semibold tracking-tight text-foreground mb-4">
                  <DollarSign className="inline h-5 w-5 mr-2" />
                  Technical Indicators
                </h3>
                <div className="grid grid-cols-3 gap-4">
                  <div>
                    <p className="text-xs text-muted-foreground uppercase tracking-wider">Price</p>
                    <p className="text-base font-medium text-foreground mt-1">
                      ₹{analysis.technical.price.toFixed(2)}
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground uppercase tracking-wider">Change</p>
                    <p
                      className={`text-base font-medium mt-1 ${
                        analysis.technical.change_percent >= 0 ? "text-buy" : "text-sell"
                      }`}
                    >
                      {analysis.technical.change_percent >= 0 ? "+" : ""}
                      {analysis.technical.change_percent.toFixed(2)}%
                    </p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground uppercase tracking-wider">Volume</p>
                    <p className="text-base font-medium text-foreground mt-1">
                      {analysis.technical.volume}
                    </p>
                  </div>
                </div>
              </Card>
            </div>

            {/* Disclaimer */}
            <Card className="p-4 border border-border bg-muted" data-testid="disclaimer">
              <p className="text-xs text-muted-foreground text-center">
                <AlertCircle className="inline h-3 w-3 mr-1" />
                This analysis is for informational purposes only and should not be considered as financial advice.
                Always do your own research before making investment decisions.
              </p>
            </Card>
          </div>
        )}
      </main>
    </div>
  );
}

export default App;
