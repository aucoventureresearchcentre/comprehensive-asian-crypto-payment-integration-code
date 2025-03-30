// Laos payment platform integrations for Asian Cryptocurrency Payment System
// Implements integrations with popular Laotian payment platforms

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

// LaosUMoneyConfig holds configuration for U-Money integration
type LaosUMoneyConfig struct {
	MerchantID     string
	MerchantKey    string
	APIEndpoint    string
	CallbackURL    string
	RedirectURL    string
	TestMode       bool
}

// LaosUMoney implements PaymentPlatform interface for Laos's U-Money
type LaosUMoney struct {
	config LaosUMoneyConfig
	client *http.Client
}

// NewLaosUMoney creates a new U-Money payment platform
func NewLaosUMoney(config LaosUMoneyConfig) *LaosUMoney {
	// Set default API endpoint if not provided
	if config.APIEndpoint == "" {
		if config.TestMode {
			config.APIEndpoint = "https://sandbox-api.umoney.la"
		} else {
			config.APIEndpoint = "https://api.umoney.la"
		}
	}

	return &LaosUMoney{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the name of the payment platform
func (p *LaosUMoney) GetName() string {
	return "U-Money"
}

// GetCountryCode returns the country code of the payment platform
func (p *LaosUMoney) GetCountryCode() string {
	return "LA"
}

// GetSupportedPaymentMethods returns the supported payment methods
func (p *LaosUMoney) GetSupportedPaymentMethods() []PaymentMethod {
	return []PaymentMethod{MethodEWallet, MethodQRCode}
}

// GetSupportedCurrencies returns the supported currencies
func (p *LaosUMoney) GetSupportedCurrencies() []string {
	return []string{"LAK"}
}

// CreatePayment creates a payment
func (p *LaosUMoney) CreatePayment(request *PaymentRequest) (*PaymentResponse, error) {
	// Validate request
	if request.Currency != "LAK" {
		return nil, errors.New("currency must be LAK for U-Money payments")
	}

	if request.PaymentMethod != MethodEWallet && request.PaymentMethod != MethodQRCode {
		return nil, ErrUnsupportedPaymentMethod
	}

	// Prepare U-Money request
	timestamp := time.Now().Format("20060102150405")
	
	uMoneyRequest := map[string]interface{}{
		"merchant_id":     p.config.MerchantID,
		"order_id":        request.OrderID,
		"amount":          int(request.Amount), // U-Money expects integer amount
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
		uMoneyRequest["payment_type"] = "wallet"
	} else {
		uMoneyRequest["payment_type"] = "qr"
	}

	// Generate signature
	signature := p.generateSignature(uMoneyRequest)
	uMoneyRequest["signature"] = signature

	// Convert to JSON
	jsonData, err := json.Marshal(uMoneyRequest)
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
	var uMoneyResponse map[string]interface{}
	if err := json.Unmarshal(body, &uMoneyResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if status, ok := uMoneyResponse["status"].(string); ok && status != "success" {
		errorMsg := "unknown error"
		if msg, ok := uMoneyResponse["message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("U-Money error: %s", errorMsg)
	}

	// Extract payment details
	data, ok := uMoneyResponse["data"].(map[string]interface{})
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
func (p *LaosUMoney) GetPaymentStatus(request *PaymentStatusRequest) (*PaymentResponse, error) {
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
		return nil, fmt.Errorf("U-Money error: %s", errorMsg)
	}

	// Extract payment details
	data, ok := statusResponse["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid response format")
	}

	paymentStatus, _ := data["status"].(string)
	amount, _ := data["amount"].(float64)
	paymentType, _ := data["payment_type"].(string)
	transactionID, _ := data["transaction_id"].(string)
	createdAtStr, _ := data["created_at"].(string)
	
	// Parse created at
	createdAt, _ := time.Parse("2006-01-02T15:04:05Z", createdAtStr)

	// Map U-Money status to our status
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
		Currency:      "LAK",
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
func (p *LaosUMoney) RefundPayment(request *RefundRequest) (*RefundResponse, error) {
	// Prepare refund request
	timestamp := time.Now().Format("20060102150405")
	
	refundRequest := map[string]interface{}{
		"merchant_id": p.config.MerchantID,
		"payment_id":  request.PaymentID,
		"refund_id":   request.RefundID,
		"amount":      int(request.Amount), // U-Money expects integer amount
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
		return nil, fmt.Errorf("U-Money refund error: %s", errorMsg)
	}

	// Extract refund details
	data, ok := refundResponse["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid response format")
	}

	refundID, _ := data["refund_id"].(string)
	refundStatus, _ := data["status"].(string)
	transactionID, _ := data["transaction_id"].(string)

	// Create response
	response := &RefundResponse{
		RefundID:      refundID,
		PaymentID:     request.PaymentID,
		Amount:        request.Amount,
		Currency:      "LAK",
		Status:        refundStatus,
		TransactionID: transactionID,
		CreatedAt:     time.Now(),
	}

	return response, nil
}

// generateSignature generates a signature for U-Money requests
func (p *LaosUMoney) generateSignature(params map[string]interface{}) string {
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

// LaosLDBConfig holds configuration for LDB integration
type LaosLDBConfig struct {
	MerchantID     string
	MerchantSecret string
	APIEndpoint    string
	CallbackURL    string
	RedirectURL    string
	TestMode       bool
}

// LaosLDB implements PaymentPlatform interface for Laos's LDB (Lao Development Bank)
type LaosLDB struct {
	config LaosLDBConfig
	client *http.Client
}

// NewLaosLDB creates a new LDB payment platform
func NewLaosLDB(config LaosLDBConfig) *LaosLDB {
	// Set default API endpoint if not provided
	if config.APIEndpoint == "" {
		if config.TestMode {
			config.APIEndpoint = "https://sandbox-api.ldb.la"
		} else {
			config.APIEndpoint = "https://api.ldb.la"
		}
	}

	return &LaosLDB{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the name of the payment platform
func (p *LaosLDB) GetName() string {
	return "LDB"
}

// GetCountryCode returns the country code of the payment platform
func (p *LaosLDB) GetCountryCode() string {
	return "LA"
}

// GetSupportedPaymentMethods returns the supported payment methods
func (p *LaosLDB) GetSupportedPaymentMethods() []PaymentMethod {
	return []PaymentMethod{MethodBankTransfer, MethodQRCode}
}

// GetSupportedCurrencies returns the supported currencies
func (p *LaosLDB) GetSupportedCurrencies() []string {
	return []string{"LAK"}
}

// CreatePayment creates a payment
func (p *LaosLDB) CreatePayment(request *PaymentRequest) (*PaymentResponse, error) {
	// Validate request
	if request.Currency != "LAK" {
		return nil, errors.New("currency must be LAK for LDB payments")
	}

	if request.PaymentMethod != MethodBankTransfer && request.PaymentMethod != MethodQRCode {
		return nil, ErrUnsupportedPaymentMethod
	}

	// Prepare LDB request
	timestamp := time.Now().Format("20060102150405")
	
	ldbRequest := map[string]interface{}{
		"merchant_id":     p.config.MerchantID,
		"order_id":        request.OrderID,
		"amount":          int(request.Amount), // LDB expects integer amount
		"description":     request.Description,
		"customer_name":   request.CustomerName,
		"customer_email":  request.CustomerEmail,
		"customer_phone":  request.CustomerPhone,
		"return_url":      p.config.RedirectURL,
		"callback_url":    p.config.CallbackURL,
		"timestamp":       timestamp,
	}

	// Set payment method
	if request.PaymentMethod == MethodBankTransfer {
		ldbRequest["payment_type"] = "bank"
	} else {
		ldbRequest["payment_type"] = "qr"
	}

	// Generate signature
	signature := p.generateSignature(ldbRequest)
	ldbRequest["signature"] = signature

	// Convert to JSON
	jsonData, err := json.Marshal(ldbRequest)
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
	var ldbResponse map[string]interface{}
	if err := json.Unmarshal(body, &ldbResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if status, ok := ldbResponse["status"].(string); ok && status != "success" {
		errorMsg := "unknown error"
		if msg, ok := ldbResponse["message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("LDB error: %s", errorMsg)
	}

	// Extract payment details
	data, ok := ldbResponse["data"].(map[string]interface{})
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
func (p *LaosLDB) GetPaymentStatus(request *PaymentStatusRequest) (*PaymentResponse, error) {
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
		return nil, fmt.Errorf("LDB error: %s", errorMsg)
	}

	// Extract payment details
	data, ok := statusResponse["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid response format")
	}

	paymentStatus, _ := data["status"].(string)
	amount, _ := data["amount"].(float64)
	paymentType, _ := data["payment_type"].(string)
	transactionID, _ := data["transaction_id"].(string)
	createdAtStr, _ := data["created_at"].(string)
	
	// Parse created at
	createdAt, _ := time.Parse("2006-01-02T15:04:05Z", createdAtStr)

	// Map LDB status to our status
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
	if paymentType == "bank" {
		method = MethodBankTransfer
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
		Currency:      "LAK",
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
func (p *LaosLDB) RefundPayment(request *RefundRequest) (*RefundResponse, error) {
	// Prepare refund request
	timestamp := time.Now().Format("20060102150405")
	
	refundRequest := map[string]interface{}{
		"merchant_id": p.config.MerchantID,
		"payment_id":  request.PaymentID,
		"refund_id":   request.RefundID,
		"amount":      int(request.Amount), // LDB expects integer amount
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
		return nil, fmt.Errorf("LDB refund error: %s", errorMsg)
	}

	// Extract refund details
	data, ok := refundResponse["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid response format")
	}

	refundID, _ := data["refund_id"].(string)
	refundStatus, _ := data["status"].(string)
	transactionID, _ := data["transaction_id"].(string)

	// Create response
	response := &RefundResponse{
		RefundID:      refundID,
		PaymentID:     request.PaymentID,
		Amount:        request.Amount,
		Currency:      "LAK",
		Status:        refundStatus,
		TransactionID: transactionID,
		CreatedAt:     time.Now(),
	}

	return response, nil
}

// generateSignature generates a signature for LDB requests
func (p *LaosLDB) generateSignature(params map[string]interface{}) string {
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
	h := hmac.New(sha256.New, []byte(p.config.MerchantSecret))
	h.Write([]byte(signStr))
	return hex.EncodeToString(h.Sum(nil))
}
