// Singapore payment platform integrations for Asian Cryptocurrency Payment System
// Implements integrations with popular Singaporean payment platforms

package integration

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// SingaporePayNowConfig holds configuration for PayNow integration
type SingaporePayNowConfig struct {
	MerchantID     string
	MerchantKey    string
	APIEndpoint    string
	CallbackURL    string
	RedirectURL    string
	TestMode       bool
}

// SingaporePayNow implements PaymentPlatform interface for Singapore's PayNow
type SingaporePayNow struct {
	config SingaporePayNowConfig
	client *http.Client
}

// NewSingaporePayNow creates a new PayNow payment platform
func NewSingaporePayNow(config SingaporePayNowConfig) *SingaporePayNow {
	// Set default API endpoint if not provided
	if config.APIEndpoint == "" {
		if config.TestMode {
			config.APIEndpoint = "https://api.sandbox.paynow.sg"
		} else {
			config.APIEndpoint = "https://api.paynow.sg"
		}
	}

	return &SingaporePayNow{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the name of the payment platform
func (p *SingaporePayNow) GetName() string {
	return "PayNow"
}

// GetCountryCode returns the country code of the payment platform
func (p *SingaporePayNow) GetCountryCode() string {
	return "SG"
}

// GetSupportedPaymentMethods returns the supported payment methods
func (p *SingaporePayNow) GetSupportedPaymentMethods() []PaymentMethod {
	return []PaymentMethod{MethodBankTransfer, MethodQRCode}
}

// GetSupportedCurrencies returns the supported currencies
func (p *SingaporePayNow) GetSupportedCurrencies() []string {
	return []string{"SGD"}
}

// CreatePayment creates a payment
func (p *SingaporePayNow) CreatePayment(request *PaymentRequest) (*PaymentResponse, error) {
	// Validate request
	if request.Currency != "SGD" {
		return nil, errors.New("currency must be SGD for PayNow payments")
	}

	if request.PaymentMethod != MethodBankTransfer && request.PaymentMethod != MethodQRCode {
		return nil, ErrUnsupportedPaymentMethod
	}

	// Prepare PayNow request
	payNowRequest := map[string]interface{}{
		"merchant_id":    p.config.MerchantID,
		"amount":         fmt.Sprintf("%.2f", request.Amount),
		"order_id":       request.OrderID,
		"description":    request.Description,
		"callback_url":   p.config.CallbackURL,
		"redirect_url":   p.config.RedirectURL,
		"customer_name":  request.CustomerName,
		"customer_email": request.CustomerEmail,
		"customer_phone": request.CustomerPhone,
		"timestamp":      time.Now().Unix(),
		"payment_type":   "paynow",
	}

	// Generate signature
	signature := p.generateSignature(payNowRequest)
	payNowRequest["signature"] = signature

	// Convert to JSON
	jsonData, err := json.Marshal(payNowRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make API request
	resp, err := p.client.Post(
		p.config.APIEndpoint+"/api/v1/payment/create",
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
	var payNowResponse map[string]interface{}
	if err := json.Unmarshal(body, &payNowResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if status, ok := payNowResponse["status"].(string); ok && status != "success" {
		errorMsg := "unknown error"
		if msg, ok := payNowResponse["message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("PayNow error: %s", errorMsg)
	}

	// Extract payment details
	data, ok := payNowResponse["data"].(map[string]interface{})
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
func (p *SingaporePayNow) GetPaymentStatus(request *PaymentStatusRequest) (*PaymentResponse, error) {
	// Prepare status request
	statusRequest := map[string]interface{}{
		"merchant_id": p.config.MerchantID,
		"payment_id":  request.PaymentID,
		"timestamp":   time.Now().Unix(),
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
		p.config.APIEndpoint+"/api/v1/payment/status",
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

	// Check for errors
	if status, ok := statusResponse["status"].(string); ok && status != "success" {
		errorMsg := "unknown error"
		if msg, ok := statusResponse["message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("PayNow error: %s", errorMsg)
	}

	// Extract payment details
	data, ok := statusResponse["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid response format")
	}

	paymentStatus, _ := data["status"].(string)
	amountStr, _ := data["amount"].(string)
	amount, _ := strconv.ParseFloat(amountStr, 64)
	transactionID, _ := data["transaction_id"].(string)
	createdAtUnix, _ := data["created_at"].(float64)
	updatedAtUnix, _ := data["updated_at"].(float64)

	// Parse timestamps
	createdAt := time.Unix(int64(createdAtUnix), 0)
	updatedAt := time.Unix(int64(updatedAtUnix), 0)

	// Map PayNow status to our status
	status := StatusPending
	var completedAt time.Time

	switch paymentStatus {
	case "completed", "success":
		status = StatusCompleted
		completedAt = updatedAt
	case "failed":
		status = StatusFailed
	case "cancelled":
		status = StatusCancelled
	}

	// Create response
	response := &PaymentResponse{
		PaymentID:     request.PaymentID,
		Status:        status,
		Amount:        amount,
		Currency:      "SGD",
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
func (p *SingaporePayNow) RefundPayment(request *RefundRequest) (*RefundResponse, error) {
	// Prepare refund request
	refundRequest := map[string]interface{}{
		"merchant_id": p.config.MerchantID,
		"payment_id":  request.PaymentID,
		"refund_id":   request.RefundID,
		"amount":      fmt.Sprintf("%.2f", request.Amount),
		"reason":      request.Reason,
		"timestamp":   time.Now().Unix(),
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
		p.config.APIEndpoint+"/api/v1/payment/refund",
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
		return nil, fmt.Errorf("PayNow refund error: %s", errorMsg)
	}

	// Extract refund details
	data, ok := refundResponse["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid response format")
	}

	refundID, _ := data["refund_id"].(string)
	status, _ := data["status"].(string)
	transactionID, _ := data["transaction_id"].(string)

	// Create response
	response := &RefundResponse{
		RefundID:      refundID,
		PaymentID:     request.PaymentID,
		Amount:        request.Amount,
		Currency:      "SGD",
		Status:        status,
		TransactionID: transactionID,
		CreatedAt:     time.Now(),
	}

	return response, nil
}

// generateSignature generates a signature for PayNow requests
func (p *SingaporePayNow) generateSignature(params map[string]interface{}) string {
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

// SingaporeNetsConfig holds configuration for NETS integration
type SingaporeNetsConfig struct {
	MerchantID     string
	MerchantKey    string
	APIEndpoint    string
	CallbackURL    string
	RedirectURL    string
	TestMode       bool
}

// SingaporeNets implements PaymentPlatform interface for Singapore's NETS
type SingaporeNets struct {
	config SingaporeNetsConfig
	client *http.Client
}

// NewSingaporeNets creates a new NETS payment platform
func NewSingaporeNets(config SingaporeNetsConfig) *SingaporeNets {
	// Set default API endpoint if not provided
	if config.APIEndpoint == "" {
		if config.TestMode {
			config.APIEndpoint = "https://api.sandbox.nets.com.sg"
		} else {
			config.APIEndpoint = "https://api.nets.com.sg"
		}
	}

	return &SingaporeNets{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the name of the payment platform
func (p *SingaporeNets) GetName() string {
	return "NETS"
}

// GetCountryCode returns the country code of the payment platform
func (p *SingaporeNets) GetCountryCode() string {
	return "SG"
}

// GetSupportedPaymentMethods returns the supported payment methods
func (p *SingaporeNets) GetSupportedPaymentMethods() []PaymentMethod {
	return []PaymentMethod{MethodCreditCard, MethodQRCode}
}

// GetSupportedCurrencies returns the supported currencies
func (p *SingaporeNets) GetSupportedCurrencies() []string {
	return []string{"SGD"}
}

// CreatePayment creates a payment
func (p *SingaporeNets) CreatePayment(request *PaymentRequest) (*PaymentResponse, error) {
	// Validate request
	if request.Currency != "SGD" {
		return nil, errors.New("currency must be SGD for NETS payments")
	}

	if request.PaymentMethod != MethodCreditCard && request.PaymentMethod != MethodQRCode {
		return nil, ErrUnsupportedPaymentMethod
	}

	// Prepare NETS request
	netsRequest := map[string]interface{}{
		"merchant_id":    p.config.MerchantID,
		"amount":         int(request.Amount * 100), // Convert to cents
		"order_id":       request.OrderID,
		"description":    request.Description,
		"callback_url":   p.config.CallbackURL,
		"redirect_url":   p.config.RedirectURL,
		"customer_name":  request.CustomerName,
		"customer_email": request.CustomerEmail,
		"customer_phone": request.CustomerPhone,
		"timestamp":      time.Now().Unix(),
	}

	// Set payment method
	if request.PaymentMethod == MethodCreditCard {
		netsRequest["payment_type"] = "card"
	} else {
		netsRequest["payment_type"] = "qr"
	}

	// Generate signature
	signature := p.generateSignature(netsRequest)
	netsRequest["signature"] = signature

	// Convert to JSON
	jsonData, err := json.Marshal(netsRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make API request
	resp, err := p.client.Post(
		p.config.APIEndpoint+"/api/v1/payment/create",
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
	var netsResponse map[string]interface{}
	if err := json.Unmarshal(body, &netsResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if status, ok := netsResponse["status"].(string); ok && status != "success" {
		errorMsg := "unknown error"
		if msg, ok := netsResponse["message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("NETS error: %s", errorMsg)
	}

	// Extract payment details
	data, ok := netsResponse["data"].(map[string]interface{})
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
func (p *SingaporeNets) GetPaymentStatus(request *PaymentStatusRequest) (*PaymentResponse, error) {
	// Prepare status request
	statusRequest := map[string]interface{}{
		"merchant_id": p.config.MerchantID,
		"payment_id":  request.PaymentID,
		"timestamp":   time.Now().Unix(),
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
		p.config.APIEndpoint+"/api/v1/payment/status",
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

	// Check for errors
	if status, ok := statusResponse["status"].(string); ok && status != "success" {
		errorMsg := "unknown error"
		if msg, ok := statusResponse["message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("NETS error: %s", errorMsg)
	}

	// Extract payment details
	data, ok := statusResponse["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid response format")
	}

	paymentStatus, _ := data["status"].(string)
	amountCents, _ := data["amount"].(float64)
	amount := amountCents / 100 // Convert from cents
	paymentType, _ := data["payment_type"].(string)
	transactionID, _ := data["transaction_id"].(string)
	createdAtUnix, _ := data["created_at"].(float64)
	updatedAtUnix, _ := data["updated_at"].(float64)

	// Parse timestamps
	createdAt := time.Unix(int64(createdAtUnix), 0)
	updatedAt := time.Unix(int64(updatedAtUnix), 0)

	// Map NETS status to our status
	status := StatusPending
	var completedAt time.Time

	switch paymentStatus {
	case "completed", "success":
		status = StatusCompleted
		completedAt = updatedAt
	case "failed":
		status = StatusFailed
	case "cancelled":
		status = StatusCancelled
	}

	// Map payment type to payment method
	paymentMethod := MethodCreditCard
	if paymentType == "qr" {
		paymentMethod = MethodQRCode
	}

	// Create response
	response := &PaymentResponse{
		PaymentID:     request.PaymentID,
		Status:        status,
		Amount:        amount,
		Currency:      "SGD",
		PaymentMethod: paymentMethod,
		TransactionID: transactionID,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
		CompletedAt:   completedAt,
		Metadata:      make(map[string]string),
	}

	return response, nil
}

// RefundPayment refunds a payment
func (p *SingaporeNets) RefundPayment(request *RefundRequest) (*RefundResponse, error) {
	// Prepare refund request
	refundRequest := map[string]interface{}{
		"merchant_id": p.config.MerchantID,
		"payment_id":  request.PaymentID,
		"refund_id":   request.RefundID,
		"amount":      int(request.Amount * 100), // Convert to cents
		"reason":      request.Reason,
		"timestamp":   time.Now().Unix(),
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
		p.config.APIEndpoint+"/api/v1/payment/refund",
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
		return nil, fmt.Errorf("NETS refund error: %s", errorMsg)
	}

	// Extract refund details
	data, ok := refundResponse["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid response format")
	}

	refundID, _ := data["refund_id"].(string)
	status, _ := data["status"].(string)
	transactionID, _ := data["transaction_id"].(string)

	// Create response
	response := &RefundResponse{
		RefundID:      refundID,
		PaymentID:     request.PaymentID,
		Amount:        request.Amount,
		Currency:      "SGD",
		Status:        status,
		TransactionID: transactionID,
		CreatedAt:     time.Now(),
	}

	return response, nil
}

// generateSignature generates a signature for NETS requests
func (p *SingaporeNets) generateSignature(params map[string]interface{}) string {
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
