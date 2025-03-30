// Indonesia payment platform integrations for Asian Cryptocurrency Payment System
// Implements integrations with popular Indonesian payment platforms

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

// IndonesiaGoPay holds configuration for GoPay integration
type IndonesiaGoPayConfig struct {
	ClientID       string
	ClientSecret   string
	MerchantID     string
	APIEndpoint    string
	CallbackURL    string
	RedirectURL    string
	TestMode       bool
}

// IndonesiaGoPay implements PaymentPlatform interface for Indonesia's GoPay
type IndonesiaGoPay struct {
	config IndonesiaGoPayConfig
	client *http.Client
}

// NewIndonesiaGoPay creates a new GoPay payment platform
func NewIndonesiaGoPay(config IndonesiaGoPayConfig) *IndonesiaGoPay {
	// Set default API endpoint if not provided
	if config.APIEndpoint == "" {
		if config.TestMode {
			config.APIEndpoint = "https://api.sandbox.midtrans.com"
		} else {
			config.APIEndpoint = "https://api.midtrans.com"
		}
	}

	return &IndonesiaGoPay{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the name of the payment platform
func (p *IndonesiaGoPay) GetName() string {
	return "GoPay"
}

// GetCountryCode returns the country code of the payment platform
func (p *IndonesiaGoPay) GetCountryCode() string {
	return "ID"
}

// GetSupportedPaymentMethods returns the supported payment methods
func (p *IndonesiaGoPay) GetSupportedPaymentMethods() []PaymentMethod {
	return []PaymentMethod{MethodEWallet, MethodQRCode}
}

// GetSupportedCurrencies returns the supported currencies
func (p *IndonesiaGoPay) GetSupportedCurrencies() []string {
	return []string{"IDR"}
}

// CreatePayment creates a payment
func (p *IndonesiaGoPay) CreatePayment(request *PaymentRequest) (*PaymentResponse, error) {
	// Validate request
	if request.Currency != "IDR" {
		return nil, errors.New("currency must be IDR for GoPay payments")
	}

	if request.PaymentMethod != MethodEWallet && request.PaymentMethod != MethodQRCode {
		return nil, ErrUnsupportedPaymentMethod
	}

	// Get access token
	token, err := p.getAccessToken()
	if err != nil {
		return nil, err
	}

	// Prepare GoPay request
	goPayRequest := map[string]interface{}{
		"transaction_details": map[string]interface{}{
			"order_id":     request.OrderID,
			"gross_amount": int(request.Amount),
		},
		"item_details": []map[string]interface{}{
			{
				"id":       "item1",
				"price":    int(request.Amount),
				"quantity": 1,
				"name":     request.Description,
			},
		},
		"customer_details": map[string]interface{}{
			"first_name": request.CustomerName,
			"email":      request.CustomerEmail,
			"phone":      request.CustomerPhone,
		},
		"payment_type": "gopay",
		"gopay": map[string]interface{}{
			"enable_callback": true,
			"callback_url":    p.config.CallbackURL,
		},
	}

	// Convert to JSON
	jsonData, err := json.Marshal(goPayRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", p.config.APIEndpoint+"/v2/charge", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+token)

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
	var goPayResponse map[string]interface{}
	if err := json.Unmarshal(body, &goPayResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if status, ok := goPayResponse["status_code"].(string); ok && status != "201" {
		errorMsg := "unknown error"
		if msg, ok := goPayResponse["status_message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("GoPay error: %s", errorMsg)
	}

	// Extract payment details
	transactionID, _ := goPayResponse["transaction_id"].(string)
	orderID, _ := goPayResponse["order_id"].(string)
	
	// Extract actions (payment URLs)
	var paymentURL, qrCodeURL string
	if actions, ok := goPayResponse["actions"].([]interface{}); ok {
		for _, action := range actions {
			if actionMap, ok := action.(map[string]interface{}); ok {
				if name, ok := actionMap["name"].(string); ok {
					if name == "deeplink-redirect" {
						paymentURL, _ = actionMap["url"].(string)
					} else if name == "generate-qr-code" {
						qrCodeURL, _ = actionMap["url"].(string)
					}
				}
			}
		}
	}

	// Create response
	response := &PaymentResponse{
		PaymentID:     transactionID,
		Status:        StatusPending,
		Amount:        request.Amount,
		Currency:      request.Currency,
		PaymentMethod: request.PaymentMethod,
		PaymentURL:    paymentURL,
		QRCodeURL:     qrCodeURL,
		RedirectURL:   paymentURL,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(24 * time.Hour),
		Metadata:      map[string]string{"order_id": orderID},
	}

	return response, nil
}

// GetPaymentStatus gets the status of a payment
func (p *IndonesiaGoPay) GetPaymentStatus(request *PaymentStatusRequest) (*PaymentResponse, error) {
	// Get access token
	token, err := p.getAccessToken()
	if err != nil {
		return nil, err
	}

	// Create HTTP request
	req, err := http.NewRequest("GET", p.config.APIEndpoint+"/v2/"+request.PaymentID+"/status", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+token)

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
	if status, ok := statusResponse["status_code"].(string); ok && status != "200" {
		errorMsg := "unknown error"
		if msg, ok := statusResponse["status_message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("GoPay error: %s", errorMsg)
	}

	// Extract payment details
	transactionID, _ := statusResponse["transaction_id"].(string)
	orderID, _ := statusResponse["order_id"].(string)
	transactionStatus, _ := statusResponse["transaction_status"].(string)
	grossAmount, _ := statusResponse["gross_amount"].(string)
	amount, _ := strconv.ParseFloat(grossAmount, 64)
	
	// Extract timestamps
	transactionTimeStr, _ := statusResponse["transaction_time"].(string)
	transactionTime, _ := time.Parse("2006-01-02 15:04:05", transactionTimeStr)
	
	// Map GoPay status to our status
	status := StatusPending
	var completedAt time.Time

	switch transactionStatus {
	case "settlement", "capture":
		status = StatusCompleted
		completedAt = time.Now()
	case "deny", "cancel", "expire":
		status = StatusFailed
	case "pending":
		status = StatusPending
	}

	// Create response
	response := &PaymentResponse{
		PaymentID:     transactionID,
		Status:        status,
		Amount:        amount,
		Currency:      "IDR",
		PaymentMethod: MethodEWallet,
		TransactionID: transactionID,
		CreatedAt:     transactionTime,
		UpdatedAt:     time.Now(),
		CompletedAt:   completedAt,
		Metadata:      map[string]string{"order_id": orderID},
	}

	return response, nil
}

// RefundPayment refunds a payment
func (p *IndonesiaGoPay) RefundPayment(request *RefundRequest) (*RefundResponse, error) {
	// Get access token
	token, err := p.getAccessToken()
	if err != nil {
		return nil, err
	}

	// Prepare refund request
	refundRequest := map[string]interface{}{
		"refund_key": request.RefundID,
		"amount":     int(request.Amount),
		"reason":     request.Reason,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(refundRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", p.config.APIEndpoint+"/v2/"+request.PaymentID+"/refund", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+token)

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
	if status, ok := refundResponse["status_code"].(string); ok && status != "200" {
		errorMsg := "unknown error"
		if msg, ok := refundResponse["status_message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("GoPay refund error: %s", errorMsg)
	}

	// Extract refund details
	refundKey, _ := refundResponse["refund_key"].(string)
	transactionID, _ := refundResponse["transaction_id"].(string)
	refundAmount, _ := refundResponse["refund_amount"].(string)
	amount, _ := strconv.ParseFloat(refundAmount, 64)

	// Create response
	response := &RefundResponse{
		RefundID:      refundKey,
		PaymentID:     request.PaymentID,
		Amount:        amount,
		Currency:      "IDR",
		Status:        "success",
		TransactionID: transactionID,
		CreatedAt:     time.Now(),
	}

	return response, nil
}

// getAccessToken gets an access token for GoPay API
func (p *IndonesiaGoPay) getAccessToken() (string, error) {
	// For Midtrans (GoPay), we use Basic Auth with Server Key
	auth := base64.StdEncoding.EncodeToString([]byte(p.config.ClientSecret + ":"))
	return auth, nil
}

// IndonesiaOVOConfig holds configuration for OVO integration
type IndonesiaOVOConfig struct {
	AppID         string
	AppKey        string
	MerchantID    string
	APIEndpoint   string
	CallbackURL   string
	RedirectURL   string
	TestMode      bool
}

// IndonesiaOVO implements PaymentPlatform interface for Indonesia's OVO
type IndonesiaOVO struct {
	config IndonesiaOVOConfig
	client *http.Client
}

// NewIndonesiaOVO creates a new OVO payment platform
func NewIndonesiaOVO(config IndonesiaOVOConfig) *IndonesiaOVO {
	// Set default API endpoint if not provided
	if config.APIEndpoint == "" {
		if config.TestMode {
			config.APIEndpoint = "https://api.sandbox.ovo.id"
		} else {
			config.APIEndpoint = "https://api.ovo.id"
		}
	}

	return &IndonesiaOVO{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the name of the payment platform
func (p *IndonesiaOVO) GetName() string {
	return "OVO"
}

// GetCountryCode returns the country code of the payment platform
func (p *IndonesiaOVO) GetCountryCode() string {
	return "ID"
}

// GetSupportedPaymentMethods returns the supported payment methods
func (p *IndonesiaOVO) GetSupportedPaymentMethods() []PaymentMethod {
	return []PaymentMethod{MethodEWallet}
}

// GetSupportedCurrencies returns the supported currencies
func (p *IndonesiaOVO) GetSupportedCurrencies() []string {
	return []string{"IDR"}
}

// CreatePayment creates a payment
func (p *IndonesiaOVO) CreatePayment(request *PaymentRequest) (*PaymentResponse, error) {
	// Validate request
	if request.Currency != "IDR" {
		return nil, errors.New("currency must be IDR for OVO payments")
	}

	if request.PaymentMethod != MethodEWallet {
		return nil, ErrUnsupportedPaymentMethod
	}

	// Prepare OVO request
	timestamp := time.Now().Format("20060102150405")
	reference := request.OrderID
	
	ovoRequest := map[string]interface{}{
		"reference_number": reference,
		"amount":           int(request.Amount),
		"phone":            request.CustomerPhone,
		"merchant_id":      p.config.MerchantID,
		"description":      request.Description,
		"callback_url":     p.config.CallbackURL,
		"timestamp":        timestamp,
	}

	// Generate signature
	signature := p.generateSignature(ovoRequest)
	ovoRequest["signature"] = signature

	// Convert to JSON
	jsonData, err := json.Marshal(ovoRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", p.config.APIEndpoint+"/api/v1/payment/push", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("App-ID", p.config.AppID)

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
	var ovoResponse map[string]interface{}
	if err := json.Unmarshal(body, &ovoResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if status, ok := ovoResponse["status"].(string); ok && status != "200" {
		errorMsg := "unknown error"
		if msg, ok := ovoResponse["message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("OVO error: %s", errorMsg)
	}

	// Extract payment details
	data, ok := ovoResponse["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid response format")
	}

	paymentID, _ := data["transaction_id"].(string)

	// Create response
	response := &PaymentResponse{
		PaymentID:     paymentID,
		Status:        StatusPending,
		Amount:        request.Amount,
		Currency:      request.Currency,
		PaymentMethod: request.PaymentMethod,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(15 * time.Minute),
		Metadata:      map[string]string{"reference": reference},
	}

	return response, nil
}

// GetPaymentStatus gets the status of a payment
func (p *IndonesiaOVO) GetPaymentStatus(request *PaymentStatusRequest) (*PaymentResponse, error) {
	// Prepare status request
	timestamp := time.Now().Format("20060102150405")
	
	statusRequest := map[string]interface{}{
		"transaction_id": request.PaymentID,
		"merchant_id":    p.config.MerchantID,
		"timestamp":      timestamp,
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
	req, err := http.NewRequest("POST", p.config.APIEndpoint+"/api/v1/payment/status", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("App-ID", p.config.AppID)

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
	if status, ok := statusResponse["status"].(string); ok && status != "200" {
		errorMsg := "unknown error"
		if msg, ok := statusResponse["message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("OVO error: %s", errorMsg)
	}

	// Extract payment details
	data, ok := statusResponse["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid response format")
	}

	paymentStatus, _ := data["status"].(string)
	amount, _ := data["amount"].(float64)
	reference, _ := data["reference_number"].(string)
	transactionTime, _ := data["transaction_time"].(string)
	
	// Parse transaction time
	createdAt, _ := time.Parse("2006-01-02 15:04:05", transactionTime)

	// Map OVO status to our status
	status := StatusPending
	var completedAt time.Time

	switch paymentStatus {
	case "SUCCESS":
		status = StatusCompleted
		completedAt = time.Now()
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
		Currency:      "IDR",
		PaymentMethod: MethodEWallet,
		TransactionID: request.PaymentID,
		CreatedAt:     createdAt,
		UpdatedAt:     time.Now(),
		CompletedAt:   completedAt,
		Metadata:      map[string]string{"reference": reference},
	}

	return response, nil
}

// RefundPayment refunds a payment
func (p *IndonesiaOVO) RefundPayment(request *RefundRequest) (*RefundResponse, error) {
	// Prepare refund request
	timestamp := time.Now().Format("20060102150405")
	
	refundRequest := map[string]interface{}{
		"transaction_id": request.PaymentID,
		"merchant_id":    p.config.MerchantID,
		"amount":         int(request.Amount),
		"reference":      request.RefundID,
		"reason":         request.Reason,
		"timestamp":      timestamp,
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
	req, err := http.NewRequest("POST", p.config.APIEndpoint+"/api/v1/payment/refund", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("App-ID", p.config.AppID)

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
	if status, ok := refundResponse["status"].(string); ok && status != "200" {
		errorMsg := "unknown error"
		if msg, ok := refundResponse["message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("OVO refund error: %s", errorMsg)
	}

	// Extract refund details
	data, ok := refundResponse["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid response format")
	}

	refundID, _ := data["refund_id"].(string)
	refundStatus, _ := data["status"].(string)

	// Create response
	response := &RefundResponse{
		RefundID:     refundID,
		PaymentID:    request.PaymentID,
		Amount:       request.Amount,
		Currency:     "IDR",
		Status:       refundStatus,
		CreatedAt:    time.Now(),
	}

	return response, nil
}

// generateSignature generates a signature for OVO requests
func (p *IndonesiaOVO) generateSignature(params map[string]interface{}) string {
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
	h := hmac.New(sha256.New, []byte(p.config.AppKey))
	h.Write([]byte(signStr))
	return hex.EncodeToString(h.Sum(nil))
}
