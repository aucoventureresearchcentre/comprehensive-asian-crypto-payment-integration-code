// Malaysia payment platform integrations for Asian Cryptocurrency Payment System
// Implements integrations with popular Malaysian payment platforms

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
	"sort"
	"strings"
	"time"
)

// MalaysiaFPXConfig holds configuration for FPX integration
type MalaysiaFPXConfig struct {
	MerchantID     string
	MerchantKey    string
	APIEndpoint    string
	CallbackURL    string
	RedirectURL    string
	TestMode       bool
}

// MalaysiaFPX implements PaymentPlatform interface for Malaysia's FPX
type MalaysiaFPX struct {
	config MalaysiaFPXConfig
	client *http.Client
}

// NewMalaysiaFPX creates a new FPX payment platform
func NewMalaysiaFPX(config MalaysiaFPXConfig) *MalaysiaFPX {
	// Set default API endpoint if not provided
	if config.APIEndpoint == "" {
		if config.TestMode {
			config.APIEndpoint = "https://sandbox.paynet.my/fpx"
		} else {
			config.APIEndpoint = "https://www.paynet.my/fpx"
		}
	}

	return &MalaysiaFPX{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the name of the payment platform
func (p *MalaysiaFPX) GetName() string {
	return "FPX"
}

// GetCountryCode returns the country code of the payment platform
func (p *MalaysiaFPX) GetCountryCode() string {
	return "MY"
}

// GetSupportedPaymentMethods returns the supported payment methods
func (p *MalaysiaFPX) GetSupportedPaymentMethods() []PaymentMethod {
	return []PaymentMethod{MethodBankTransfer}
}

// GetSupportedCurrencies returns the supported currencies
func (p *MalaysiaFPX) GetSupportedCurrencies() []string {
	return []string{"MYR"}
}

// CreatePayment creates a payment
func (p *MalaysiaFPX) CreatePayment(request *PaymentRequest) (*PaymentResponse, error) {
	// Validate request
	if request.Currency != "MYR" {
		return nil, errors.New("currency must be MYR for FPX payments")
	}

	if request.PaymentMethod != MethodBankTransfer {
		return nil, ErrUnsupportedPaymentMethod
	}

	// Prepare FPX request
	fpxRequest := map[string]string{
		"merchantId":       p.config.MerchantID,
		"orderNo":          request.OrderID,
		"amount":           fmt.Sprintf("%.2f", request.Amount),
		"customerName":     request.CustomerName,
		"customerEmail":    request.CustomerEmail,
		"description":      request.Description,
		"callbackUrl":      p.config.CallbackURL,
		"redirectUrl":      p.config.RedirectURL,
		"transactionTime":  time.Now().Format("20060102150405"),
		"testMode":         fmt.Sprintf("%t", p.config.TestMode),
	}

	// Generate signature
	signature := p.generateSignature(fpxRequest)
	fpxRequest["signature"] = signature

	// Convert to JSON
	jsonData, err := json.Marshal(fpxRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make API request
	resp, err := p.client.Post(
		p.config.APIEndpoint+"/api/payment",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
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
	var fpxResponse map[string]interface{}
	if err := json.Unmarshal(body, &fpxResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if status, ok := fpxResponse["status"].(string); ok && status != "success" {
		errorMsg := "unknown error"
		if msg, ok := fpxResponse["message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("FPX error: %s", errorMsg)
	}

	// Extract payment URL and ID
	paymentURL, _ := fpxResponse["paymentUrl"].(string)
	paymentID, _ := fpxResponse["paymentId"].(string)

	// Create response
	response := &PaymentResponse{
		PaymentID:     paymentID,
		Status:        StatusPending,
		Amount:        request.Amount,
		Currency:      request.Currency,
		PaymentMethod: request.PaymentMethod,
		PaymentURL:    paymentURL,
		RedirectURL:   paymentURL,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(1 * time.Hour),
		Metadata:      make(map[string]string),
	}

	return response, nil
}

// GetPaymentStatus gets the status of a payment
func (p *MalaysiaFPX) GetPaymentStatus(request *PaymentStatusRequest) (*PaymentResponse, error) {
	// Prepare status request
	statusRequest := map[string]string{
		"merchantId": p.config.MerchantID,
		"paymentId":  request.PaymentID,
	}

	// Generate signature
	signature := p.generateSignature(statusRequest)
	statusRequest["signature"] = signature

	// Convert to JSON
	jsonData, err := json.Marshal(statusRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make API request
	resp, err := p.client.Post(
		p.config.APIEndpoint+"/api/status",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
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

	// Extract payment details
	status, _ := statusResponse["status"].(string)
	amount, _ := statusResponse["amount"].(float64)
	transactionID, _ := statusResponse["transactionId"].(string)
	createdAtStr, _ := statusResponse["createdAt"].(string)
	updatedAtStr, _ := statusResponse["updatedAt"].(string)

	// Parse timestamps
	createdAt, _ := time.Parse(time.RFC3339, createdAtStr)
	updatedAt, _ := time.Parse(time.RFC3339, updatedAtStr)

	// Map FPX status to our status
	paymentStatus := StatusPending
	var completedAt time.Time

	switch status {
	case "PAYMENT_SUCCESSFUL":
		paymentStatus = StatusCompleted
		completedAt = updatedAt
	case "PAYMENT_FAILED":
		paymentStatus = StatusFailed
	case "PAYMENT_CANCELLED":
		paymentStatus = StatusCancelled
	}

	// Create response
	response := &PaymentResponse{
		PaymentID:     request.PaymentID,
		Status:        paymentStatus,
		Amount:        amount,
		Currency:      "MYR",
		PaymentMethod: MethodBankTransfer,
		TransactionID: transactionID,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
		CompletedAt:   completedAt,
		Metadata:      make(map[string]string),
	}

	return response, nil
}

// RefundPayment refunds a payment
func (p *MalaysiaFPX) RefundPayment(request *RefundRequest) (*RefundResponse, error) {
	// Prepare refund request
	refundRequest := map[string]string{
		"merchantId": p.config.MerchantID,
		"paymentId":  request.PaymentID,
		"refundId":   request.RefundID,
		"amount":     fmt.Sprintf("%.2f", request.Amount),
		"reason":     request.Reason,
	}

	// Generate signature
	signature := p.generateSignature(refundRequest)
	refundRequest["signature"] = signature

	// Convert to JSON
	jsonData, err := json.Marshal(refundRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make API request
	resp, err := p.client.Post(
		p.config.APIEndpoint+"/api/refund",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
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
		return nil, fmt.Errorf("FPX refund error: %s", errorMsg)
	}

	// Extract refund details
	refundID, _ := refundResponse["refundId"].(string)
	status, _ := refundResponse["status"].(string)
	transactionID, _ := refundResponse["transactionId"].(string)

	// Create response
	response := &RefundResponse{
		RefundID:      refundID,
		PaymentID:     request.PaymentID,
		Amount:        request.Amount,
		Currency:      "MYR",
		Status:        status,
		TransactionID: transactionID,
		CreatedAt:     time.Now(),
	}

	return response, nil
}

// generateSignature generates a signature for FPX requests
func (p *MalaysiaFPX) generateSignature(params map[string]string) string {
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
		sb.WriteString(params[k])
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

// MalaysiaGrabPayConfig holds configuration for GrabPay integration
type MalaysiaGrabPayConfig struct {
	MerchantID     string
	ClientID       string
	ClientSecret   string
	APIEndpoint    string
	CallbackURL    string
	RedirectURL    string
	TestMode       bool
}

// MalaysiaGrabPay implements PaymentPlatform interface for Malaysia's GrabPay
type MalaysiaGrabPay struct {
	config MalaysiaGrabPayConfig
	client *http.Client
}

// NewMalaysiaGrabPay creates a new GrabPay payment platform
func NewMalaysiaGrabPay(config MalaysiaGrabPayConfig) *MalaysiaGrabPay {
	// Set default API endpoint if not provided
	if config.APIEndpoint == "" {
		if config.TestMode {
			config.APIEndpoint = "https://partner-api.sandbox.grab.com"
		} else {
			config.APIEndpoint = "https://partner-api.grab.com"
		}
	}

	return &MalaysiaGrabPay{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the name of the payment platform
func (p *MalaysiaGrabPay) GetName() string {
	return "GrabPay"
}

// GetCountryCode returns the country code of the payment platform
func (p *MalaysiaGrabPay) GetCountryCode() string {
	return "MY"
}

// GetSupportedPaymentMethods returns the supported payment methods
func (p *MalaysiaGrabPay) GetSupportedPaymentMethods() []PaymentMethod {
	return []PaymentMethod{MethodEWallet, MethodQRCode}
}

// GetSupportedCurrencies returns the supported currencies
func (p *MalaysiaGrabPay) GetSupportedCurrencies() []string {
	return []string{"MYR"}
}

// CreatePayment creates a payment
func (p *MalaysiaGrabPay) CreatePayment(request *PaymentRequest) (*PaymentResponse, error) {
	// Validate request
	if request.Currency != "MYR" {
		return nil, errors.New("currency must be MYR for GrabPay payments")
	}

	if request.PaymentMethod != MethodEWallet && request.PaymentMethod != MethodQRCode {
		return nil, ErrUnsupportedPaymentMethod
	}

	// Get access token
	token, err := p.getAccessToken()
	if err != nil {
		return nil, err
	}

	// Prepare GrabPay request
	grabPayRequest := map[string]interface{}{
		"partnerTxID":       request.OrderID,
		"partnerGroupTxID":  request.OrderID,
		"amount":            int(request.Amount * 100), // Convert to cents
		"currency":          request.Currency,
		"description":       request.Description,
		"merchantID":        p.config.MerchantID,
		"metaInfo": map[string]interface{}{
			"customerName":  request.CustomerName,
			"customerEmail": request.CustomerEmail,
			"customerPhone": request.CustomerPhone,
		},
	}

	// Add redirect URLs
	if p.config.RedirectURL != "" {
		grabPayRequest["redirectURL"] = p.config.RedirectURL
	}
	if p.config.CallbackURL != "" {
		grabPayRequest["webhookURL"] = p.config.CallbackURL
	}

	// Convert to JSON
	jsonData, err := json.Marshal(grabPayRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", p.config.APIEndpoint+"/grabpay/partner/v2/charge/init", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-GID-AUX-POP", p.generatePOP(req.URL.Path, "POST", string(jsonData), token))

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
	var grabPayResponse map[string]interface{}
	if err := json.Unmarshal(body, &grabPayResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		errorMsg := "unknown error"
		if msg, ok := grabPayResponse["message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("GrabPay error: %s", errorMsg)
	}

	// Extract payment details
	paymentID, _ := grabPayResponse["txID"].(string)
	paymentURL, _ := grabPayResponse["request"].(string)
	qrCodeURL, _ := grabPayResponse["qrCodeURL"].(string)

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
func (p *MalaysiaGrabPay) GetPaymentStatus(request *PaymentStatusRequest) (*PaymentResponse, error) {
	// Get access token
	token, err := p.getAccessToken()
	if err != nil {
		return nil, err
	}

	// Create HTTP request
	req, err := http.NewRequest("GET", p.config.APIEndpoint+"/grabpay/partner/v2/charge/"+request.PaymentID+"/status", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-GID-AUX-POP", p.generatePOP(req.URL.Path, "GET", "", token))

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
	if resp.StatusCode != http.StatusOK {
		errorMsg := "unknown error"
		if msg, ok := statusResponse["message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("GrabPay error: %s", errorMsg)
	}

	// Extract payment details
	status, _ := statusResponse["status"].(string)
	amountCents, _ := statusResponse["amount"].(float64)
	amount := amountCents / 100 // Convert from cents
	currency, _ := statusResponse["currency"].(string)
	transactionID, _ := statusResponse["txID"].(string)

	// Map GrabPay status to our status
	paymentStatus := StatusPending
	var completedAt time.Time

	switch status {
	case "success", "completed":
		paymentStatus = StatusCompleted
		completedAt = time.Now()
	case "failed":
		paymentStatus = StatusFailed
	case "cancelled":
		paymentStatus = StatusCancelled
	}

	// Create response
	response := &PaymentResponse{
		PaymentID:     request.PaymentID,
		Status:        paymentStatus,
		Amount:        amount,
		Currency:      currency,
		PaymentMethod: MethodEWallet,
		TransactionID: transactionID,
		CreatedAt:     time.Now(), // We don't have the actual creation time
		UpdatedAt:     time.Now(),
		CompletedAt:   completedAt,
		Metadata:      make(map[string]string),
	}

	return response, nil
}

// RefundPayment refunds a payment
func (p *MalaysiaGrabPay) RefundPayment(request *RefundRequest) (*RefundResponse, error) {
	// Get access token
	token, err := p.getAccessToken()
	if err != nil {
		return nil, err
	}

	// Prepare refund request
	refundRequest := map[string]interface{}{
		"partnerTxID":      request.RefundID,
		"partnerGroupTxID": request.RefundID,
		"amount":           int(request.Amount * 100), // Convert to cents
		"currency":         "MYR",
		"txID":             request.PaymentID,
		"description":      request.Reason,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(refundRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", p.config.APIEndpoint+"/grabpay/partner/v2/refund", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-GID-AUX-POP", p.generatePOP(req.URL.Path, "POST", string(jsonData), token))

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
	if resp.StatusCode != http.StatusOK {
		errorMsg := "unknown error"
		if msg, ok := refundResponse["message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("GrabPay refund error: %s", errorMsg)
	}

	// Extract refund details
	refundID, _ := refundResponse["txID"].(string)
	status, _ := refundResponse["status"].(string)

	// Create response
	response := &RefundResponse{
		RefundID:     refundID,
		PaymentID:    request.PaymentID,
		Amount:       request.Amount,
		Currency:     "MYR",
		Status:       status,
		CreatedAt:    time.Now(),
	}

	return response, nil
}

// getAccessToken gets an access token for GrabPay API
func (p *MalaysiaGrabPay) getAccessToken() (string, error) {
	// Prepare token request
	tokenRequest := map[string]string{
		"client_id":     p.config.ClientID,
		"client_secret": p.config.ClientSecret,
		"grant_type":    "client_credentials",
	}

	// Convert to form data
	formData := url.Values{}
	for k, v := range tokenRequest {
		formData.Add(k, v)
	}

	// Make API request
	resp, err := p.client.PostForm(p.config.APIEndpoint+"/grabid/v1/oauth2/token", formData)
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

	// Extract token
	token, ok := tokenResponse["access_token"].(string)
	if !ok {
		return "", errors.New("failed to get access token")
	}

	return token, nil
}

// generatePOP generates a proof of possession for GrabPay API
func (p *MalaysiaGrabPay) generatePOP(path, method, body, token string) string {
	// Generate timestamp
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	// Generate nonce
	nonce := fmt.Sprintf("%d", time.Now().UnixNano())

	// Build string to sign
	signStr := method + "&" + path + "&" + timestamp + "&" + nonce + "&" + token + "&" + body

	// Generate HMAC-SHA256
	h := hmac.New(sha256.New, []byte(p.config.ClientSecret))
	h.Write([]byte(signStr))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Build POP
	pop := fmt.Sprintf("HS256 timestamp=%s,nonce=%s,signature=%s", timestamp, nonce, signature)

	return pop
}
