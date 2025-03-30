// CoinGecko exchange rate provider for Asian Cryptocurrency Payment System
// Integrates with the CoinGecko API to fetch cryptocurrency exchange rates

package exchange

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// CoinGeckoProvider implements the RateProvider interface using CoinGecko API
type CoinGeckoProvider struct {
	apiURL             string
	httpClient         *http.Client
	supportedCurrencies []string
	coinIdMap          map[string]string // Maps currency codes to CoinGecko IDs
}

// NewCoinGeckoProvider creates a new CoinGecko provider
func NewCoinGeckoProvider() *CoinGeckoProvider {
	provider := &CoinGeckoProvider{
		apiURL: "https://api.coingecko.com/api/v3",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		supportedCurrencies: []string{
			// Cryptocurrencies
			"BTC", "ETH", "USDT", "USDC", "BNB", "XRP", "ADA", "SOL", "DOT", "DOGE",
			// Fiat currencies
			"USD", "EUR", "JPY", "GBP", "AUD", "CAD", "CHF", "CNY", "HKD", "NZD",
			// Asian currencies
			"MYR", "SGD", "IDR", "THB", "BND", "KHR", "VND", "LAK", "PHP", "MMK",
		},
		coinIdMap: map[string]string{
			"BTC": "bitcoin",
			"ETH": "ethereum",
			"USDT": "tether",
			"USDC": "usd-coin",
			"BNB": "binancecoin",
			"XRP": "ripple",
			"ADA": "cardano",
			"SOL": "solana",
			"DOT": "polkadot",
			"DOGE": "dogecoin",
		},
	}
	
	return provider
}

// GetName returns the name of the provider
func (p *CoinGeckoProvider) GetName() string {
	return "CoinGecko"
}

// GetRate returns the exchange rate between two currencies
func (p *CoinGeckoProvider) GetRate(baseCurrency, targetCurrency string) (*ExchangeRate, error) {
	baseCurrency = strings.ToUpper(baseCurrency)
	targetCurrency = strings.ToUpper(targetCurrency)
	
	// Validate currencies
	if !p.isSupportedCurrency(baseCurrency) || !p.isSupportedCurrency(targetCurrency) {
		return nil, ErrInvalidCurrency
	}
	
	// Handle different scenarios:
	// 1. Crypto to Fiat (most common)
	// 2. Fiat to Crypto
	// 3. Crypto to Crypto
	// 4. Fiat to Fiat
	
	if p.isCryptoCurrency(baseCurrency) && !p.isCryptoCurrency(targetCurrency) {
		// Crypto to Fiat
		return p.getCryptoToFiatRate(baseCurrency, targetCurrency)
	} else if !p.isCryptoCurrency(baseCurrency) && p.isCryptoCurrency(targetCurrency) {
		// Fiat to Crypto - get inverse rate and then invert it
		rate, err := p.getCryptoToFiatRate(targetCurrency, baseCurrency)
		if err != nil {
			return nil, err
		}
		
		// Invert the rate
		rate.BaseCurrency = baseCurrency
		rate.TargetCurrency = targetCurrency
		rate.Rate = 1.0 / rate.Rate
		
		return rate, nil
	} else if p.isCryptoCurrency(baseCurrency) && p.isCryptoCurrency(targetCurrency) {
		// Crypto to Crypto - get both in USD and then calculate
		baseToUSD, err := p.getCryptoToFiatRate(baseCurrency, "USD")
		if err != nil {
			return nil, err
		}
		
		targetToUSD, err := p.getCryptoToFiatRate(targetCurrency, "USD")
		if err != nil {
			return nil, err
		}
		
		// Calculate cross rate
		rate := &ExchangeRate{
			BaseCurrency:   baseCurrency,
			TargetCurrency: targetCurrency,
			Rate:           baseToUSD.Rate / targetToUSD.Rate,
			Source:         p.GetName(),
			Timestamp:      time.Now(),
		}
		
		return rate, nil
	} else {
		// Fiat to Fiat - use a third-party API or service
		// For simplicity, we'll use USD as an intermediate currency
		// In a real implementation, you might want to use a dedicated forex API
		
		// This is a placeholder implementation
		return &ExchangeRate{
			BaseCurrency:   baseCurrency,
			TargetCurrency: targetCurrency,
			Rate:           1.0, // Placeholder
			Source:         p.GetName(),
			Timestamp:      time.Now(),
		}, nil
	}
}

// getCryptoToFiatRate gets the exchange rate from a cryptocurrency to a fiat currency
func (p *CoinGeckoProvider) getCryptoToFiatRate(cryptoCurrency, fiatCurrency string) (*ExchangeRate, error) {
	// Convert currency codes to CoinGecko format
	coinId, exists := p.coinIdMap[cryptoCurrency]
	if !exists {
		return nil, ErrInvalidCurrency
	}
	
	fiatCurrency = strings.ToLower(fiatCurrency)
	
	// Build API URL
	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=%s", p.apiURL, coinId, fiatCurrency)
	
	// Make request
	resp, err := p.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exchange rate: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, ErrProviderUnavailable
	}
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	// Parse response
	var result map[string]map[string]float64
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	// Extract rate
	coinData, exists := result[coinId]
	if !exists {
		return nil, ErrRateNotFound
	}
	
	rate, exists := coinData[fiatCurrency]
	if !exists {
		return nil, ErrRateNotFound
	}
	
	// Create exchange rate object
	exchangeRate := &ExchangeRate{
		BaseCurrency:   cryptoCurrency,
		TargetCurrency: strings.ToUpper(fiatCurrency),
		Rate:           rate,
		Source:         p.GetName(),
		Timestamp:      time.Now(),
	}
	
	return exchangeRate, nil
}

// GetSupportedCurrencies returns the list of supported currencies
func (p *CoinGeckoProvider) GetSupportedCurrencies() []string {
	return p.supportedCurrencies
}

// isSupportedCurrency checks if a currency is supported
func (p *CoinGeckoProvider) isSupportedCurrency(currency string) bool {
	for _, c := range p.supportedCurrencies {
		if c == currency {
			return true
		}
	}
	return false
}

// isCryptoCurrency checks if a currency is a cryptocurrency
func (p *CoinGeckoProvider) isCryptoCurrency(currency string) bool {
	_, isCrypto := p.coinIdMap[currency]
	return isCrypto
}
