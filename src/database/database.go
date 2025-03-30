// Database package for Asian Cryptocurrency Payment System
// Provides database connectivity and operations for both PostgreSQL and MongoDB

package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DBConfig holds database configuration
type DBConfig struct {
	PostgresHost     string
	PostgresPort     int
	PostgresUser     string
	PostgresPassword string
	PostgresDBName   string
	PostgresSSLMode  string
	MongoHost        string
	MongoPort        int
	MongoUser        string
	MongoPassword    string
	MongoDBName      string
}

// DBManager manages database connections and operations
type DBManager struct {
	PostgresDB *gorm.DB
	MongoDB    *mongo.Database
	Config     *DBConfig
}

// NewDBManager creates a new database manager with the provided configuration
func NewDBManager(config *DBConfig) *DBManager {
	return &DBManager{
		Config: config,
	}
}

// ConnectPostgres establishes connection to PostgreSQL database
func (m *DBManager) ConnectPostgres() error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		m.Config.PostgresHost,
		m.Config.PostgresPort,
		m.Config.PostgresUser,
		m.Config.PostgresPassword,
		m.Config.PostgresDBName,
		m.Config.PostgresSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	m.PostgresDB = db
	log.Println("Connected to PostgreSQL database")
	return nil
}

// ConnectMongo establishes connection to MongoDB database
func (m *DBManager) ConnectMongo() error {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d",
		m.Config.MongoUser,
		m.Config.MongoPassword,
		m.Config.MongoHost,
		m.Config.MongoPort,
	)

	clientOptions := options.Client().ApplyURI(uri)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the MongoDB server to verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	m.MongoDB = client.Database(m.Config.MongoDBName)
	log.Println("Connected to MongoDB database")
	return nil
}

// Close closes all database connections
func (m *DBManager) Close() error {
	// Close PostgreSQL connection
	if m.PostgresDB != nil {
		sqlDB, err := m.PostgresDB.DB()
		if err != nil {
			return fmt.Errorf("failed to get underlying SQL DB: %w", err)
		}
		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("failed to close PostgreSQL connection: %w", err)
		}
	}

	// Close MongoDB connection
	if m.MongoDB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := m.MongoDB.Client().Disconnect(ctx); err != nil {
			return fmt.Errorf("failed to close MongoDB connection: %w", err)
		}
	}

	log.Println("Database connections closed")
	return nil
}

// MigrateSchema migrates the database schema for PostgreSQL
func (m *DBManager) MigrateSchema() error {
	if m.PostgresDB == nil {
		return fmt.Errorf("PostgreSQL connection not established")
	}

	// Auto migrate all models
	// Add all models that need to be migrated here
	err := m.PostgresDB.AutoMigrate(
		&Transaction{},
		&Wallet{},
		&Merchant{},
		&Customer{},
		&ExchangeRate{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate schema: %w", err)
	}

	log.Println("Database schema migrated successfully")
	return nil
}

// DefaultConfig returns a default database configuration for development
func DefaultConfig() *DBConfig {
	return &DBConfig{
		PostgresHost:     "localhost",
		PostgresPort:     5432,
		PostgresUser:     "postgres",
		PostgresPassword: "postgres",
		PostgresDBName:   "asian_crypto_payment",
		PostgresSSLMode:  "disable",
		MongoHost:        "localhost",
		MongoPort:        27017,
		MongoUser:        "mongodb",
		MongoPassword:    "mongodb",
		MongoDBName:      "asian_crypto_payment_logs",
	}
}
