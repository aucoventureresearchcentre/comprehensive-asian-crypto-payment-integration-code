// Fraud detection module for Asian Cryptocurrency Payment System
// Provides fraud detection and prevention capabilities

package security

import (
	"errors"
	"net"
	"strings"
	"sync"
	"time"
)

// Common errors
var (
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrSuspiciousActivity = errors.New("suspicious activity detected")
	ErrBlockedIP = errors.New("IP address is blocked")
	ErrBlockedCountry = errors.New("country is blocked")
)

// FraudDetectionConfig holds configuration for fraud detection
type FraudDetectionConfig struct {
	RateLimitEnabled      bool
	RateLimitWindow       time.Duration
	RateLimitMaxRequests  int
	BlockedIPs            []string
	BlockedCountries      []string
	TransactionThreshold  float64
	VelocityCheckEnabled  bool
	VelocityCheckWindow   time.Duration
	VelocityCheckLimit    int
	SuspiciousPatterns    []string
}

// DefaultFraudDetectionConfig returns a default configuration
func DefaultFraudDetectionConfig() *FraudDetectionConfig {
	return &FraudDetectionConfig{
		RateLimitEnabled:      true,
		RateLimitWindow:       time.Minute,
		RateLimitMaxRequests:  100,
		BlockedIPs:            []string{},
		BlockedCountries:      []string{},
		TransactionThreshold:  10000.0, // Transactions above this amount trigger additional checks
		VelocityCheckEnabled:  true,
		VelocityCheckWindow:   time.Hour,
		VelocityCheckLimit:    10, // Maximum number of transactions per hour
		SuspiciousPatterns:    []string{},
	}
}

// RateLimiter implements rate limiting functionality
type RateLimiter struct {
	window      time.Duration
	maxRequests int
	requests    map[string][]time.Time
	mutex       sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(window time.Duration, maxRequests int) *RateLimiter {
	return &RateLimiter{
		window:      window,
		maxRequests: maxRequests,
		requests:    make(map[string][]time.Time),
	}
}

// CheckLimit checks if a key has exceeded the rate limit
func (r *RateLimiter) CheckLimit(key string) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	// Get existing requests for this key
	times, exists := r.requests[key]
	if !exists {
		r.requests[key] = []time.Time{now}
		return true
	}

	// Filter out old requests
	var newTimes []time.Time
	for _, t := range times {
		if t.After(cutoff) {
			newTimes = append(newTimes, t)
		}
	}

	// Check if limit is exceeded
	if len(newTimes) >= r.maxRequests {
		r.requests[key] = newTimes
		return false
	}

	// Add current request
	r.requests[key] = append(newTimes, now)
	return true
}

// FraudDetectionService provides fraud detection functionality
type FraudDetectionService struct {
	config      *FraudDetectionConfig
	rateLimiter *RateLimiter
	ipCache     map[string]bool
	countryCache map[string]bool
	transactions map[string][]time.Time // Key is user/merchant ID
	mutex       sync.RWMutex
}

// NewFraudDetectionService creates a new fraud detection service
func NewFraudDetectionService(config *FraudDetectionConfig) *FraudDetectionService {
	if config == nil {
		config = DefaultFraudDetectionConfig()
	}

	// Initialize IP cache
	ipCache := make(map[string]bool)
	for _, ip := range config.BlockedIPs {
		ipCache[ip] = true
	}

	// Initialize country cache
	countryCache := make(map[string]bool)
	for _, country := range config.BlockedCountries {
		countryCache[country] = true
	}

	return &FraudDetectionService{
		config:      config,
		rateLimiter: NewRateLimiter(config.RateLimitWindow, config.RateLimitMaxRequests),
		ipCache:     ipCache,
		countryCache: countryCache,
		transactions: make(map[string][]time.Time),
	}
}

// CheckRequest checks if a request should be allowed
func (s *FraudDetectionService) CheckRequest(ipAddress, userID string) error {
	// Check if IP is blocked
	if s.IsIPBlocked(ipAddress) {
		return ErrBlockedIP
	}

	// Check rate limit
	if s.config.RateLimitEnabled {
		if !s.rateLimiter.CheckLimit(ipAddress) {
			return ErrRateLimitExceeded
		}
	}

	return nil
}

