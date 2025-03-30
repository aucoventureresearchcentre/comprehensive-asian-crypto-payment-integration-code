// Vietnam payment platform integrations for Asian Cryptocurrency Payment System
// Implements integrations with popular Vietnamese payment platforms

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

// VietnamMoMoConfig holds configuration for MoMo integration
type VietnamMoMoConfig struct {
	PartnerCode   string
	AccessKey     string
	SecretKey     string
	APIEndpoint   string
	CallbackURL   string
	RedirectURL   string
	TestMode      bool
}

// VietnamMoMo implements PaymentPlatform interface for Vietnam's MoMo
type VietnamMoMo struct {
	config VietnamMoMoConfig
	client *http.Client
}

// NewVietnamMoMo creates a new MoMo payment platform
func NewVietnamMoMo(config VietnamMoMoConfig) *VietnamMoMo {
	// Set default API endpoint if not provided
	if config.APIEndpoint == "" {
		if config.TestMode {
			config.APIEndpoint = "https://test-payment.momo.vn"
		} else {
			config.APIEndpoint = "https://payment.momo.vn"
		}
	}

	return &VietnamMoMo{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the name of the payment platform
func (p *VietnamMoMo) GetName() string {
	return "MoMo"
}

// GetCountryCode returns the country code of the payment platform
func (p *VietnamMoMo) GetCountryCode() string {
	return "VN"
}

// GetSupportedPaymentMethods returns the supported payment methods
func (p *VietnamMoMo) GetSupportedPaymentMethods() []PaymentMethod {
	return []PaymentMethod{MethodEWallet, MethodQRCode}
}

// GetSupportedCurrencies returns the supported currencies
func (p *VietnamMoMo) GetSupportedCurrencies() []string {
	return []string{"VND"}
}

// CreatePayment creates a payment
func (p *VietnamMoMo) CreatePayment(request *PaymentRequest) (*PaymentResponse, error) {
	// Validate request
	if request.Currency != "VND" {
		return nil, errors.New("currency must be VND for MoMo payments")
	}

	if request.PaymentMethod != MethodEWallet && request.PaymentMethod != MethodQRCode {
		return nil, ErrUnsupportedPaymentMethod
	}

	// Prepare MoMo request
	requestID := fmt.Sprintf("%s_%d", request.OrderID, time.Now().UnixNano())
	orderInfo := fmt.Sprintf("Payment for order %s", request.OrderID)
	
	// Convert amount to integer (MoMo requires integer amount)
	amount := int64(request.Amount)
	
	// Prepare raw signature
	rawSignature := fmt.Sprintf("accessKey=%s&amount=%d&extraData=&ipnUrl=%s&orderId=%s&orderInfo=%s&partnerCode=%s&redirectUrl=%s&requestId=%s&requestType=captureMoMoWallet",
		p.config.AccessKey,
		amount,
		p.config.CallbackURL,
		request.OrderID,
		orderInfo,
		p.config.PartnerCode,
		p.config.RedirectURL,
		requestID,
	)
	
	// Generate signature
	h := hmac.New(sha256.New, []byte(p.config.SecretKey))
	h.Write([]byte(rawSignature))
	signature := hex.EncodeToString(h.Sum(nil))
	
	// Prepare MoMo request
	momoRequest := map[string]interface{}{
		"partnerCode": p.config.PartnerCode,
		"accessKey":   p.config.AccessKey,
		"requestId":   requestID,
		"amount":      amount,
		"orderId":     request.OrderID,
		"orderInfo":   orderInfo,
		"redirectUrl": p.config.RedirectURL,
		"ipnUrl":      p.config.CallbackURL,
		"extraData":   "",
		"requestType": "captureMoMoWallet",
		"signature":   signature,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(momoRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", p.config.APIEndpoint+"/v2/gateway/api/create", bytes.NewBuffer(jsonData))
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
	var momoResponse map[string]interface{}
	if err := json.Unmarshal(body, &momoResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if errorCode, ok := momoResponse["errorCode"].(float64); ok && errorCode != 0 {
		errorMsg := "unknown error"
		if msg, ok := momoResponse["message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("MoMo error: %s", errorMsg)
	}

	// Extract payment details
	paymentID, _ := momoResponse["orderId"].(string)
	payURL, _ := momoResponse["payUrl"].(string)
	qrCodeURL, _ := momoResponse["qrCodeUrl"].(string)

	// Create response
	response := &PaymentResponse{
		PaymentID:     paymentID,
		Status:        StatusPending,
		Amount:        float64(amount),
		Currency:      request.Currency,
		PaymentMethod: request.PaymentMethod,
		PaymentURL:    payURL,
		QRCodeURL:     qrCodeURL,
		RedirectURL:   payURL,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(15 * time.Minute),
		Metadata:      map[string]string{"request_id": requestID},
	}

	return response, nil
}

// GetPaymentStatus gets the status of a payment
func (p *VietnamMoMo) GetPaymentStatus(request *PaymentStatusRequest) (*PaymentResponse, error) {
	// Prepare status request
	requestID := fmt.Sprintf("status_%s_%d", request.PaymentID, time.Now().UnixNano())
	
	// Prepare raw signature
	rawSignature := fmt.Sprintf("accessKey=%s&orderId=%s&partnerCode=%s&requestId=%s",
		p.config.AccessKey,
		request.PaymentID,
		p.config.PartnerCode,
		requestID,
	)
	
	// Generate signature
	h := hmac.New(sha256.New, []byte(p.config.SecretKey))
	h.Write([]byte(rawSignature))
	signature := hex.EncodeToString(h.Sum(nil))
	
	// Prepare status request
	statusRequest := map[string]interface{}{
		"partnerCode": p.config.PartnerCode,
		"accessKey":   p.config.AccessKey,
		"requestId":   requestID,
		"orderId":     request.PaymentID,
		"requestType": "transactionStatus",
		"signature":   signature,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(statusRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", p.config.APIEndpoint+"/v2/gateway/api/query", bytes.NewBuffer(jsonData))
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
	if errorCode, ok := statusResponse["errorCode"].(float64); ok && errorCode != 0 {
		errorMsg := "unknown error"
		if msg, ok := statusResponse["message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("MoMo error: %s", errorMsg)
	}

	// Extract payment details
	amount, _ := statusResponse["amount"].(float64)
	transID, _ := statusResponse["transId"].(string)
	payType, _ := statusResponse["payType"].(string)
	responseTime, _ := statusResponse["responseTime"].(float64)
	
	// Map MoMo status to our status
	status := StatusPending
	var completedAt time.Time

	if payType == "3" || payType == "4" {
		status = StatusCompleted
		completedAt = time.Unix(int64(responseTime/1000), 0)
	} else if payType == "0" {
		status = StatusPending
	} else {
		status = StatusFailed
	}

	// Create response
	response := &PaymentResponse{
		PaymentID:     request.PaymentID,
		Status:        status,
		Amount:        amount,
		Currency:      "VND",
		PaymentMethod: MethodEWallet,
		TransactionID: transID,
		CreatedAt:     time.Now(), // MoMo doesn't provide creation time
		UpdatedAt:     time.Now(),
		CompletedAt:   completedAt,
		Metadata:      make(map[string]string),
	}

	return response, nil
}

// RefundPayment refunds a payment
func (p *VietnamMoMo) RefundPayment(request *RefundRequest) (*RefundResponse, error) {
	// Prepare refund request
	requestID := fmt.Sprintf("refund_%s_%d", request.PaymentID, time.Now().UnixNano())
	
	// Convert amount to integer (MoMo requires integer amount)
	amount := int64(request.Amount)
	
	// Prepare raw signature
	rawSignature := fmt.Sprintf("accessKey=%s&amount=%d&description=%s&orderId=%s&partnerCode=%s&requestId=%s&transId=%s",
		p.config.AccessKey,
		amount,
		request.Reason,
		request.PaymentID,
		p.config.PartnerCode,
		requestID,
		request.PaymentID, // Using payment ID as transaction ID
	)
	
	// Generate signature
	h := hmac.New(sha256.New, []byte(p.config.SecretKey))
	h.Write([]byte(rawSignature))
	signature := hex.EncodeToString(h.Sum(nil))
	
	// Prepare refund request
	refundRequest := map[string]interface{}{
		"partnerCode": p.config.PartnerCode,
		"accessKey":   p.config.AccessKey,
		"requestId":   requestID,
		"amount":      amount,
		"orderId":     request.PaymentID,
		"transId":     request.PaymentID, // Using payment ID as transaction ID
		"description": request.Reason,
		"signature":   signature,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(refundRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", p.config.APIEndpoint+"/v2/gateway/api/refund", bytes.NewBuffer(jsonData))
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
	if errorCode, ok := refundResponse["errorCode"].(float64); ok && errorCode != 0 {
		errorMsg := "unknown error"
		if msg, ok := refundResponse["message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("MoMo refund error: %s", errorMsg)
	}

	// Extract refund details
	refundID, _ := refundResponse["requestId"].(string)
	transID, _ := refundResponse["transId"].(string)

	// Create response
	response := &RefundResponse{
		RefundID:      refundID,
		PaymentID:     request.PaymentID,
		Amount:        request.Amount,
		Currency:      "VND",
		Status:        "success",
		TransactionID: transID,
		CreatedAt:     time.Now(),
	}

	return response, nil
}

// VietnamVNPayConfig holds configuration for VNPay integration
type VietnamVNPayConfig struct {
	MerchantID     string
	SecureHash     string
	APIEndpoint    string
	CallbackURL    string
	RedirectURL    string
	TestMode       bool
}

// VietnamVNPay implements PaymentPlatform interface for Vietnam's VNPay
type VietnamVNPay struct {
	config VietnamVNPayConfig
	client *http.Client
}

// NewVietnamVNPay creates a new VNPay payment platform
func NewVietnamVNPay(config VietnamVNPayConfig) *VietnamVNPay {
	// Set default API endpoint if not provided
	if config.APIEndpoint == "" {
		if config.TestMode {
			config.APIEndpoint = "https://sandbox.vnpayment.vn"
		} else {
			config.APIEndpoint = "https://vnpayment.vn"
		}
	}

	return &VietnamVNPay{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName returns the name of the payment platform
func (p *VietnamVNPay) GetName() string {
	return "VNPay"
}

// GetCountryCode returns the country code of the payment platform
func (p *VietnamVNPay) GetCountryCode() string {
	return "VN"
}

// GetSupportedPaymentMethods returns the supported payment methods
func (p *VietnamVNPay) GetSupportedPaymentMethods() []PaymentMethod {
	return []PaymentMethod{MethodCreditCard, MethodBankTransfer, MethodQRCode}
}

// GetSupportedCurrencies returns the supported currencies
func (p *VietnamVNPay) GetSupportedCurrencies() []string {
	return []string{"VND"}
}

// CreatePayment creates a payment
func (p *VietnamVNPay) CreatePayment(request *PaymentRequest) (*PaymentResponse, error) {
	// Validate request
	if request.Currency != "VND" {
		return nil, errors.New("currency must be VND for VNPay payments")
	}

	if request.PaymentMethod != MethodCreditCard && request.PaymentMethod != MethodBankTransfer && request.PaymentMethod != MethodQRCode {
		return nil, ErrUnsupportedPaymentMethod
	}

	// Prepare VNPay request
	vnpParams := url.Values{}
	vnpParams.Add("vnp_Version", "2.1.0")
	vnpParams.Add("vnp_Command", "pay")
	vnpParams.Add("vnp_TmnCode", p.config.MerchantID)
	vnpParams.Add("vnp_Amount", fmt.Sprintf("%d", int64(request.Amount*100))) // Convert to smallest currency unit
	vnpParams.Add("vnp_CurrCode", "VND")
	vnpParams.Add("vnp_TxnRef", request.OrderID)
	vnpParams.Add("vnp_OrderInfo", request.Description)
	vnpParams.Add("vnp_OrderType", "other")
	vnpParams.Add("vnp_Locale", "vn")
	vnpParams.Add("vnp_ReturnUrl", p.config.RedirectURL)
	vnpParams.Add("vnp_IpAddr", "127.0.0.1") // Should be replaced with actual IP in production
	
	// Add create date in VNPay format
	vnpParams.Add("vnp_CreateDate", time.Now().Format("20060102150405"))
	
	// Set payment method
	if request.PaymentMethod == MethodCreditCard {
		vnpParams.Add("vnp_BankCode", "INTCARD")
	} else if request.PaymentMethod == MethodBankTransfer {
		vnpParams.Add("vnp_BankCode", "VNBANK")
	} else if request.PaymentMethod == MethodQRCode {
		vnpParams.Add("vnp_BankCode", "VNPAYQR")
	}
	
	// Sort parameters by key
	var sortedKeys []string
	for k := range vnpParams {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	
	// Build query string
	var queryBuilder strings.Builder
	for _, k := range sortedKeys {
		queryBuilder.WriteString(k)
		queryBuilder.WriteString("=")
		queryBuilder.WriteString(vnpParams.Get(k))
		queryBuilder.WriteString("&")
	}
	
	// Remove trailing &
	queryString := queryBuilder.String()
	if len(queryString) > 0 {
		queryString = queryString[:len(queryString)-1]
	}
	
	// Generate secure hash
	h := hmac.New(sha256.New, []byte(p.config.SecureHash))
	h.Write([]byte(queryString))
	secureHash := hex.EncodeToString(h.Sum(nil))
	
	// Add secure hash to query string
	vnpParams.Add("vnp_SecureHash", secureHash)
	
	// Build payment URL
	paymentURL := p.config.APIEndpoint + "/vpcpay.html?" + vnpParams.Encode()

	// Create response
	response := &PaymentResponse{
		PaymentID:     request.OrderID,
		Status:        StatusPending,
		Amount:        request.Amount,
		Currency:      request.Currency,
		PaymentMethod: request.PaymentMethod,
		PaymentURL:    paymentURL,
		RedirectURL:   paymentURL,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(15 * time.Minute),
		Metadata:      make(map[string]string),
	}

	return response, nil
}

// GetPaymentStatus gets the status of a payment
func (p *VietnamVNPay) GetPaymentStatus(request *PaymentStatusRequest) (*PaymentResponse, error) {
	// Prepare status request
	vnpParams := url.Values{}
	vnpParams.Add("vnp_Version", "2.1.0")
	vnpParams.Add("vnp_Command", "querydr")
	vnpParams.Add("vnp_TmnCode", p.config.MerchantID)
	vnpParams.Add("vnp_TxnRef", request.PaymentID)
	vnpParams.Add("vnp_OrderInfo", "Query transaction status")
	vnpParams.Add("vnp_TransDate", time.Now().Format("20060102150405"))
	vnpParams.Add("vnp_CreateDate", time.Now().Format("20060102150405"))
	vnpParams.Add("vnp_IpAddr", "127.0.0.1") // Should be replaced with actual IP in production
	
	// Sort parameters by key
	var sortedKeys []string
	for k := range vnpParams {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	
	// Build query string
	var queryBuilder strings.Builder
	for _, k := range sortedKeys {
		queryBuilder.WriteString(k)
		queryBuilder.WriteString("=")
		queryBuilder.WriteString(vnpParams.Get(k))
		queryBuilder.WriteString("&")
	}
	
	// Remove trailing &
	queryString := queryBuilder.String()
	if len(queryString) > 0 {
		queryString = queryString[:len(queryString)-1]
	}
	
	// Generate secure hash
	h := hmac.New(sha256.New, []byte(p.config.SecureHash))
	h.Write([]byte(queryString))
	secureHash := hex.EncodeToString(h.Sum(nil))
	
	// Add secure hash to query string
	vnpParams.Add("vnp_SecureHash", secureHash)
	
	// Create HTTP request
	req, err := http.NewRequest("GET", p.config.APIEndpoint+"/merchant_webapi/api/transaction?"+vnpParams.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

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
	if responseCode, ok := statusResponse["vnp_ResponseCode"].(string); ok && responseCode != "00" {
		errorMsg := "unknown error"
		if msg, ok := statusResponse["vnp_Message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("VNPay error: %s", errorMsg)
	}

	// Extract payment details
	transactionStatus, _ := statusResponse["vnp_TransactionStatus"].(string)
	amountStr, _ := statusResponse["vnp_Amount"].(string)
	amount, _ := strconv.ParseFloat(amountStr, 64)
	amount = amount / 100 // Convert from smallest currency unit
	bankCode, _ := statusResponse["vnp_BankCode"].(string)
	transactionDate, _ := statusResponse["vnp_PayDate"].(string)
	
	// Parse transaction date
	var createdAt time.Time
	if transactionDate != "" {
		createdAt, _ = time.Parse("20060102150405", transactionDate)
	} else {
		createdAt = time.Now()
	}

	// Map VNPay status to our status
	status := StatusPending
	var completedAt time.Time

	if transactionStatus == "00" {
		status = StatusCompleted
		completedAt = createdAt
	} else if transactionStatus == "01" || transactionStatus == "02" {
		status = StatusPending
	} else {
		status = StatusFailed
	}

	// Map bank code to payment method
	var method PaymentMethod
	if bankCode == "INTCARD" {
		method = MethodCreditCard
	} else if bankCode == "VNBANK" {
		method = MethodBankTransfer
	} else if bankCode == "VNPAYQR" {
		method = MethodQRCode
	} else {
		method = request.PaymentMethod
	}

	// Create response
	response := &PaymentResponse{
		PaymentID:     request.PaymentID,
		Status:        status,
		Amount:        amount,
		Currency:      "VND",
		PaymentMethod: method,
		TransactionID: request.PaymentID,
		CreatedAt:     createdAt,
		UpdatedAt:     time.Now(),
		CompletedAt:   completedAt,
		Metadata:      make(map[string]string),
	}

	return response, nil
}

// RefundPayment refunds a payment
func (p *VietnamVNPay) RefundPayment(request *RefundRequest) (*RefundResponse, error) {
	// Prepare refund request
	vnpParams := url.Values{}
	vnpParams.Add("vnp_Version", "2.1.0")
	vnpParams.Add("vnp_Command", "refund")
	vnpParams.Add("vnp_TmnCode", p.config.MerchantID)
	vnpParams.Add("vnp_Amount", fmt.Sprintf("%d", int64(request.Amount*100))) // Convert to smallest currency unit
	vnpParams.Add("vnp_TxnRef", request.PaymentID)
	vnpParams.Add("vnp_OrderInfo", request.Reason)
	vnpParams.Add("vnp_TransDate", time.Now().Format("20060102150405"))
	vnpParams.Add("vnp_CreateDate", time.Now().Format("20060102150405"))
	vnpParams.Add("vnp_IpAddr", "127.0.0.1") // Should be replaced with actual IP in production
	vnpParams.Add("vnp_TransactionType", "02") // 02 for refund
	
	// Sort parameters by key
	var sortedKeys []string
	for k := range vnpParams {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)
	
	// Build query string
	var queryBuilder strings.Builder
	for _, k := range sortedKeys {
		queryBuilder.WriteString(k)
		queryBuilder.WriteString("=")
		queryBuilder.WriteString(vnpParams.Get(k))
		queryBuilder.WriteString("&")
	}
	
	// Remove trailing &
	queryString := queryBuilder.String()
	if len(queryString) > 0 {
		queryString = queryString[:len(queryString)-1]
	}
	
	// Generate secure hash
	h := hmac.New(sha256.New, []byte(p.config.SecureHash))
	h.Write([]byte(queryString))
	secureHash := hex.EncodeToString(h.Sum(nil))
	
	// Add secure hash to query string
	vnpParams.Add("vnp_SecureHash", secureHash)
	
	// Create HTTP request
	req, err := http.NewRequest("GET", p.config.APIEndpoint+"/merchant_webapi/api/transaction?"+vnpParams.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

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
	if responseCode, ok := refundResponse["vnp_ResponseCode"].(string); ok && responseCode != "00" {
		errorMsg := "unknown error"
		if msg, ok := refundResponse["vnp_Message"].(string); ok {
			errorMsg = msg
		}
		return nil, fmt.Errorf("VNPay refund error: %s", errorMsg)
	}

	// Create response
	response := &RefundResponse{
		RefundID:     request.RefundID,
		PaymentID:    request.PaymentID,
		Amount:       request.Amount,
		Currency:     "VND",
		Status:       "success",
		CreatedAt:    time.Now(),
	}

	return response, nil
}
