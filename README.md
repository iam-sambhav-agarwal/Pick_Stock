# 📊 Stock Analysis App - Real-time Indian Market Data

A modern, real-time stock analysis application focused on Indian markets (NSE/BSE) with intelligent BUY/SELL/HOLD recommendations.

![Stock Analysis](https://img.shields.io/badge/Status-Production%20Ready-brightgreen)
![Go](https://img.shields.io/badge/Go-1.21+-blue)
![React](https://img.shields.io/badge/React-18+-blue)
![NSE](https://img.shields.io/badge/NSE-Live%20Data-orange)

## ✨ Features

### 📈 Real-time Stock Data
- **Live NSE prices**: TCS (₹2,377.9), RELIANCE (₹1,413.8), INFY (₹1,278.6)
- **150+ Indian stocks** supported with dynamic validation
- **Multi-provider fallback**: NSE India → Alpha Vantage → Finnhub → Yahoo Finance

### 🎯 Smart Analysis
- **BUY/SELL/HOLD recommendations**
- Technical indicators (RSI, Moving Averages)
- Risk assessment and price targets
- Support/Resistance level detection

### 🛡️ Robust Error Handling
- Proper "stock not found" errors for invalid symbols
- Graceful provider fallbacks
- Real-time validation against NSE API

## 🚀 Quick Start

### Local Development
```bash
# Clone the repository
git clone https://github.com/YOUR_USERNAME/stock-analysis-app.git
cd stock-analysis-app

# Build frontend
cd frontend
npm install
npm run build

# Run Go backend
cd ../backend-go
go mod download
go run main.go

# Access at http://localhost:8001
```

### Environment Variables
```env
ALPHA_VANTAGE_API_KEY=your_key_here
FINNHUB_API_KEY=your_key_here
PORT=8001
CORS_ORIGINS=*
GIN_MODE=release
```

## 🌐 Deployment

### Deploy to Render (Recommended)
1. Push this repository to GitHub
2. Go to [Render.com](https://render.com)
3. Connect your repository
4. Render will auto-detect the configuration from `render.yaml`
5. Click Deploy!

The app will be live at: `https://your-app-name.onrender.com`

### Deploy to Railway
```bash
npm install -g @railway/cli
railway login
railway up
```

## 🏗️ Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   React Frontend │────│   Go Backend     │────│   NSE India API │
│   (Port 3000)   │    │   (Port 8001)    │    │   (Live Data)   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                              │
                         ┌────▼────┐
                         │ Fallback│
                         │ Alpha V │
                         │ Finnhub │
                         │ Yahoo   │
                         └─────────┘
```

## 📱 API Endpoints

### Stock Analysis
- `GET /` - React frontend application
- `GET /api/stocks/popular` - Get popular Indian stocks
- `POST /api/analyze` - Analyze specific stock
- `GET /api/history/:symbol` - Get analysis history

### Example API Usage
```javascript
// Analyze TCS stock
fetch('/api/analyze', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ symbol: 'TCS' })
})
```

## 🧪 Supported Stocks

### Major Indian Stocks
- **IT**: TCS, INFY, WIPRO, HCL, TECHM
- **Banking**: HDFCBANK, ICICIBANK, SBIN, KOTAKBANK
- **Energy**: RELIANCE, ONGC, BPCL, IOC
- **Telecom**: BHARTIARTL, TTML, IDEA
- **Auto**: MARUTI, TATAMOTORS, M&M
- **And 150+ more...**

## 🔧 Tech Stack

### Backend
- **Go 1.21+** with Gin framework
- **Real-time APIs**: NSE India, Alpha Vantage, Finnhub
- **MongoDB** for analysis history (optional)
- **CORS** enabled for cross-origin requests

### Frontend
- **React 18** with hooks
- **Tailwind CSS** for styling
- **Radix UI** components
- **Axios** for API calls
- **Responsive** mobile-first design

## 📊 Performance

- **Response Time**: <500ms for cached data
- **Build Size**: Frontend ~2MB, Backend ~15MB
- **API Rate Limits**: 
  - NSE India: Unlimited (primary)
  - Alpha Vantage: 25/day
  - Finnhub: 60/minute

## 🛠️ Development

### Project Structure
```
├── backend-go/           # Go backend
│   ├── main.go          # Main application
│   ├── handlers/        # HTTP handlers
│   ├── models/          # Data models
│   └── services/        # External API services
├── frontend/            # React frontend
│   ├── src/
│   │   ├── components/  # UI components
│   │   └── hooks/       # Custom hooks
│   └── public/          # Static assets
├── render.yaml          # Deployment config
└── Dockerfile           # Container config
```

### Testing
```bash
# Test Go backend
cd backend-go
go test ./...

# Test frontend
cd frontend
npm test
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Open a Pull Request

## 📄 License

MIT License - feel free to use this project for your own stock analysis needs!

## 🆘 Support

Having issues? Check these:
1. **Build fails**: Ensure Node.js 18+ and Go 1.21+ are installed
2. **API errors**: Check your API keys in environment variables
3. **CORS issues**: Verify CORS_ORIGINS is set correctly
4. **Stock not found**: Ensure you're using valid NSE symbols

---

**Built with ❤️ for Indian stock market analysis**

⭐ Star this repository if you found it useful!
