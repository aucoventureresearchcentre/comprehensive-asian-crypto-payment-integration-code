// Cambodia payment platform integrations for Asian Cryptocurrency Payment System
// Implements integrations with popular Cambodian payment platforms

package integration

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// CambodiaWingConfig holds configuration for Wing integration
type CambodiaWingConfig struct {
	MerchantID     string
	MerchantKey    string
	APIEndpoint    string
	CallbackURL    string
	RedirectURL    string
	TestMode       bool
}

// CambodiaWing implements PaymentPlatform interface for Cambodia's Wing
type CambodiaWing struct {
	config CambodiaWingConfig
	client *http.Client
}

// NewCambodiaWing creates a new Wing payment platform
func NewCambodiaWing(config CambodiaWingConfig) *CambodiaWing {
	// Set default API endpoint if not provided
	if config.APIEndpoint == "" {
		if config.TestMode {
			config.APIEndpoint = "https://sandbox.wingmoney.com/api"
		} else {
			config.APIEndpoint = "https://api.wingmoney.com/api"
		}
	}

	return &CambodiaWing{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the name of the payment platform
func (p *CambodiaWing) GetName() string {
	return "Wing"
}

// GetCountryCode returns the country code of the payment platform
func (p *CambodiaWing) GetCountryCode() string {
	return "KH"
}

// GetSupportedPaymentMethods returns the supported payment methods
func (p *CambodiaWing) GetSupportedPaymentMethods() []PaymentMethod {
	return []PaymentMethod{MethodEWallet, MethodQRCode}
}

// GetSupportedCurrencies returns the supported currencies
func (p *CambodiaWing) GetSupportedCurrencies() []string {
	return []string{"USD", "KHR"}
}

// CreatePayment creates a payment
func (p *CambodiaWing) CreatePayment(request *PaymentRequest) (*PaymentResponse, error) {
	// Validate request
	if request.Currency != "USD" && request.Currency != "KHR" {
		return nil, errors.New("currency must be USD or KHR for Wing payments")
	}

	if request.PaymentMethod != MethodEWallet && request.PaymentMethod != MethodQRCode {
		return nil, ErrUnsupportedPaymentMethod
	}

	// Prepare Wing request
	timestamp := time.Now().Format("20060102150405")
	
	wingRequest := map[string]interface{}{
		"merchant_id":     p.config.MerchantID,
		"order_id":        request.OrderID,
		"amount":          fmt.Sprintf("%.2f", request.Amount),
		"currency":        request.Currency,
		"description":     request.Description,
		"customer_name":   request.CustomerName,
		"customer_email":  request.CustomerEmail,
		"customer_phone":  request.CustomerPhone,
		"return_url":      p.config.RedirectURL,
		"callback_url":    p.config.CallbackURL,
		"timestamp":       timestamp,
	}

	// Set payment method
	if request.PaymentMethod == MethodEWallet {
		wingRequest["payment_type"] = "wallet"
	} else {
		wingRequest["payment_type"] = "qr"
	}

	// Generate signature
	signature := p.generateSignature(wingRequest)
	wingRequest["signature"] = signature

	// Convert to JSON
	jsonData, err := json.Marshal(wingRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", p.config.APIEndpoint+"/v1/payment/create", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Make API request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var wingResponse map[string]interface{}
	if err := json.Unmarshal(body, &wingResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if status, ok := wingResponse["status"].(string); ok && status != "success" {
		errorMsg := "unknown error"
		if msg, ok := wingResponse["message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("Wing error: %s", errorMsg)
	}

	// Extract payment details
	data, ok := wingResponse["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid response format")
	}

	paymentID, _ := data["payment_id"].(string)
	paymentURL, _ := data["payment_url"].(string)
	qrCodeURL, _ := data["qr_code_url"].(string)
	expiryTime, _ := data["expiry_time"].(float64)

	// Create response
	response := &PaymentResponse{
		PaymentID:     paymentID,
		Status:        StatusPending,
		Amount:        request.Amount,
		Currency:      request.Currency,
		PaymentMethod: request.PaymentMethod,
		PaymentURL:    paymentURL,
		QRCodeURL:     qrCodeURL,
		RedirectURL:   paymentURL,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		ExpiresAt:     time.Unix(int64(expiryTime), 0),
		Metadata:      make(map[string]string),
	}

	return response, nil
}

// GetPaymentStatus gets the status of a payment
func (p *CambodiaWing) GetPaymentStatus(request *PaymentStatusRequest) (*PaymentResponse, error) {
	// Prepare status request
	timestamp := time.Now().Format("20060102150405")
	
	statusRequest := map[string]interface{}{
		"merchant_id": p.config.MerchantID,
		"payment_id":  request.PaymentID,
		"timestamp":   timestamp,
	}

	// Generate signature
	signature := p.generateSignature(statusRequest)
	statusRequest["signature"] = signature

	// Convert to JSON
	jsonData, err := json.Marshal(statusRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", p.config.APIEndpoint+"/v1/payment/status", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Make API request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var statusResponse map[string]interface{}
	if err := json.Unmarshal(body, &statusResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if status, ok := statusResponse["status"].(string); ok && status != "success" {
		errorMsg := "unknown error"
		if msg, ok := statusResponse["message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("Wing error: %s", errorMsg)
	}

	// Extract payment details
	data, ok := statusResponse["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid response format")
	}

	paymentStatus, _ := data["status"].(string)
	amountStr, _ := data["amount"].(string)
	amount, _ := strconv.ParseFloat(amountStr, 64)
	currency, _ := data["currency"].(string)
	paymentType, _ := data["payment_type"].(string)
	transactionID, _ := data["transaction_id"].(string)
	createdAtStr, _ := data["created_at"].(string)
	
	// Parse created at
	createdAt, _ := time.Parse("2006-01-02T15:04:05Z", createdAtStr)

	// Map Wing status to our status
	status := StatusPending
	var completedAt time.Time

	switch paymentStatus {
	case "completed", "success":
		status = StatusCompleted
		completedAt = time.Now()
	case "failed", "cancelled":
		status = StatusFailed
	case "pending":
		status = StatusPending
	}

	// Map payment type
	var method PaymentMethod
	if paymentType == "wallet" {
		method = MethodEWallet
	} else if paymentType == "qr" {
		method = MethodQRCode
	} else {
		method = request.PaymentMethod
	}

	// Create response
	response := &PaymentResponse{
		PaymentID:     request.PaymentID,
		Status:        status,
		Amount:        amount,
		Currency:      currency,
		PaymentMethod: method,
		TransactionID: transactionID,
		CreatedAt:     createdAt,
		UpdatedAt:     time.Now(),
		CompletedAt:   completedAt,
		Metadata:      make(map[string]string),
	}

	return response, nil
}

// RefundPayment refunds a payment
func (p *CambodiaWing) RefundPayment(request *RefundRequest) (*RefundResponse, error) {
	// Prepare refund request
	timestamp := time.Now().Format("20060102150405")
	
	refundRequest := map[string]interface{}{
		"merchant_id": p.config.MerchantID,
		"payment_id":  request.PaymentID,
		"refund_id":   request.RefundID,
		"amount":      fmt.Sprintf("%.2f", request.Amount),
		"reason":      request.Reason,
		"timestamp":   timestamp,
	}

	// Generate signature
	signature := p.generateSignature(refundRequest)
	refundRequest["signature"] = signature

	// Convert to JSON
	jsonData, err := json.Marshal(refundRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", p.config.APIEndpoint+"/v1/payment/refund", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Make API request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var refundResponse map[string]interface{}
	if err := json.Unmarshal(body, &refundResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if status, ok := refundResponse["status"].(string); ok && status != "success" {
		errorMsg := "unknown error"
		if msg, ok := refundResponse["message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("Wing refund error: %s", errorMsg)
	}

	// Extract refund details
	data, ok := refundResponse["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid response format")
	}

	refundID, _ := data["refund_id"].(string)
	refundStatus, _ := data["status"].(string)
	transactionID, _ := data["transaction_id"].(string)
	currency, _ := data["currency"].(string)

	// Create response
	response := &RefundResponse{
		RefundID:      refundID,
		PaymentID:     request.PaymentID,
		Amount:        request.Amount,
		Currency:      currency,
		Status:        refundStatus,
		TransactionID: transactionID,
		CreatedAt:     time.Now(),
	}

	return response, nil
}

// generateSignature generates a signature for Wing requests
func (p *CambodiaWing) generateSignature(params map[string]interface{}) string {
	// Sort keys
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build string to sign
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(fmt.Sprintf("%v", params[k]))
		sb.WriteString("&")
	}
	// Remove trailing &
	signStr := sb.String()
	if len(signStr) > 0 {
		signStr = signStr[:len(signStr)-1]
	}

	// Generate HMAC-SHA256
	h := hmac.New(sha256.New, []byte(p.config.MerchantKey))
	h.Write([]byte(signStr))
	return hex.EncodeToString(h.Sum(nil))
}

// CambodiaABAConfig holds configuration for ABA integration
type CambodiaABAConfig struct {
	MerchantID     string
	MerchantAPIKey string
	MerchantSecret string
	APIEndpoint    string
	CallbackURL    string
	RedirectURL    string
	TestMode       bool
}

// CambodiaABA implements PaymentPlatform interface for Cambodia's ABA
type CambodiaABA struct {
	config CambodiaABAConfig
	client *http.Client
}

// NewCambodiaABA creates a new ABA payment platform
func NewCambodiaABA(config CambodiaABAConfig) *CambodiaABA {
	// Set default API endpoint if not provided
	if config.APIEndpoint == "" {
		if config.TestMode {
			config.APIEndpoint = "https://checkout-sandbox.payway.com.kh/api"
		} else {
			config.APIEndpoint = "https://checkout.payway.com.kh/api"
		}
	}

	return &CambodiaABA{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the name of the payment platform
func (p *CambodiaABA) GetName() string {
	return "ABA"
}

// GetCountryCode returns the country code of the payment platform
func (p *CambodiaABA) GetCountryCode() string {
	return "KH"
}

// GetSupportedPaymentMethods returns the supported payment methods
func (p *CambodiaABA) GetSupportedPaymentMethods() []PaymentMethod {
	return []PaymentMethod{MethodCreditCard, MethodEWallet, MethodQRCode}
}

// GetSupportedCurrencies returns the supported currencies
func (p *CambodiaABA) GetSupportedCurrencies() []string {
	return []string{"USD", "KHR"}
}

// CreatePayment creates a payment
func (p *CambodiaABA) CreatePayment(request *PaymentRequest) (*PaymentResponse, error) {
	// Validate request
	if request.Currency != "USD" && request.Currency != "KHR" {
		return nil, errors.New("currency must be USD or KHR for ABA payments")
	}

	if request.PaymentMethod != MethodCreditCard && request.PaymentMethod != MethodEWallet && request.PaymentMethod != MethodQRCode {
		return nil, ErrUnsupportedPaymentMethod
	}

	// Prepare ABA request
	timestamp := time.Now().Format("20060102150405")
	
	abaRequest := map[string]interface{}{
		"merchant_id":     p.config.MerchantID,
		"order_id":        request.OrderID,
		"amount":          request.Amount,
		"currency":        request.Currency,
		"description":     request.Description,
		"customer_name":   request.CustomerName,
		"customer_email":  request.CustomerEmail,
		"customer_phone":  request.CustomerPhone,
		"return_url":      p.config.RedirectURL,
		"continue_success_url": p.config.RedirectURL,
		"callback_url":    p.config.CallbackURL,
		"timestamp":       timestamp,
	}

	// Set payment method
	if request.PaymentMethod == MethodCreditCard {
		abaRequest["payment_option"] = "cards"
	} else if request.PaymentMethod == MethodEWallet {
		abaRequest["payment_option"] = "abapay"
	} else {
		abaRequest["payment_option"] = "qr"
	}

	// Generate hash
	hash := p.generateHash(abaRequest)
	abaRequest["hash"] = hash

	// Convert to JSON
	jsonData, err := json.Marshal(abaRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", p.config.APIEndpoint+"/payment-gateway/v1/payments", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Merchant-ID", p.config.MerchantID)
	req.Header.Set("API-Key", p.config.MerchantAPIKey)

	// Make API request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var abaResponse map[string]interface{}
	if err := json.Unmarshal(body, &abaResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if status, ok := abaResponse["status"].(float64); ok && status != 0 {
		errorMsg := "unknown error"
		if msg, ok := abaResponse["description"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("ABA error: %s", errorMsg)
	}

	// Extract payment details
	paymentID, _ := abaResponse["tran_id"].(string)
	checkoutURL, _ := abaResponse["checkout_url"].(string)

	// Create response
	response := &PaymentResponse{
		PaymentID:     paymentID,
		Status:        StatusPending,
		Amount:        request.Amount,
		Currency:      request.Currency,
		PaymentMethod: request.PaymentMethod,
		PaymentURL:    checkoutURL,
		RedirectURL:   checkoutURL,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(24 * time.Hour),
		Metadata:      make(map[string]string),
	}

	return response, nil
}

// GetPaymentStatus gets the status of a payment
func (p *CambodiaABA) GetPaymentStatus(request *PaymentStatusRequest) (*PaymentResponse, error) {
	// Prepare status request
	timestamp := time.Now().Format("20060102150405")
	
	statusRequest := map[string]interface{}{
		"merchant_id": p.config.MerchantID,
		"tran_id":     request.PaymentID,
		"timestamp":   timestamp,
	}

	// Generate hash
	hash := p.generateHash(statusRequest)
	statusRequest["hash"] = hash

	// Convert to JSON
	jsonData, err := json.Marshal(statusRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", p.config.APIEndpoint+"/payment-gateway/v1/payments/check-transaction", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Merchant-ID", p.config.MerchantID)
	req.Header.Set("API-Key", p.config.MerchantAPIKey)

	// Make API request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var statusResponse map[string]interface{}
	if err := json.Unmarshal(body, &statusResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if status, ok := statusResponse["status"].(float64); ok && status != 0 {
		errorMsg := "unknown error"
		if msg, ok := statusResponse["description"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("ABA error: %s", errorMsg)
	}

	// Extract payment details
	paymentStatus, _ := statusResponse["payment_status"].(string)
	amount, _ := statusResponse["amount"].(float64)
	currency, _ := statusResponse["currency"].(string)
	paymentOption, _ := statusResponse["payment_option"].(string)
	transactionID, _ := statusResponse["transaction_id"].(string)
	createdAtStr, _ := statusResponse["created_date"].(string)
	
	// Parse created at
	createdAt, _ := time.Parse("2006-01-02 15:04:05", createdAtStr)

	// Map ABA status to our status
	status := StatusPending
	var completedAt time.Time

	switch paymentStatus {
	case "2":
		status = StatusCompleted
		completedAt = time.Now()
	case "0", "1":
		status = StatusPending
	default:
		status = StatusFailed
	}

	// Map payment option
	var method PaymentMethod
	if paymentOption == "cards" {
		method = MethodCreditCard
	} else if paymentOption == "abapay" {
		method = MethodEWallet
	} else if paymentOption == "qr" {
		method = MethodQRCode
	} else {
		method = request.PaymentMethod
	}

	// Create response
	response := &PaymentResponse{
		PaymentID:     request.PaymentID,
		Status:        status,
		Amount:        amount,
		Currency:      currency,
		PaymentMethod: method,
		TransactionID: transactionID,
		CreatedAt:     createdAt,
		UpdatedAt:     time.Now(),
		CompletedAt:   completedAt,
		Metadata:      make(map[string]string),
	}

	return response, nil
}

// RefundPayment refunds a payment
func (p *CambodiaABA) RefundPayment(request *RefundRequest) (*RefundResponse, error) {
	// Prepare refund request
	timestamp := time.Now().Format("20060102150405")
	
	refundRequest := map[string]interface{}{
		"merchant_id": p.config.MerchantID,
		"tran_id":     request.PaymentID,
		"refund_id":   request.RefundID,
		"amount":      request.Amount,
		"reason":      request.Reason,
		"timestamp":   timestamp,
	}

	// Generate hash
	hash := p.generateHash(refundRequest)
	refundRequest["hash"] = hash

	// Convert to JSON
	jsonData, err := json.Marshal(refundRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", p.config.APIEndpoint+"/payment-gateway/v1/payments/refund", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Merchant-ID", p.config.MerchantID)
	req.Header.Set("API-Key", p.config.MerchantAPIKey)

	// Make API request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var refundResponse map[string]interface{}
	if err := json.Unmarshal(body, &refundResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if status, ok := refundResponse["status"].(float64); ok && status != 0 {
		errorMsg := "unknown error"
		if msg, ok := refundResponse["description"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("ABA refund error: %s", errorMsg)
	}

	// Extract refund details
	refundID, _ := refundResponse["refund_id"].(string)
	refundStatus, _ := refundResponse["status"].(string)
	transactionID, _ := refundResponse["transaction_id"].(string)
	currency, _ := refundResponse["currency"].(string)

	// Create response
	response := &RefundResponse{
		RefundID:      refundID,
		PaymentID:     request.PaymentID,
		Amount:        request.Amount,
		Currency:      currency,
		Status:        refundStatus,
		TransactionID: transactionID,
		CreatedAt:     time.Now(),
	}

	return response, nil
}

// generateHash generates a hash for ABA requests
func (p *CambodiaABA) generateHash(params map[string]interface{}) string {
	// Sort keys
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build string to sign
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(fmt.Sprintf("%v", params[k]))
		sb.WriteString("&")
	}
	// Remove trailing &
	signStr := sb.String()
	if len(signStr) > 0 {
		signStr = signStr[:len(signStr)-1]
	}

	// Add merchant secret
	signStr += p.config.MerchantSecret

	// Generate SHA-256
	hash := sha256.Sum256([]byte(signStr))
	return hex.EncodeToString(hash[:])
}
