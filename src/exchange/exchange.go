// Exchange rate service for Asian Cryptocurrency Payment System
// Provides real-time exchange rates between cryptocurrencies and fiat currencies

package exchange

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Common errors
var (
	ErrRateNotFound      = errors.New("exchange rate not found")
	ErrInvalidCurrency   = errors.New("invalid currency")
	ErrProviderUnavailable = errors.New("exchange rate provider unavailable")
)

// ExchangeRate represents a currency exchange rate
type ExchangeRate struct {
	BaseCurrency   string    `json:"base_currency"`
	TargetCurrency string    `json:"target_currency"`
	Rate           float64   `json:"rate"`
	Source         string    `json:"source"`
	Timestamp      time.Time `json:"timestamp"`
}

// RateProvider defines the interface for exchange rate providers
type RateProvider interface {
	// GetName returns the name of the provider
	GetName() string
	
	// GetRate returns the exchange rate between two currencies
	GetRate(baseCurrency, targetCurrency string) (*ExchangeRate, error)
	
	// GetSupportedCurrencies returns the list of supported currencies
	GetSupportedCurrencies() []string
}

// ExchangeRateService manages exchange rates from multiple providers
type ExchangeRateService struct {
	providers     []RateProvider
	cacheEnabled  bool
	cacheDuration time.Duration
	cache         map[string]*ExchangeRate
	cacheMutex    sync.RWMutex
}

// NewExchangeRateService creates a new exchange rate service
func NewExchangeRateService(cacheEnabled bool, cacheDuration time.Duration) *ExchangeRateService {
	service := &ExchangeRateService{
		providers:     make([]RateProvider, 0),
		cacheEnabled:  cacheEnabled,
		cacheDuration: cacheDuration,
		cache:         make(map[string]*ExchangeRate),
	}
	
	return service
}

// RegisterProvider adds a rate provider to the service
func (s *ExchangeRateService) RegisterProvider(provider RateProvider) {
	s.providers = append(s.providers, provider)
}

// GetRate returns the exchange rate between two currencies
func (s *ExchangeRateService) GetRate(baseCurrency, targetCurrency string) (*ExchangeRate, error) {
	// Check cache first if enabled
	if s.cacheEnabled {
		cacheKey := fmt.Sprintf("%s-%s", baseCurrency, targetCurrency)
		s.cacheMutex.RLock()
		cachedRate, exists := s.cache[cacheKey]
		s.cacheMutex.RUnlock()
		
		if exists && time.Since(cachedRate.Timestamp) < s.cacheDuration {
			return cachedRate, nil
		}
	}
	
	// Try each provider until we get a rate
	var lastError error
	for _, provider := range s.providers {
		rate, err := provider.GetRate(baseCurrency, targetCurrency)
		if err == nil {
			// Update cache if enabled
			if s.cacheEnabled {
				cacheKey := fmt.Sprintf("%s-%s", baseCurrency, targetCurrency)
				s.cacheMutex.Lock()
				s.cache[cacheKey] = rate
				s.cacheMutex.Unlock()
			}
			return rate, nil
		}
		lastError = err
	}
	
	// If we get here, all providers failed
	if lastError != nil {
		return nil, lastError
	}
	return nil, ErrRateNotFound
}

// GetRateWithSpread returns the exchange rate with a spread applied
func (s *ExchangeRateService) GetRateWithSpread(baseCurrency, targetCurrency string, spreadPercentage float64) (*ExchangeRate, error) {
	rate, err := s.GetRate(baseCurrency, targetCurrency)
	if err != nil {
		return nil, err
	}
	
	// Apply spread
	spreadFactor := 1.0 + (spreadPercentage / 100.0)
	rate.Rate = rate.Rate * spreadFactor
	
	return rate, nil
}

// ClearCache clears the exchange rate cache
func (s *ExchangeRateService) ClearCache() {
	s.cacheMutex.Lock()
	s.cache = make(map[string]*ExchangeRate)
	s.cacheMutex.Unlock()
}

// GetSupportedCurrencies returns the list of supported currencies across all providers
func (s *ExchangeRateService) GetSupportedCurrencies() []string {
	currencyMap := make(map[string]bool)
	
	for _, provider := range s.providers {
		for _, currency := range provider.GetSupportedCurrencies() {
			currencyMap[currency] = true
		}
	}
	
	currencies := make([]string, 0, len(currencyMap))
	for currency := range currencyMap {
		currencies = append(currencies, currency)
	}
	
	return currencies
}

// ConvertAmount converts an amount from one currency to another
func (s *ExchangeRateService) ConvertAmount(amount float64, fromCurrency, toCurrency string) (float64, error) {
	rate, err := s.GetRate(fromCurrency, toCurrency)
	if err != nil {
		return 0, err
	}
	
	return amount * rate.Rate, nil
}

// ConvertAmountWithSpread converts an amount with a spread applied
func (s *ExchangeRateService) ConvertAmountWithSpread(amount float64, fromCurrency, toCurrency string, spreadPercentage float64) (float64, error) {
	rate, err := s.GetRateWithSpread(fromCurrency, toCurrency, spreadPercentage)
	if err != nil {
		return 0, err
	}
	
	return amount * rate.Rate, nil
}
