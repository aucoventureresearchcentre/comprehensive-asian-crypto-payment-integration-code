// Integration package for Asian Cryptocurrency Payment System
// Provides interfaces and implementations for payment platform integrations

package integration

import (
	"errors"
	"time"
)

// Common errors
var (
	ErrUnsupportedPaymentMethod = errors.New("unsupported payment method")
	ErrPaymentFailed            = errors.New("payment failed")
	ErrInvalidConfiguration     = errors.New("invalid configuration")
	ErrConnectionFailed         = errors.New("connection to payment platform failed")
	ErrInvalidResponse          = errors.New("invalid response from payment platform")
)

// PaymentStatus defines the status of a payment
type PaymentStatus string

const (
	// StatusPending indicates payment is pending
	StatusPending PaymentStatus = "pending"
	// StatusCompleted indicates payment is completed
	StatusCompleted PaymentStatus = "completed"
	// StatusFailed indicates payment has failed
	StatusFailed PaymentStatus = "failed"
	// StatusCancelled indicates payment was cancelled
	StatusCancelled PaymentStatus = "cancelled"
	// StatusRefunded indicates payment was refunded
	StatusRefunded PaymentStatus = "refunded"
)

// PaymentMethod defines the payment method
type PaymentMethod string

const (
	// MethodCreditCard for credit card payments
	MethodCreditCard PaymentMethod = "credit_card"
	// MethodBankTransfer for bank transfer payments
	MethodBankTransfer PaymentMethod = "bank_transfer"
	// MethodEWallet for e-wallet payments
	MethodEWallet PaymentMethod = "e_wallet"
	// MethodQRCode for QR code payments
	MethodQRCode PaymentMethod = "qr_code"
	// MethodCryptocurrency for cryptocurrency payments
	MethodCryptocurrency PaymentMethod = "cryptocurrency"
)

// PaymentRequest represents a payment request
type PaymentRequest struct {
	Amount          float64       `json:"amount"`
	Currency        string        `json:"currency"`
	Description     string        `json:"description"`
	OrderID         string        `json:"order_id"`
	CustomerID      string        `json:"customer_id,omitempty"`
	CustomerEmail   string        `json:"customer_email,omitempty"`
	CustomerName    string        `json:"customer_name,omitempty"`
	CustomerPhone   string        `json:"customer_phone,omitempty"`
	CustomerAddress string        `json:"customer_address,omitempty"`
	PaymentMethod   PaymentMethod `json:"payment_method"`
	ReturnURL       string        `json:"return_url,omitempty"`
	CallbackURL     string        `json:"callback_url,omitempty"`
	CancelURL       string        `json:"cancel_url,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	ExpiryTime      time.Time     `json:"expiry_time,omitempty"`
}

// PaymentResponse represents a payment response
type PaymentResponse struct {
	PaymentID       string        `json:"payment_id"`
	Status          PaymentStatus `json:"status"`
	Amount          float64       `json:"amount"`
	Currency        string        `json:"currency"`
	PaymentMethod   PaymentMethod `json:"payment_method"`
	TransactionID   string        `json:"transaction_id,omitempty"`
	PaymentURL      string        `json:"payment_url,omitempty"`
	QRCodeURL       string        `json:"qr_code_url,omitempty"`
	RedirectURL     string        `json:"redirect_url,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	CompletedAt     time.Time     `json:"completed_at,omitempty"`
	ExpiresAt       time.Time     `json:"expires_at,omitempty"`
	ErrorCode       string        `json:"error_code,omitempty"`
	ErrorMessage    string        `json:"error_message,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// PaymentStatusRequest represents a payment status request
type PaymentStatusRequest struct {
	PaymentID     string `json:"payment_id"`
	OrderID       string `json:"order_id,omitempty"`
	TransactionID string `json:"transaction_id,omitempty"`
}

// RefundRequest represents a refund request
type RefundRequest struct {
	PaymentID     string  `json:"payment_id"`
	Amount        float64 `json:"amount,omitempty"` // If not specified, full amount is refunded
	Reason        string  `json:"reason,omitempty"`
	RefundID      string  `json:"refund_id,omitempty"`
}

// RefundResponse represents a refund response
type RefundResponse struct {
	RefundID      string    `json:"refund_id"`
	PaymentID     string    `json:"payment_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	TransactionID string    `json:"transaction_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	CompletedAt   time.Time `json:"completed_at,omitempty"`
	ErrorCode     string    `json:"error_code,omitempty"`
	ErrorMessage  string    `json:"error_message,omitempty"`
}

// PaymentPlatform defines the interface for payment platform integrations
type PaymentPlatform interface {
	// GetName returns the name of the payment platform
	GetName() string
	
	// GetCountryCode returns the country code of the payment platform
	GetCountryCode() string
	
	// GetSupportedPaymentMethods returns the supported payment methods
	GetSupportedPaymentMethods() []PaymentMethod
	
	// GetSupportedCurrencies returns the supported currencies
	GetSupportedCurrencies() []string
	
	// CreatePayment creates a payment
	CreatePayment(request *PaymentRequest) (*PaymentResponse, error)
	
	// GetPaymentStatus gets the status of a payment
	GetPaymentStatus(request *PaymentStatusRequest) (*PaymentResponse, error)
	
	// RefundPayment refunds a payment
	RefundPayment(request *RefundRequest) (*RefundResponse, error)
}

// PaymentPlatformRegistry maintains a registry of payment platforms
type PaymentPlatformRegistry struct {
	platforms map[string]PaymentPlatform
}

// NewPaymentPlatformRegistry creates a new payment platform registry
func NewPaymentPlatformRegistry() *PaymentPlatformRegistry {
	return &PaymentPlatformRegistry{
		platforms: make(map[string]PaymentPlatform),
	}
}

// RegisterPlatform adds a payment platform to the registry
func (r *PaymentPlatformRegistry) RegisterPlatform(platform PaymentPlatform) {
	key := platform.GetCountryCode() + "_" + platform.GetName()
	r.platforms[key] = platform
}

// GetPlatform retrieves a payment platform by country code and name
func (r *PaymentPlatformRegistry) GetPlatform(countryCode, name string) (PaymentPlatform, bool) {
	key := countryCode + "_" + name
	platform, exists := r.platforms[key]
	return platform, exists
}

// GetPlatformsByCountry retrieves all payment platforms for a country
func (r *PaymentPlatformRegistry) GetPlatformsByCountry(countryCode string) []PaymentPlatform {
	var platforms []PaymentPlatform
	for key, platform := range r.platforms {
		if key[:2] == countryCode {
			platforms = append(platforms, platform)
		}
	}
	return platforms
}

// GetAllPlatforms returns all registered payment platforms
func (r *PaymentPlatformRegistry) GetAllPlatforms() []PaymentPlatform {
	platforms := make([]PaymentPlatform, 0, len(r.platforms))
	for _, platform := range r.platforms {
		platforms = append(platforms, platform)
	}
	return platforms
}
