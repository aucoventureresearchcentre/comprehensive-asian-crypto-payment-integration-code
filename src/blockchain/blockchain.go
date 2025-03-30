// Blockchain package for Asian Cryptocurrency Payment System
// Provides interfaces and implementations for blockchain interactions

package blockchain

import (
	"errors"
	"time"
)

// Common errors
var (
	ErrInvalidAddress      = errors.New("invalid blockchain address")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrTransactionFailed   = errors.New("transaction failed")
	ErrNetworkUnavailable  = errors.New("blockchain network unavailable")
	ErrInvalidTransaction  = errors.New("invalid transaction")
)

// TransactionStatus defines the status of a blockchain transaction
type TransactionStatus string

const (
	// StatusPending indicates transaction is not yet confirmed
	StatusPending TransactionStatus = "pending"
	// StatusConfirmed indicates transaction has been confirmed
	StatusConfirmed TransactionStatus = "confirmed"
	// StatusFailed indicates transaction has failed
	StatusFailed TransactionStatus = "failed"
)

// Transaction represents a blockchain transaction
type Transaction struct {
	TxID            string           `json:"tx_id"`
	BlockHash       string           `json:"block_hash,omitempty"`
	BlockNumber     uint64           `json:"block_number,omitempty"`
	From            string           `json:"from"`
	To              string           `json:"to"`
	Amount          float64          `json:"amount"`
	Fee             float64          `json:"fee"`
	Confirmations   uint64           `json:"confirmations"`
	Status          TransactionStatus `json:"status"`
	Timestamp       time.Time        `json:"timestamp"`
	Currency        string           `json:"currency"`
	ExplorerURL     string           `json:"explorer_url,omitempty"`
	RawTransaction  string           `json:"raw_transaction,omitempty"`
}

// BlockchainClient defines the interface for blockchain interactions
type BlockchainClient interface {
	// GetName returns the name of the blockchain
	GetName() string
	
	// GetCurrency returns the currency code of the blockchain
	GetCurrency() string
	
	// GenerateAddress generates a new address for receiving payments
	GenerateAddress() (string, error)
	
	// ValidateAddress validates if an address is valid for this blockchain
	ValidateAddress(address string) bool
	
	// GetBalance returns the balance of an address
	GetBalance(address string) (float64, error)
	
	// GetTransaction returns transaction details by transaction ID
	GetTransaction(txID string) (*Transaction, error)
	
	// SendTransaction sends a transaction from one address to another
	SendTransaction(fromAddress, toAddress string, amount float64, privateKey string) (string, error)
	
	// EstimateFee estimates the fee for a transaction
	EstimateFee(fromAddress, toAddress string, amount float64) (float64, error)
	
	// GetTransactionsByAddress returns transactions for a specific address
	GetTransactionsByAddress(address string, limit int) ([]Transaction, error)
	
	// GetConfirmations returns the number of confirmations for a transaction
	GetConfirmations(txID string) (uint64, error)
	
	// GetExplorerURL returns the URL to view the transaction in a block explorer
	GetExplorerURL(txID string) string
}

// BlockchainClientFactory creates blockchain clients for different cryptocurrencies
type BlockchainClientFactory struct {
	clients map[string]BlockchainClient
}

// NewBlockchainClientFactory creates a new blockchain client factory
func NewBlockchainClientFactory() *BlockchainClientFactory {
	return &BlockchainClientFactory{
		clients: make(map[string]BlockchainClient),
	}
}

// RegisterClient registers a blockchain client for a specific cryptocurrency
func (f *BlockchainClientFactory) RegisterClient(currency string, client BlockchainClient) {
	f.clients[currency] = client
}

// GetClient returns a blockchain client for a specific cryptocurrency
func (f *BlockchainClientFactory) GetClient(currency string) (BlockchainClient, error) {
	client, exists := f.clients[currency]
	if !exists {
		return nil, errors.New("blockchain client not found for currency: " + currency)
	}
	return client, nil
}

// GetSupportedCurrencies returns a list of supported cryptocurrencies
func (f *BlockchainClientFactory) GetSupportedCurrencies() []string {
	currencies := make([]string, 0, len(f.clients))
	for currency := range f.clients {
		currencies = append(currencies, currency)
	}
	return currencies
}
