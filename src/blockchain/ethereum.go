// Ethereum blockchain implementation for Asian Cryptocurrency Payment System

package blockchain

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
)

// EthereumClient implements the BlockchainClient interface for Ethereum
type EthereumClient struct {
	client      *ethclient.Client
	explorerURL string
	testMode    bool
	chainID     *big.Int
}

// EthereumConfig holds configuration for Ethereum client
type EthereumConfig struct {
	NodeURL     string
	ExplorerURL string
	TestMode    bool
}

// NewEthereumClient creates a new Ethereum client
func NewEthereumClient(config EthereumConfig) (*EthereumClient, error) {
	// Connect to Ethereum node
	client, err := ethclient.Dial(config.NodeURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	// Get chain ID
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	// Set default explorer URL if not provided
	explorerURL := config.ExplorerURL
	if explorerURL == "" {
		if config.TestMode {
			explorerURL = "https://sepolia.etherscan.io/tx/"
		} else {
			explorerURL = "https://etherscan.io/tx/"
		}
	}

	return &EthereumClient{
		client:      client,
		explorerURL: explorerURL,
		testMode:    config.TestMode,
		chainID:     chainID,
	}, nil
}

// GetName returns the name of the blockchain
func (c *EthereumClient) GetName() string {
	return "Ethereum"
}

// GetCurrency returns the currency code of the blockchain
func (c *EthereumClient) GetCurrency() string {
	return "ETH"
}

// GenerateAddress generates a new Ethereum address
func (c *EthereumClient) GenerateAddress() (string, error) {
	// Generate a new private key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate private key: %w", err)
	}

	// Get public key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", errors.New("failed to cast public key to ECDSA")
	}

	// Generate address
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Store private key securely (in a real implementation)
	// For now, we'll just log it
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := hexutil.Encode(privateKeyBytes)
	log.Printf("Generated new Ethereum address: %s with private key: %s", address.Hex(), privateKeyHex)

	return address.Hex(), nil
}

// ValidateAddress validates if an address is valid for Ethereum
func (c *EthereumClient) ValidateAddress(address string) bool {
	return common.IsHexAddress(address)
}

// GetBalance returns the balance of an Ethereum address
func (c *EthereumClient) GetBalance(address string) (float64, error) {
	// Validate address
	if !c.ValidateAddress(address) {
		return 0, ErrInvalidAddress
	}

	// Get balance
	account := common.HexToAddress(address)
	balance, err := c.client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}

	// Convert from wei to ether
	weiPerEther := big.NewFloat(params.Ether)
	balanceFloat := new(big.Float).SetInt(balance)
	balanceFloat.Quo(balanceFloat, weiPerEther)

	// Convert to float64
	balanceFloat64, _ := balanceFloat.Float64()
	return balanceFloat64, nil
}

