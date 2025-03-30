// Thailand payment platform integrations for Asian Cryptocurrency Payment System
// Implements integrations with popular Thai payment platforms

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

// ThailandPromptPayConfig holds configuration for PromptPay integration
type ThailandPromptPayConfig struct {
	MerchantID     string
	MerchantKey    string
	APIEndpoint    string
	CallbackURL    string
	RedirectURL    string
	TestMode       bool
}

// ThailandPromptPay implements PaymentPlatform interface for Thailand's PromptPay
type ThailandPromptPay struct {
	config ThailandPromptPayConfig
	client *http.Client
}

// NewThailandPromptPay creates a new PromptPay payment platform
func NewThailandPromptPay(config ThailandPromptPayConfig) *ThailandPromptPay {
	// Set default API endpoint if not provided
	if config.APIEndpoint == "" {
		if config.TestMode {
			config.APIEndpoint = "https://api.sandbox.scb.co.th"
		} else {
			config.APIEndpoint = "https://api.scb.co.th"
		}
	}

	return &ThailandPromptPay{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the name of the payment platform
func (p *ThailandPromptPay) GetName() string {
	return "PromptPay"
}

// GetCountryCode returns the country code of the payment platform
func (p *ThailandPromptPay) GetCountryCode() string {
	return "TH"
}

// GetSupportedPaymentMethods returns the supported payment methods
func (p *ThailandPromptPay) GetSupportedPaymentMethods() []PaymentMethod {
	return []PaymentMethod{MethodQRCode, MethodBankTransfer}
}

// GetSupportedCurrencies returns the supported currencies
func (p *ThailandPromptPay) GetSupportedCurrencies() []string {
	return []string{"THB"}
}

// CreatePayment creates a payment
func (p *ThailandPromptPay) CreatePayment(request *PaymentRequest) (*PaymentResponse, error) {
	// Validate request
	if request.Currency != "THB" {
		return nil, errors.New("currency must be THB for PromptPay payments")
	}

	if request.PaymentMethod != MethodQRCode && request.PaymentMethod != MethodBankTransfer {
		return nil, ErrUnsupportedPaymentMethod
	}

	// Get access token
	token, err := p.getAccessToken()
	if err != nil {
		return nil, err
	}

	// Prepare PromptPay request
	promptPayRequest := map[string]interface{}{
		"qrType":        "PP",
		"ppType":        "BILLERID",
		"ppId":          p.config.MerchantID,
		"amount":        fmt.Sprintf("%.2f", request.Amount),
		"ref1":          request.OrderID,
		"ref2":          "PAYMENT",
		"ref3":          time.Now().Format("20060102150405"),
		"merchantId":    p.config.MerchantID,
		"terminalId":    "TERM001",
		"invoice":       request.OrderID,
		"description":   request.Description,
		"customerName":  request.CustomerName,
		"customerEmail": request.CustomerEmail,
		"customerPhone": request.CustomerPhone,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(promptPayRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", p.config.APIEndpoint+"/partners/sandbox/v1/payment/qrcode/create", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("resourceOwnerId", p.config.MerchantID)
	req.Header.Set("requestUId", uuid.New().String())
	req.Header.Set("channel", "scb_app")

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
	var promptPayResponse map[string]interface{}
	if err := json.Unmarshal(body, &promptPayResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if status, ok := promptPayResponse["status"].(map[string]interface{}); ok {
		if code, ok := status["code"].(float64); ok && code != 1000 {
			errorMsg := "unknown error"
			if msg, ok := status["description"].(string); ok {
				errorMsg = msg
			}
			return nil, fmt.Errorf("PromptPay error: %s", errorMsg)
		}
	}

	// Extract payment details
	data, ok := promptPayResponse["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid response format")
	}

	qrCode, _ := data["qrImage"].(string)
	qrCodeRaw, _ := data["qrRawData"].(string)
	transactionID, _ := data["transactionId"].(string)

	// Create response
	response := &PaymentResponse{
		PaymentID:     transactionID,
		Status:        StatusPending,
		Amount:        request.Amount,
		Currency:      request.Currency,
		PaymentMethod: request.PaymentMethod,
		QRCodeURL:     qrCode,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(15 * time.Minute),
		Metadata:      map[string]string{"qr_raw_data": qrCodeRaw},
	}

	return response, nil
}

// GetPaymentStatus gets the status of a payment
func (p *ThailandPromptPay) GetPaymentStatus(request *PaymentStatusRequest) (*PaymentResponse, error) {
	// Get access token
	token, err := p.getAccessToken()
	if err != nil {
		return nil, err
	}

	// Create HTTP request
	req, err := http.NewRequest("GET", p.config.APIEndpoint+"/partners/sandbox/v1/payment/billpayment/transactions/"+request.PaymentID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("resourceOwnerId", p.config.MerchantID)
	req.Header.Set("requestUId", uuid.New().String())
	req.Header.Set("channel", "scb_app")

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
	if status, ok := statusResponse["status"].(map[string]interface{}); ok {
		if code, ok := status["code"].(float64); ok && code != 1000 {
			errorMsg := "unknown error"
			if msg, ok := status["description"].(string); ok {
				errorMsg = msg
			}
			return nil, fmt.Errorf("PromptPay error: %s", errorMsg)
		}
	}

	// Extract payment details
	data, ok := statusResponse["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid response format")
	}

	transactionStatus, _ := data["transactionStatus"].(string)
	amountStr, _ := data["amount"].(string)
	amount, _ := strconv.ParseFloat(amountStr, 64)
	transactionDateStr, _ := data["transactionDate"].(string)
	
	// Parse transaction date
	transactionDate, _ := time.Parse("2006-01-02T15:04:05-07:00", transactionDateStr)

	// Map PromptPay status to our status
	status := StatusPending
	var completedAt time.Time

	switch transactionStatus {
	case "SUCCESS":
		status = StatusCompleted
		completedAt = transactionDate
	case "FAILED":
		status = StatusFailed
	case "PENDING":
		status = StatusPending
	}

	// Create response
	response := &PaymentResponse{
		PaymentID:     request.PaymentID,
		Status:        status,
		Amount:        amount,
		Currency:      "THB",
		PaymentMethod: MethodQRCode,
		TransactionID: request.PaymentID,
		CreatedAt:     transactionDate,
		UpdatedAt:     time.Now(),
		CompletedAt:   completedAt,
		Metadata:      make(map[string]string),
	}

	return response, nil
}

// RefundPayment refunds a payment
func (p *ThailandPromptPay) RefundPayment(request *RefundRequest) (*RefundResponse, error) {
	// Get access token
	token, err := p.getAccessToken()
	if err != nil {
		return nil, err
	}

	// Prepare refund request
	refundRequest := map[string]interface{}{
		"transactionId": request.PaymentID,
		"amount":        fmt.Sprintf("%.2f", request.Amount),
		"ref1":          request.RefundID,
		"ref2":          "REFUND",
		"ref3":          time.Now().Format("20060102150405"),
	}

	// Convert to JSON
	jsonData, err := json.Marshal(refundRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", p.config.APIEndpoint+"/partners/sandbox/v1/payment/billpayment/refund", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("resourceOwnerId", p.config.MerchantID)
	req.Header.Set("requestUId", uuid.New().String())
	req.Header.Set("channel", "scb_app")

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
	if status, ok := refundResponse["status"].(map[string]interface{}); ok {
		if code, ok := status["code"].(float64); ok && code != 1000 {
			errorMsg := "unknown error"
			if msg, ok := status["description"].(string); ok {
				errorMsg = msg
			}
			return nil, fmt.Errorf("PromptPay refund error: %s", errorMsg)
		}
	}

	// Extract refund details
	data, ok := refundResponse["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid response format")
	}

	refundID, _ := data["refundId"].(string)
	refundStatus, _ := data["status"].(string)

	// Create response
	response := &RefundResponse{
		RefundID:     refundID,
		PaymentID:    request.PaymentID,
		Amount:       request.Amount,
		Currency:     "THB",
		Status:       refundStatus,
		CreatedAt:    time.Now(),
	}

	return response, nil
}

// getAccessToken gets an access token for PromptPay API
func (p *ThailandPromptPay) getAccessToken() (string, error) {
	// Prepare token request
	tokenRequest := map[string]string{
		"applicationKey":    p.config.MerchantID,
		"applicationSecret": p.config.MerchantKey,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(tokenRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", p.config.APIEndpoint+"/partners/sandbox/v1/oauth/token", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("resourceOwnerId", p.config.MerchantID)
	req.Header.Set("requestUId", uuid.New().String())

	// Make API request
	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make token request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token response: %w", err)
	}

	// Parse response
	var tokenResponse map[string]interface{}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	// Check for errors
	if status, ok := tokenResponse["status"].(map[string]interface{}); ok {
		if code, ok := status["code"].(float64); ok && code != 1000 {
			errorMsg := "unknown error"
			if msg, ok := status["description"].(string); ok {
				errorMsg = msg
			}
			return "", fmt.Errorf("PromptPay token error: %s", errorMsg)
		}
	}

	// Extract token
	data, ok := tokenResponse["data"].(map[string]interface{})
	if !ok {
		return "", errors.New("invalid token response format")
	}

	accessToken, ok := data["accessToken"].(string)
	if !ok {
		return "", errors.New("failed to get access token")
	}

	return accessToken, nil
}

// ThailandTrueMoneyConfig holds configuration for TrueMoney integration
type ThailandTrueMoneyConfig struct {
	MerchantID     string
	MerchantKey    string
	APIEndpoint    string
	CallbackURL    string
	RedirectURL    string
	TestMode       bool
}

// ThailandTrueMoney implements PaymentPlatform interface for Thailand's TrueMoney
type ThailandTrueMoney struct {
	config ThailandTrueMoneyConfig
	client *http.Client
}

// NewThailandTrueMoney creates a new TrueMoney payment platform
func NewThailandTrueMoney(config ThailandTrueMoneyConfig) *ThailandTrueMoney {
	// Set default API endpoint if not provided
	if config.APIEndpoint == "" {
		if config.TestMode {
			config.APIEndpoint = "https://api-sandbox.truemoney.com"
		} else {
			config.APIEndpoint = "https://api.truemoney.com"
		}
	}

	return &ThailandTrueMoney{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the name of the payment platform
func (p *ThailandTrueMoney) GetName() string {
	return "TrueMoney"
}

// GetCountryCode returns the country code of the payment platform
func (p *ThailandTrueMoney) GetCountryCode() string {
	return "TH"
}

// GetSupportedPaymentMethods returns the supported payment methods
func (p *ThailandTrueMoney) GetSupportedPaymentMethods() []PaymentMethod {
	return []PaymentMethod{MethodEWallet, MethodQRCode}
}

// GetSupportedCurrencies returns the supported currencies
func (p *ThailandTrueMoney) GetSupportedCurrencies() []string {
	return []string{"THB"}
}

// CreatePayment creates a payment
func (p *ThailandTrueMoney) CreatePayment(request *PaymentRequest) (*PaymentResponse, error) {
	// Validate request
	if request.Currency != "THB" {
		return nil, errors.New("currency must be THB for TrueMoney payments")
	}

	if request.PaymentMethod != MethodEWallet && request.PaymentMethod != MethodQRCode {
		return nil, ErrUnsupportedPaymentMethod
	}

	// Prepare TrueMoney request
	timestamp := time.Now().Format("20060102150405")
	
	trueMoneyRequest := map[string]interface{}{
		"merchant_id":     p.config.MerchantID,
		"order_id":        request.OrderID,
		"amount":          fmt.Sprintf("%.2f", request.Amount),
		"currency":        request.Currency,
		"payment_method":  "wallet",
		"description":     request.Description,
		"customer_email":  request.CustomerEmail,
		"customer_phone":  request.CustomerPhone,
		"return_url":      p.config.RedirectURL,
		"notify_url":      p.config.CallbackURL,
		"timestamp":       timestamp,
	}

	// Generate signature
	signature := p.generateSignature(trueMoneyRequest)
	trueMoneyRequest["signature"] = signature

	// Convert to JSON
	jsonData, err := json.Marshal(trueMoneyRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", p.config.APIEndpoint+"/payments/v1/payment", bytes.NewBuffer(jsonData))
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
	var trueMoneyResponse map[string]interface{}
	if err := json.Unmarshal(body, &trueMoneyResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if status, ok := trueMoneyResponse["status"].(string); ok && status != "success" {
		errorMsg := "unknown error"
		if msg, ok := trueMoneyResponse["message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("TrueMoney error: %s", errorMsg)
	}

	// Extract payment details
	data, ok := trueMoneyResponse["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid response format")
	}

	paymentID, _ := data["payment_id"].(string)
	paymentURL, _ := data["payment_url"].(string)
	qrCodeURL, _ := data["qr_code"].(string)

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
		ExpiresAt:     time.Now().Add(15 * time.Minute),
		Metadata:      make(map[string]string),
	}

	return response, nil
}

// GetPaymentStatus gets the status of a payment
func (p *ThailandTrueMoney) GetPaymentStatus(request *PaymentStatusRequest) (*PaymentResponse, error) {
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
	req, err := http.NewRequest("POST", p.config.APIEndpoint+"/payments/v1/payment/status", bytes.NewBuffer(jsonData))
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
		return nil, fmt.Errorf("TrueMoney error: %s", errorMsg)
	}

	// Extract payment details
	data, ok := statusResponse["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid response format")
	}

	paymentStatus, _ := data["status"].(string)
	amountStr, _ := data["amount"].(string)
	amount, _ := strconv.ParseFloat(amountStr, 64)
	paymentMethod, _ := data["payment_method"].(string)
	transactionID, _ := data["transaction_id"].(string)
	createdAtStr, _ := data["created_at"].(string)
	
	// Parse created at
	createdAt, _ := time.Parse("2006-01-02T15:04:05Z", createdAtStr)

	// Map TrueMoney status to our status
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

	// Map payment method
	var method PaymentMethod
	if paymentMethod == "wallet" {
		method = MethodEWallet
	} else if paymentMethod == "qr" {
		method = MethodQRCode
	} else {
		method = request.PaymentMethod
	}

	// Create response
	response := &PaymentResponse{
		PaymentID:     request.PaymentID,
		Status:        status,
		Amount:        amount,
		Currency:      "THB",
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
func (p *ThailandTrueMoney) RefundPayment(request *RefundRequest) (*RefundResponse, error) {
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
	req, err := http.NewRequest("POST", p.config.APIEndpoint+"/payments/v1/payment/refund", bytes.NewBuffer(jsonData))
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
		return nil, fmt.Errorf("TrueMoney refund error: %s", errorMsg)
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
		Currency:      "THB",
		Status:        refundStatus,
		TransactionID: transactionID,
		CreatedAt:     time.Now(),
	}

	return response, nil
}

// generateSignature generates a signature for TrueMoney requests
func (p *ThailandTrueMoney) generateSignature(params map[string]interface{}) string {
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
