// Database models for Asian Cryptocurrency Payment System
// Defines all database entities and their relationships

package database

import (
	"time"

	"gorm.io/gorm"
)

// Transaction represents a cryptocurrency payment transaction
type Transaction struct {
	gorm.Model
	ID                 string    `gorm:"primaryKey;type:uuid"`
	Amount             float64   `gorm:"not null"`
	Currency           string    `gorm:"size:10;not null"`
	CryptoCurrency     string    `gorm:"size:10;not null"`
	CryptoAmount       float64   `gorm:"not null"`
	SourceAddress      string    `gorm:"size:100"`
	DestinationAddress string    `gorm:"size:100;not null"`
	Status             string    `gorm:"size:20;not null"`
	CreatedAt          time.Time `gorm:"not null"`
	UpdatedAt          time.Time `gorm:"not null"`
	CompletedAt        time.Time
	ExchangeRate       float64   `gorm:"not null"`
	Fee                float64   `gorm:"not null"`
	MerchantID         string    `gorm:"size:50;not null"`
	CustomerID         string    `gorm:"size:50"`
	CountryCode        string    `gorm:"size:2;not null"`
	PaymentMethod      string    `gorm:"size:20;not null"`
	BlockchainTxID     string    `gorm:"size:100"`
	Confirmations      int       `gorm:"default:0"`
	CallbackURL        string    `gorm:"size:255"`
	SuccessURL         string    `gorm:"size:255"`
	CancelURL          string    `gorm:"size:255"`
	IPAddress          string    `gorm:"size:45"`
	UserAgent          string    `gorm:"size:255"`
	Metadata           string    `gorm:"type:jsonb"`
}

// Wallet represents a cryptocurrency wallet
type Wallet struct {
	gorm.Model
	ID           string    `gorm:"primaryKey;type:uuid"`
	Currency     string    `gorm:"size:10;not null"`
	Address      string    `gorm:"size:100;not null;uniqueIndex"`
	Balance      float64   `gorm:"not null"`
	Type         string    `gorm:"size:10;not null"` // hot or cold
	MerchantID   string    `gorm:"size:50"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
	LastSyncedAt time.Time `gorm:"not null"`
	PublicKey    string    `gorm:"size:255"`
	PrivateKey   string    `gorm:"size:255"` // Encrypted
	IsActive     bool      `gorm:"default:true"`
	Labels       string    `gorm:"type:jsonb"`
}

// Merchant represents a business using the payment system
type Merchant struct {
	gorm.Model
	ID               string    `gorm:"primaryKey;type:uuid"`
	Name             string    `gorm:"size:100;not null"`
	Email            string    `gorm:"size:100;not null;uniqueIndex"`
	Phone            string    `gorm:"size:20"`
	CountryCode      string    `gorm:"size:2;not null"`
	APIKey           string    `gorm:"size:64;not null;uniqueIndex"`
	APISecret        string    `gorm:"size:128;not null"` // Encrypted
	WebhookURL       string    `gorm:"size:255"`
	WebhookSecret    string    `gorm:"size:64"`
	CallbackURL      string    `gorm:"size:255"`
	SuccessURL       string    `gorm:"size:255"`
	CancelURL        string    `gorm:"size:255"`
	CreatedAt        time.Time `gorm:"not null"`
	UpdatedAt        time.Time `gorm:"not null"`
	IsActive         bool      `gorm:"default:true"`
	VerificationStatus string  `gorm:"size:20;default:'pending'"`
	FeePercentage    float64   `gorm:"default:1.0"`
	SettlementCurrency string  `gorm:"size:10;default:'USD'"`
	SettlementAddress string   `gorm:"size:100"`
	Settings         string    `gorm:"type:jsonb"`
}

// Customer represents a customer making payments
type Customer struct {
	gorm.Model
	ID          string    `gorm:"primaryKey;type:uuid"`
	Email       string    `gorm:"size:100;uniqueIndex"`
	Name        string    `gorm:"size:100"`
	Phone       string    `gorm:"size:20"`
	CountryCode string    `gorm:"size:2"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
	MerchantID  string    `gorm:"size:50;not null"`
	IPAddress   string    `gorm:"size:45"`
	UserAgent   string    `gorm:"size:255"`
	Metadata    string    `gorm:"type:jsonb"`
}