// GetTransaction returns transaction details by transaction ID
func (c *EthereumClient) GetTransaction(txID string) (*Transaction, error) {
	// Parse transaction ID
	if !strings.HasPrefix(txID, "0x") {
		txID = "0x" + txID
	}
	hash := common.HexToHash(txID)

	// Get transaction
	tx, isPending, err := c.client.TransactionByHash(context.Background(), hash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	// Get transaction receipt if not pending
	var receipt *types.Receipt
	var blockNumber uint64
	var blockHash string
	var confirmations uint64
	var status TransactionStatus
	var timestamp time.Time

	if !isPending {
		receipt, err = c.client.TransactionReceipt(context.Background(), hash)
		if err != nil {
			return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
		}

		// Get block details
		block, err := c.client.BlockByHash(context.Background(), receipt.BlockHash)
		if err != nil {
			return nil, fmt.Errorf("failed to get block: %w", err)
		}

		blockNumber = block.NumberU64()
		blockHash = block.Hash().Hex()
		timestamp = time.Unix(int64(block.Time()), 0)

		// Get current block number for confirmations
		currentBlock, err := c.client.BlockNumber(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to get current block number: %w", err)
		}

		confirmations = currentBlock - blockNumber
		if receipt.Status == 1 {
			status = StatusConfirmed
		} else {
			status = StatusFailed
		}
	} else {
		status = StatusPending
		timestamp = time.Now()
	}

	// Get sender address
	msg, err := tx.AsMessage(types.NewEIP155Signer(c.chainID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender: %w", err)
	}
	from := msg.From().Hex()

	// Get recipient address
	var to string
	if tx.To() != nil {
		to = tx.To().Hex()
	} else {
		to = "contract creation"
	}

	// Get value
	value := tx.Value()
	valueFloat := new(big.Float).SetInt(value)
	valueFloat.Quo(valueFloat, big.NewFloat(params.Ether))
	amount, _ := valueFloat.Float64()

	// Get gas price and gas used for fee calculation
	gasPrice := tx.GasPrice()
	var gasUsed uint64
	if receipt != nil {
		gasUsed = receipt.GasUsed
	} else {
		gasUsed = tx.Gas()
	}

	// Calculate fee
	fee := new(big.Float).SetInt(new(big.Int).Mul(gasPrice, big.NewInt(int64(gasUsed))))
	fee.Quo(fee, big.NewFloat(params.Ether))
	feeFloat, _ := fee.Float64()

	// Create transaction object
	transaction := &Transaction{
		TxID:          txID,
		BlockHash:     blockHash,
		BlockNumber:   blockNumber,
		From:          from,
		To:            to,
		Amount:        amount,
		Fee:           feeFloat,
		Confirmations: confirmations,
		Status:        status,
		Timestamp:     timestamp,
		Currency:      "ETH",
		ExplorerURL:   c.GetExplorerURL(txID),
	}

	return transaction, nil
}

// SendTransaction sends an Ethereum transaction
func (c *EthereumClient) SendTransaction(fromAddress, toAddress string, amount float64, privateKeyHex string) (string, error) {
	// Validate addresses
	if !c.ValidateAddress(fromAddress) || !c.ValidateAddress(toAddress) {
		return "", ErrInvalidAddress
	}

	// Parse private key
	if !strings.HasPrefix(privateKeyHex, "0x") {
		privateKeyHex = "0x" + privateKeyHex
	}
	privateKeyBytes, err := hexutil.Decode(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}
	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}

	// Get public key and verify address
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", errors.New("failed to cast public key to ECDSA")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	if address.Hex() != fromAddress {
		return "", errors.New("private key does not match from address")
	}

	// Get nonce
	nonce, err := c.client.PendingNonceAt(context.Background(), address)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get gas price
	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get gas price: %w", err)
	}

	// Convert amount from ether to wei
	value := new(big.Float).Mul(big.NewFloat(amount), big.NewFloat(params.Ether))
	valueInt, _ := value.Int(nil)

	// Create transaction
	toAddress = common.HexToAddress(toAddress)
	gasLimit := uint64(21000) // Standard gas limit for ETH transfer
	tx := types.NewTransaction(nonce, common.HexToAddress(toAddress.Hex()), valueInt, gasLimit, gasPrice, nil)

	// Sign transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(c.chainID), privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	err = c.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTx.Hash().Hex(), nil
}

// EstimateFee estimates the fee for an Ethereum transaction
func (c *EthereumClient) EstimateFee(fromAddress, toAddress string, amount float64) (float64, error) {
	// Validate addresses
	if !c.ValidateAddress(fromAddress) || !c.ValidateAddress(toAddress) {
		return 0, ErrInvalidAddress
	}

	// Get gas price
	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Standard gas limit for ETH transfer
	gasLimit := uint64(21000)

	// Calculate fee
	fee := new(big.Float).SetInt(new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit))))
	fee.Quo(fee, big.NewFloat(params.Ether))
	feeFloat, _ := fee.Float64()

	return feeFloat, nil
}

// GetTransactionsByAddress returns transactions for a specific address
func (c *EthereumClient) GetTransactionsByAddress(address string, limit int) ([]Transaction, error) {
	// Validate address
	if !c.ValidateAddress(address) {
		return nil, ErrInvalidAddress
	}

	// In a real implementation, we would use a blockchain explorer API or indexer
	// For now, we'll return an empty slice
	return []Transaction{}, nil
}

// GetConfirmations returns the number of confirmations for a transaction
func (c *EthereumClient) GetConfirmations(txID string) (uint64, error) {
	// Parse transaction ID
	if !strings.HasPrefix(txID, "0x") {
		txID = "0x" + txID
	}
	hash := common.HexToHash(txID)

	// Get transaction receipt
	receipt, err := c.client.TransactionReceipt(context.Background(), hash)
	if err != nil {
		if errors.Is(err, ethereum.NotFound) {
			return 0, nil // Transaction not yet mined
		}
		return 0, fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	// Get current block number
	currentBlock, err := c.client.BlockNumber(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to get current block number: %w", err)
	}

	// Calculate confirmations
	return currentBlock - receipt.BlockNumber.Uint64(), nil
}

// GetExplorerURL returns the URL to view the transaction in a block explorer
func (c *EthereumClient) GetExplorerURL(txID string) string {
	if !strings.HasPrefix(txID, "0x") {
		txID = "0x" + txID
	}
	return c.explorerURL + txID
}

// Close closes the Ethereum client connection
func (c *EthereumClient) Close() {
	c.client.Close()
}