// CheckTransaction checks if a transaction should be allowed
func (s *FraudDetectionService) CheckTransaction(userID, ipAddress, countryCode string, amount float64) error {
	// Check if IP is blocked
	if s.IsIPBlocked(ipAddress) {
		return ErrBlockedIP
	}

	// Check if country is blocked
	if s.IsCountryBlocked(countryCode) {
		return ErrBlockedCountry
	}

	// Check transaction amount
	if amount > s.config.TransactionThreshold {
		// In a real implementation, we might trigger additional verification
		// For now, we'll just log it
	}

	// Check transaction velocity
	if s.config.VelocityCheckEnabled {
		if !s.checkVelocity(userID) {
			return ErrRateLimitExceeded
		}
	}

	return nil
}

// IsIPBlocked checks if an IP address is blocked
func (s *FraudDetectionService) IsIPBlocked(ipAddress string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Check exact match
	if s.ipCache[ipAddress] {
		return true
	}

	// Check CIDR blocks
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return false
	}

	for blockedIP := range s.ipCache {
		if strings.Contains(blockedIP, "/") {
			_, ipNet, err := net.ParseCIDR(blockedIP)
			if err == nil && ipNet.Contains(ip) {
				return true
			}
		}
	}

	return false
}

// IsCountryBlocked checks if a country is blocked
func (s *FraudDetectionService) IsCountryBlocked(countryCode string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.countryCache[strings.ToUpper(countryCode)]
}

// BlockIP adds an IP address to the block list
func (s *FraudDetectionService) BlockIP(ipAddress string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.ipCache[ipAddress] = true
	s.config.BlockedIPs = append(s.config.BlockedIPs, ipAddress)
}

// UnblockIP removes an IP address from the block list
func (s *FraudDetectionService) UnblockIP(ipAddress string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.ipCache, ipAddress)

	// Update config
	var newBlockedIPs []string
	for _, ip := range s.config.BlockedIPs {
		if ip != ipAddress {
			newBlockedIPs = append(newBlockedIPs, ip)
		}
	}
	s.config.BlockedIPs = newBlockedIPs
}

// BlockCountry adds a country to the block list
func (s *FraudDetectionService) BlockCountry(countryCode string) {
	countryCode = strings.ToUpper(countryCode)
	
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.countryCache[countryCode] = true
	s.config.BlockedCountries = append(s.config.BlockedCountries, countryCode)
}

// UnblockCountry removes a country from the block list
func (s *FraudDetectionService) UnblockCountry(countryCode string) {
	countryCode = strings.ToUpper(countryCode)
	
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.countryCache, countryCode)

	// Update config
	var newBlockedCountries []string
	for _, country := range s.config.BlockedCountries {
		if country != countryCode {
			newBlockedCountries = append(newBlockedCountries, country)
		}
	}
	s.config.BlockedCountries = newBlockedCountries
}

// checkVelocity checks if a user has exceeded the transaction velocity limit
func (s *FraudDetectionService) checkVelocity(userID string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-s.config.VelocityCheckWindow)

	// Get existing transactions for this user
	times, exists := s.transactions[userID]
	if !exists {
		s.transactions[userID] = []time.Time{now}
		return true
	}

	// Filter out old transactions
	var newTimes []time.Time
	for _, t := range times {
		if t.After(cutoff) {
			newTimes = append(newTimes, t)
		}
	}

	// Check if limit is exceeded
	if len(newTimes) >= s.config.VelocityCheckLimit {
		s.transactions[userID] = newTimes
		return false
	}

	// Add current transaction
	s.transactions[userID] = append(newTimes, now)
	return true
}

// RecordTransaction records a transaction for velocity checking
func (s *FraudDetectionService) RecordTransaction(userID string) {
	if !s.config.VelocityCheckEnabled {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	times, exists := s.transactions[userID]
	if !exists {
		s.transactions[userID] = []time.Time{now}
		return
	}

	s.transactions[userID] = append(times, now)
}

// IsSuspiciousPattern checks if a string contains suspicious patterns
func (s *FraudDetectionService) IsSuspiciousPattern(text string) bool {
	for _, pattern := range s.config.SuspiciousPatterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

// AddSuspiciousPattern adds a pattern to the suspicious patterns list
func (s *FraudDetectionService) AddSuspiciousPattern(pattern string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.config.SuspiciousPatterns = append(s.config.SuspiciousPatterns, pattern)
}
