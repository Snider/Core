package collect

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	core "github.com/host-uk/core/pkg/framework/core"
)

// coinGeckoBaseURL is the base URL for the CoinGecko API.
// It is a variable so it can be overridden in tests.
var coinGeckoBaseURL = "https://api.coingecko.com/api/v3"

// MarketCollector collects market data from CoinGecko.
type MarketCollector struct {
	// CoinID is the CoinGecko coin identifier (e.g. "bitcoin", "ethereum").
	CoinID string

	// Historical enables collection of historical market chart data.
	Historical bool

	// FromDate is the start date for historical data in YYYY-MM-DD format.
	FromDate string
}

// Name returns the collector name.
func (m *MarketCollector) Name() string {
	return fmt.Sprintf("market:%s", m.CoinID)
}

// coinData represents the current coin data from CoinGecko.
type coinData struct {
	ID         string     `json:"id"`
	Symbol     string     `json:"symbol"`
	Name       string     `json:"name"`
	MarketData marketData `json:"market_data"`
}

type marketData struct {
	CurrentPrice      map[string]float64 `json:"current_price"`
	MarketCap         map[string]float64 `json:"market_cap"`
	TotalVolume       map[string]float64 `json:"total_volume"`
	High24h           map[string]float64 `json:"high_24h"`
	Low24h            map[string]float64 `json:"low_24h"`
	PriceChange24h    float64            `json:"price_change_24h"`
	PriceChangePct24h float64            `json:"price_change_percentage_24h"`
	MarketCapRank     int                `json:"market_cap_rank"`
	TotalSupply       float64            `json:"total_supply"`
	CirculatingSupply float64            `json:"circulating_supply"`
	LastUpdated       string             `json:"last_updated"`
}

// historicalData represents historical market chart data from CoinGecko.
type historicalData struct {
	Prices       [][]float64 `json:"prices"`
	MarketCaps   [][]float64 `json:"market_caps"`
	TotalVolumes [][]float64 `json:"total_volumes"`
}

// Collect gathers market data from CoinGecko.
func (m *MarketCollector) Collect(ctx context.Context, cfg *Config) (*Result, error) {
	result := &Result{Source: m.Name()}

	if m.CoinID == "" {
		return result, core.E("collect.Market.Collect", "coin ID is required", nil)
	}

	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitStart(m.Name(), fmt.Sprintf("Starting market data collection for %s", m.CoinID))
	}

	if cfg.DryRun {
		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitProgress(m.Name(), fmt.Sprintf("[dry-run] Would collect market data for %s", m.CoinID), nil)
		}
		return result, nil
	}

	baseDir := filepath.Join(cfg.OutputDir, "market", m.CoinID)
	if err := cfg.Output.EnsureDir(baseDir); err != nil {
		return result, core.E("collect.Market.Collect", "failed to create output directory", err)
	}

	// Collect current data
	currentResult, err := m.collectCurrent(ctx, cfg, baseDir)
	if err != nil {
		result.Errors++
		if cfg.Dispatcher != nil {
			cfg.Dispatcher.EmitError(m.Name(), fmt.Sprintf("Failed to collect current data: %v", err), nil)
		}
	} else {
		result.Items += currentResult.Items
		result.Files = append(result.Files, currentResult.Files...)
	}

	// Collect historical data if requested
	if m.Historical {
		histResult, err := m.collectHistorical(ctx, cfg, baseDir)
		if err != nil {
			result.Errors++
			if cfg.Dispatcher != nil {
				cfg.Dispatcher.EmitError(m.Name(), fmt.Sprintf("Failed to collect historical data: %v", err), nil)
			}
		} else {
			result.Items += histResult.Items
			result.Files = append(result.Files, histResult.Files...)
		}
	}

	if cfg.Dispatcher != nil {
		cfg.Dispatcher.EmitComplete(m.Name(), fmt.Sprintf("Collected market data for %s", m.CoinID), result)
	}

	return result, nil
}

// collectCurrent fetches current coin data from CoinGecko.
func (m *MarketCollector) collectCurrent(ctx context.Context, cfg *Config, baseDir string) (*Result, error) {
	result := &Result{Source: m.Name()}

	if cfg.Limiter != nil {
		if err := cfg.Limiter.Wait(ctx, "coingecko"); err != nil {
			return result, err
		}
	}

	url := fmt.Sprintf("%s/coins/%s", coinGeckoBaseURL, m.CoinID)
	data, err := fetchJSON[coinData](ctx, url)
	if err != nil {
		return result, core.E("collect.Market.collectCurrent", "failed to fetch coin data", err)
	}

	// Write raw JSON
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return result, core.E("collect.Market.collectCurrent", "failed to marshal data", err)
	}

	jsonPath := filepath.Join(baseDir, "current.json")
	if err := cfg.Output.Write(jsonPath, string(jsonBytes)); err != nil {
		return result, core.E("collect.Market.collectCurrent", "failed to write JSON", err)
	}
	result.Items++
	result.Files = append(result.Files, jsonPath)

	// Write summary markdown
	summary := formatMarketSummary(data)
	summaryPath := filepath.Join(baseDir, "summary.md")
	if err := cfg.Output.Write(summaryPath, summary); err != nil {
		return result, core.E("collect.Market.collectCurrent", "failed to write summary", err)
	}
	result.Items++
	result.Files = append(result.Files, summaryPath)

	return result, nil
}