// ExchangeRate represents cryptocurrency exchange rates
type ExchangeRate struct {
	gorm.Model
	ID             uint      `gorm:"primaryKey;autoIncrement"`
	BaseCurrency   string    `gorm:"size:10;not null"`
	TargetCurrency string    `gorm:"size:10;not null"`
	Rate           float64   `gorm:"not null"`
	Source         string    `gorm:"size:50;not null"` // Exchange name
	Timestamp      time.Time `gorm:"not null;index"`
}

// AuditLog represents system audit logs (stored in MongoDB)
type AuditLog struct {
	ID        string    `bson:"_id,omitempty"`
	Action    string    `bson:"action"`
	EntityType string   `bson:"entity_type"`
	EntityID  string    `bson:"entity_id"`
	UserID    string    `bson:"user_id,omitempty"`
	IPAddress string    `bson:"ip_address,omitempty"`
	Timestamp time.Time `bson:"timestamp"`
	Details   map[string]interface{} `bson:"details,omitempty"`
}

// SystemLog represents system logs (stored in MongoDB)
type SystemLog struct {
	ID        string    `bson:"_id,omitempty"`
	Level     string    `bson:"level"` // info, warning, error, critical
	Message   string    `bson:"message"`
	Component string    `bson:"component"`
	Timestamp time.Time `bson:"timestamp"`
	Details   map[string]interface{} `bson:"details,omitempty"`
}

// TransactionRepository handles database operations for transactions
type TransactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Create creates a new transaction
func (r *TransactionRepository) Create(tx *Transaction) error {
	return r.db.Create(tx).Error
}

// FindByID finds a transaction by ID
func (r *TransactionRepository) FindByID(id string) (*Transaction, error) {
	var tx Transaction
	err := r.db.Where("id = ?", id).First(&tx).Error
	return &tx, err
}

// Update updates a transaction
func (r *TransactionRepository) Update(tx *Transaction) error {
	return r.db.Save(tx).Error
}

// UpdateStatus updates a transaction's status
func (r *TransactionRepository) UpdateStatus(id string, status string) error {
	return r.db.Model(&Transaction{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// FindByMerchantID finds transactions by merchant ID
func (r *TransactionRepository) FindByMerchantID(merchantID string, limit, offset int) ([]Transaction, error) {
	var transactions []Transaction
	err := r.db.Where("merchant_id = ?", merchantID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

// WalletRepository handles database operations for wallets
type WalletRepository struct {
	db *gorm.DB
}

// NewWalletRepository creates a new wallet repository
func NewWalletRepository(db *gorm.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

// Create creates a new wallet
func (r *WalletRepository) Create(wallet *Wallet) error {
	return r.db.Create(wallet).Error
}

// FindByID finds a wallet by ID
func (r *WalletRepository) FindByID(id string) (*Wallet, error) {
	var wallet Wallet
	err := r.db.Where("id = ?", id).First(&wallet).Error
	return &wallet, err
}

// FindByAddress finds a wallet by address
func (r *WalletRepository) FindByAddress(address string) (*Wallet, error) {
	var wallet Wallet
	err := r.db.Where("address = ?", address).First(&wallet).Error
	return &wallet, err
}

// Update updates a wallet
func (r *WalletRepository) Update(wallet *Wallet) error {
	return r.db.Save(wallet).Error
}

// UpdateBalance updates a wallet's balance
func (r *WalletRepository) UpdateBalance(id string, balance float64) error {
	return r.db.Model(&Wallet{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"balance":        balance,
			"updated_at":     time.Now(),
			"last_synced_at": time.Now(),
		}).Error
}

// FindByMerchantID finds wallets by merchant ID
func (r *WalletRepository) FindByMerchantID(merchantID string) ([]Wallet, error) {
	var wallets []Wallet
	err := r.db.Where("merchant_id = ?", merchantID).Find(&wallets).Error
	return wallets, err
}
