package collect

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/host-uk/core/pkg/io"
	"github.com/stretchr/testify/assert"
)

func TestMarketCollector_Name_Good(t *testing.T) {
	m := &MarketCollector{CoinID: "bitcoin"}
	assert.Equal(t, "market:bitcoin", m.Name())
}

func TestMarketCollector_Collect_Bad_NoCoinID(t *testing.T) {
	mock := io.NewMockMedium()
	cfg := NewConfigWithMedium(mock, "/output")

	m := &MarketCollector{}
	_, err := m.Collect(context.Background(), cfg)
	assert.Error(t, err)
}

func TestMarketCollector_Collect_Good_DryRun(t *testing.T) {
	mock := io.NewMockMedium()
	cfg := NewConfigWithMedium(mock, "/output")
	cfg.DryRun = true

	m := &MarketCollector{CoinID: "bitcoin"}
	result, err := m.Collect(context.Background(), cfg)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.Items)
}

func TestMarketCollector_Collect_Good_CurrentData(t *testing.T) {
	// Set up a mock CoinGecko server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := coinData{
			ID:     "bitcoin",
			Symbol: "btc",
			Name:   "Bitcoin",
			MarketData: marketData{
				CurrentPrice:      map[string]float64{"usd": 42000.50},
				MarketCap:         map[string]float64{"usd": 800000000000},
				TotalVolume:       map[string]float64{"usd": 25000000000},
				High24h:           map[string]float64{"usd": 43000},
				Low24h:            map[string]float64{"usd": 41000},
				PriceChange24h:    500.25,
				PriceChangePct24h: 1.2,
				MarketCapRank:     1,
				CirculatingSupply: 19500000,
				TotalSupply:       21000000,
				LastUpdated:       "2025-01-15T10:00:00Z",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(data)
	}))
	defer server.Close()

	// Override base URL
	oldURL := coinGeckoBaseURL
	coinGeckoBaseURL = server.URL
	defer func() { coinGeckoBaseURL = oldURL }()

	mock := io.NewMockMedium()
	cfg := NewConfigWithMedium(mock, "/output")
	// Disable rate limiter to avoid delays in tests
	cfg.Limiter = nil

	m := &MarketCollector{CoinID: "bitcoin"}
	result, err := m.Collect(context.Background(), cfg)

	assert.NoError(t, err)
	assert.Equal(t, 2, result.Items) // current.json + summary.md
	assert.Len(t, result.Files, 2)

	// Verify current.json was written
	content, err := mock.Read("/output/market/bitcoin/current.json")
	assert.NoError(t, err)
	assert.Contains(t, content, "bitcoin")

	// Verify summary.md was written
	summary, err := mock.Read("/output/market/bitcoin/summary.md")
	assert.NoError(t, err)
	assert.Contains(t, summary, "Bitcoin")
	assert.Contains(t, summary, "42000.50")
}

func TestMarketCollector_Collect_Good_Historical(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		if callCount == 1 {
			// Current data response
			data := coinData{
				ID:     "ethereum",
				Symbol: "eth",
				Name:   "Ethereum",
				MarketData: marketData{
					CurrentPrice: map[string]float64{"usd": 3000},
				},
			}
			_ = json.NewEncoder(w).Encode(data)
		} else {
			// Historical data response
			data := historicalData{
				Prices:       [][]float64{{1705305600000, 3000.0}, {1705392000000, 3100.0}},
				MarketCaps:   [][]float64{{1705305600000, 360000000000}},
				TotalVolumes: [][]float64{{1705305600000, 15000000000}},
			}
			_ = json.NewEncoder(w).Encode(data)
		}
	}))
	defer server.Close()

	oldURL := coinGeckoBaseURL
	coinGeckoBaseURL = server.URL
	defer func() { coinGeckoBaseURL = oldURL }()

	mock := io.NewMockMedium()
	cfg := NewConfigWithMedium(mock, "/output")
	cfg.Limiter = nil

	m := &MarketCollector{CoinID: "ethereum", Historical: true}
	result, err := m.Collect(context.Background(), cfg)

	assert.NoError(t, err)
	assert.Equal(t, 3, result.Items) // current.json + summary.md + historical.json
	assert.Len(t, result.Files, 3)

	// Verify historical.json was written
	content, err := mock.Read("/output/market/ethereum/historical.json")
	assert.NoError(t, err)
	assert.Contains(t, content, "3000")
}

func TestFormatMarketSummary_Good(t *testing.T) {
	data := &coinData{
		Name:   "Bitcoin",
		Symbol: "btc",
		MarketData: marketData{
			CurrentPrice:      map[string]float64{"usd": 50000},
			MarketCap:         map[string]float64{"usd": 1000000000000},
			MarketCapRank:     1,
			CirculatingSupply: 19500000,
			TotalSupply:       21000000,
		},
	}

	summary := FormatMarketSummary(data)

	assert.Contains(t, summary, "# Bitcoin (BTC)")
	assert.Contains(t, summary, "$50000.00")
	assert.Contains(t, summary, "Market Cap Rank:** #1")
	assert.Contains(t, summary, "Circulating Supply")
	assert.Contains(t, summary, "Total Supply")
}

func TestMarketCollector_Collect_Bad_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	oldURL := coinGeckoBaseURL
	coinGeckoBaseURL = server.URL
	defer func() { coinGeckoBaseURL = oldURL }()

	mock := io.NewMockMedium()
	cfg := NewConfigWithMedium(mock, "/output")
	cfg.Limiter = nil

	m := &MarketCollector{CoinID: "bitcoin"}
	result, err := m.Collect(context.Background(), cfg)

	// Should have errors but not fail entirely
	assert.NoError(t, err)
	assert.Equal(t, 1, result.Errors)
}