// collectHistorical fetches historical market chart data from CoinGecko.
func (m *MarketCollector) collectHistorical(ctx context.Context, cfg *Config, baseDir string) (*Result, error) {
	result := &Result{Source: m.Name()}

	if cfg.Limiter != nil {
		if err := cfg.Limiter.Wait(ctx, "coingecko"); err != nil {
			return result, err
		}
	}

	days := "365"
	if m.FromDate != "" {
		fromTime, err := time.Parse("2006-01-02", m.FromDate)
		if err == nil {
			dayCount := int(time.Since(fromTime).Hours() / 24)
			if dayCount > 0 {
				days = fmt.Sprintf("%d", dayCount)
			}
		}
	}

	url := fmt.Sprintf("%s/coins/%s/market_chart?vs_currency=usd&days=%s", coinGeckoBaseURL, m.CoinID, days)
	data, err := fetchJSON[historicalData](ctx, url)
	if err != nil {
		return result, core.E("collect.Market.collectHistorical", "failed to fetch historical data", err)
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return result, core.E("collect.Market.collectHistorical", "failed to marshal data", err)
	}

	jsonPath := filepath.Join(baseDir, "historical.json")
	if err := cfg.Output.Write(jsonPath, string(jsonBytes)); err != nil {
		return result, core.E("collect.Market.collectHistorical", "failed to write JSON", err)
	}
	result.Items++
	result.Files = append(result.Files, jsonPath)

	return result, nil
}

// fetchJSON fetches JSON from a URL and unmarshals it into the given type.
func fetchJSON[T any](ctx context.Context, url string) (*T, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, core.E("collect.fetchJSON", "failed to create request", err)
	}
	req.Header.Set("User-Agent", "CoreCollector/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, core.E("collect.fetchJSON", "request failed", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, core.E("collect.fetchJSON",
			fmt.Sprintf("unexpected status code: %d for %s", resp.StatusCode, url), nil)
	}

	var data T
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, core.E("collect.fetchJSON", "failed to decode response", err)
	}

	return &data, nil
}

// formatMarketSummary formats coin data as a markdown summary.
func formatMarketSummary(data *coinData) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s (%s)\n\n", data.Name, strings.ToUpper(data.Symbol))

	md := data.MarketData

	if price, ok := md.CurrentPrice["usd"]; ok {
		fmt.Fprintf(&b, "- **Current Price (USD):** $%.2f\n", price)
	}
	if cap, ok := md.MarketCap["usd"]; ok {
		fmt.Fprintf(&b, "- **Market Cap (USD):** $%.0f\n", cap)
	}
	if vol, ok := md.TotalVolume["usd"]; ok {
		fmt.Fprintf(&b, "- **24h Volume (USD):** $%.0f\n", vol)
	}
	if high, ok := md.High24h["usd"]; ok {
		fmt.Fprintf(&b, "- **24h High (USD):** $%.2f\n", high)
	}
	if low, ok := md.Low24h["usd"]; ok {
		fmt.Fprintf(&b, "- **24h Low (USD):** $%.2f\n", low)
	}

	fmt.Fprintf(&b, "- **24h Price Change:** $%.2f (%.2f%%)\n", md.PriceChange24h, md.PriceChangePct24h)

	if md.MarketCapRank > 0 {
		fmt.Fprintf(&b, "- **Market Cap Rank:** #%d\n", md.MarketCapRank)
	}
	if md.CirculatingSupply > 0 {
		fmt.Fprintf(&b, "- **Circulating Supply:** %.0f\n", md.CirculatingSupply)
	}
	if md.TotalSupply > 0 {
		fmt.Fprintf(&b, "- **Total Supply:** %.0f\n", md.TotalSupply)
	}
	if md.LastUpdated != "" {
		fmt.Fprintf(&b, "\n*Last updated: %s*\n", md.LastUpdated)
	}

	return b.String()
}

// FormatMarketSummary is exported for testing.
func FormatMarketSummary(data *coinData) string {
	return formatMarketSummary(data)
}
