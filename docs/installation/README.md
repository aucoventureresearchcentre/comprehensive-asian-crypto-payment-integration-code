# Installation Guide

This guide provides step-by-step instructions for installing and configuring the Asian Cryptocurrency Payment System.

## System Requirements

### Hardware Requirements
- **CPU**: 2+ cores
- **RAM**: 4GB minimum, 8GB recommended
- **Storage**: 20GB minimum free space
- **Network**: Stable internet connection

### Software Requirements
- **Operating System**: Linux (Ubuntu 20.04 LTS or newer recommended), macOS 10.15+, or Windows 10/11
- **Go**: Version 1.18 or newer
- **PostgreSQL**: Version 12 or newer
- **Redis**: Version 6 or newer
- **Node.js**: Version 14 or newer (for web SDK)
- **Java**: JDK 11 or newer (for POS SDK)
- **C++ Compiler**: GCC 9+ or equivalent (for kiosk SDK)

## Installation Steps

### 1. Clone the Repository

```bash
git clone https://github.com/your-organization/asian-crypto-payment.git
cd asian-crypto-payment
```

### 2. Install Dependencies

#### Ubuntu/Debian

```bash
# Install system dependencies
sudo apt update
sudo apt install -y build-essential golang postgresql postgresql-contrib redis-server nodejs npm openjdk-11-jdk

# Install Go dependencies
go mod download

# Install Node.js dependencies for web SDK
cd sdk/web
npm install
cd ../..

# Install Java dependencies for POS SDK
cd sdk/pos
./gradlew build
cd ../..
```

#### macOS

```bash
# Install Homebrew if not already installed
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install system dependencies
brew update
brew install go postgresql redis node openjdk@11

# Start services
brew services start postgresql
brew services start redis

# Install Go dependencies
go mod download

# Install Node.js dependencies for web SDK
cd sdk/web
npm install
cd ../..

# Install Java dependencies for POS SDK
cd sdk/pos
./gradlew build
cd ../..
```

#### Windows

1. Install Go from [https://golang.org/dl/](https://golang.org/dl/)
2. Install PostgreSQL from [https://www.postgresql.org/download/windows/](https://www.postgresql.org/download/windows/)
3. Install Redis from [https://github.com/microsoftarchive/redis/releases](https://github.com/microsoftarchive/redis/releases)
4. Install Node.js from [https://nodejs.org/](https://nodejs.org/)
5. Install JDK 11 from [https://adoptopenjdk.net/](https://adoptopenjdk.net/)
6. Install Visual Studio Build Tools for C++ compilation

Then open Command Prompt or PowerShell:

```powershell
# Install Go dependencies
go mod download

# Install Node.js dependencies for web SDK
cd sdk/web
npm install
cd ../..

# Install Java dependencies for POS SDK
cd sdk/pos
gradlew.bat build
cd ../..
```

### 3. Set Up the Database

```bash
# Create PostgreSQL database and user
sudo -u postgres psql -c "CREATE DATABASE asian_crypto_payment;"
sudo -u postgres psql -c "CREATE USER payment_user WITH ENCRYPTED PASSWORD 'your_secure_password';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE asian_crypto_payment TO payment_user;"

# Run database migrations
go run cmd/migrate/main.go
```

### 4. Configure the System

Create a configuration file by copying the example:

```bash
cp config.example.yaml config.yaml
```

Edit `config.yaml` to set your specific configuration parameters:

```yaml
# Database configuration
database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  name: "asian_crypto_payment"
  user: "payment_user"
  password: "your_secure_password"
  sslmode: "disable"

# Redis configuration
redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

# Blockchain configuration
blockchain:
  bitcoin:
    network: "mainnet"  # or "testnet"
    rpc_url: "http://localhost:8332"
    rpc_user: "your_bitcoin_rpc_user"
    rpc_password: "your_bitcoin_rpc_password"
  ethereum:
    network: "mainnet"  # or "ropsten", "rinkeby", etc.
    rpc_url: "https://mainnet.infura.io/v3/your_infura_project_id"

# Exchange rate service
exchange:
  provider: "coingecko"  # or "binance", "coinmarketcap"
  api_key: "your_api_key"  # if required
  update_interval: 60  # in seconds

# Security settings
security:
  encryption_key: "your_secure_encryption_key"
  jwt_secret: "your_secure_jwt_secret"
  token_expiry: 86400  # in seconds (24 hours)

# Payment platforms
payment_platforms:
  # Malaysia
  malaysia_fpx:
    merchant_id: "your_merchant_id"
    merchant_key: "your_merchant_key"
    api_endpoint: "https://api.example.com"
    callback_url: "https://your-domain.com/callback/malaysia/fpx"
    redirect_url: "https://your-domain.com/redirect/malaysia/fpx"
    test_mode: false
  
  # Add configurations for other payment platforms as needed
```

### 5. Build the System

```bash
# Build the main server
go build -o bin/server cmd/server/main.go

# Build the web SDK
cd sdk/web
npm run build
cd ../..

# Build the POS SDK
cd sdk/pos
./gradlew jar
cd ../..

# Build the kiosk SDK
cd sdk/kiosk
make
cd ../..
```

### 6. Run the System

```bash
# Start the server
./bin/server
```

The server will start on port 8080 by default. You can access the admin dashboard at `http://localhost:8080/admin`.

## Next Steps

After installation, you should:

1. Set up SSL/TLS for secure communication
2. Configure your firewall to allow necessary traffic
3. Set up monitoring and logging
4. Integrate with your existing systems using our [Integration Guides](../integration/README.md)

## Troubleshooting

If you encounter issues during installation, please refer to the [Troubleshooting Guide](troubleshooting.md) for common problems and solutions.
