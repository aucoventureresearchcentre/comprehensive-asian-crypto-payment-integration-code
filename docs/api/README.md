# API Reference

This document provides detailed information about the API endpoints available in the Asian Cryptocurrency Payment System.

## Base URL

All API endpoints are relative to the base URL of your server:

```
https://your-domain.com/api/v1
```

## Authentication

Most API endpoints require authentication. Include your API key in the request header:

```
Authorization: Bearer YOUR_API_KEY
```

You can generate API keys in the admin dashboard.

## Error Handling

The API uses standard HTTP status codes to indicate the success or failure of a request. In case of an error, the response body will contain additional information:

```json
{
  "error": true,
  "code": "ERROR_CODE",
  "message": "A human-readable error message",
  "details": {}  // Optional additional details
}
```

## Rate Limiting

API requests are rate-limited to 100 requests per minute per API key. If you exceed this limit, you'll receive a 429 Too Many Requests response.

## Endpoints

### Payments

#### Create Payment

```
POST /payments
```

Creates a new payment request.

**Request Body:**

```json
{
  "amount": 100.50,
  "currency": "MYR",
  "payment_method": "ewallet",
  "payment_platform": "malaysia_grabpay",
  "order_id": "ORDER123456",
  "description": "Payment for Order #123456",
  "customer_name": "John Doe",
  "customer_email": "john.doe@example.com",
  "customer_phone": "+60123456789",
  "metadata": {
    "product_id": "PROD123",
    "custom_field": "custom_value"
  }
}
```

**Response:**

```json
{
  "payment_id": "PAY123456789",
  "status": "pending",
  "amount": 100.50,
  "currency": "MYR",
  "payment_method": "ewallet",
  "payment_url": "https://payment-platform.com/pay/123456",
  "qr_code_url": "https://payment-platform.com/qr/123456",
  "redirect_url": "https://payment-platform.com/pay/123456",
  "created_at": "2025-03-30T10:15:30Z",
  "expires_at": "2025-03-30T10:30:30Z",
  "metadata": {
    "product_id": "PROD123",
    "custom_field": "custom_value"
  }
}
```

#### Get Payment Status

```
GET /payments/{payment_id}
```

Retrieves the status of a payment.

**Response:**

```json
{
  "payment_id": "PAY123456789",
  "status": "completed",
  "amount": 100.50,
  "currency": "MYR",
  "payment_method": "ewallet",
  "transaction_id": "TRANS123456789",
  "created_at": "2025-03-30T10:15:30Z",
  "updated_at": "2025-03-30T10:20:45Z",
  "completed_at": "2025-03-30T10:20:45Z",
  "metadata": {
    "product_id": "PROD123",
    "custom_field": "custom_value"
  }
}
```

#### List Payments

```
GET /payments
```

Lists all payments with optional filtering.

**Query Parameters:**

- `status` - Filter by payment status (pending, completed, failed)
- `from_date` - Filter by creation date (ISO 8601 format)
- `to_date` - Filter by creation date (ISO 8601 format)
- `limit` - Number of results per page (default: 20, max: 100)
- `offset` - Pagination offset (default: 0)

**Response:**

```json
{
  "total": 125,
  "limit": 20,
  "offset": 0,
  "payments": [
    {
      "payment_id": "PAY123456789",
      "status": "completed",
      "amount": 100.50,
      "currency": "MYR",
      "payment_method": "ewallet",
      "created_at": "2025-03-30T10:15:30Z",
      "updated_at": "2025-03-30T10:20:45Z"
    },
    // More payments...
  ]
}
```

#### Refund Payment

```
POST /payments/{payment_id}/refund
```

Initiates a refund for a payment.

**Request Body:**

```json
{
  "amount": 100.50,  // Optional, defaults to full amount
  "reason": "Customer requested refund",
  "refund_id": "REF123456"  // Optional, system will generate if not provided
}
```

**Response:**

```json
{
  "refund_id": "REF123456",
  "payment_id": "PAY123456789",
  "amount": 100.50,
  "currency": "MYR",
  "status": "pending",
  "transaction_id": "REFTRANS123456",
  "created_at": "2025-03-30T11:15:30Z"
}
```

### Wallets

#### Get Wallet Balance

```
GET /wallets/{currency}
```

Retrieves the balance of a specific cryptocurrency wallet.

**Response:**

```json
{
  "currency": "BTC",
  "balance": 0.12345678,
  "available_balance": 0.12345678,
  "pending_balance": 0.0,
  "address": "bc1q9h6tlfv5mzpkruk32df8v5d9wzu9v42jyjpnkm",
  "updated_at": "2025-03-30T10:15:30Z"
}
```

#### List Wallet Transactions

```
GET /wallets/{currency}/transactions
```

Lists transactions for a specific cryptocurrency wallet.

**Query Parameters:**

- `type` - Filter by transaction type (deposit, withdrawal, transfer)
- `from_date` - Filter by date (ISO 8601 format)
- `to_date` - Filter by date (ISO 8601 format)
- `limit` - Number of results per page (default: 20, max: 100)
- `offset` - Pagination offset (default: 0)

**Response:**

```json
{
  "total": 45,
  "limit": 20,
  "offset": 0,
  "transactions": [
    {
      "transaction_id": "WTRANS123456",
      "type": "deposit",
      "amount": 0.01234567,
      "currency": "BTC",
      "status": "confirmed",
      "blockchain_tx_id": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
      "confirmations": 6,
      "created_at": "2025-03-30T10:15:30Z",
      "updated_at": "2025-03-30T10:45:30Z"
    },
    // More transactions...
  ]
}
```

### Exchange Rates

#### Get Current Exchange Rates

```
GET /exchange-rates
```

Retrieves current exchange rates for all supported cryptocurrencies and fiat currencies.

**Query Parameters:**

- `base` - Base currency (default: USD)
- `currencies` - Comma-separated list of currencies to include (optional)

**Response:**

```json
{
  "base": "USD",
  "timestamp": "2025-03-30T10:15:30Z",
  "rates": {
    "BTC": 0.000016,
    "ETH": 0.00031,
    "MYR": 4.15,
    "SGD": 1.35,
    "IDR": 14500,
    "THB": 32.5,
    "VND": 23000,
    "KHR": 4100,
    "LAK": 17000
  }
}
```

#### Get Historical Exchange Rates

```
GET /exchange-rates/historical
```

Retrieves historical exchange rates.

**Query Parameters:**

- `date` - Date in ISO 8601 format (required)
- `base` - Base currency (default: USD)
- `currencies` - Comma-separated list of currencies to include (optional)

**Response:**

```json
{
  "base": "USD",
  "date": "2025-03-25",
  "rates": {
    "BTC": 0.000015,
    "ETH": 0.00030,
    "MYR": 4.12,
    "SGD": 1.34,
    "IDR": 14400,
    "THB": 32.2,
    "VND": 22900,
    "KHR": 4050,
    "LAK": 16900
  }
}
```

### Webhooks

#### Create Webhook

```
POST /webhooks
```

Creates a new webhook subscription.

**Request Body:**

```json
{
  "url": "https://your-domain.com/webhook-handler",
  "events": ["payment.created", "payment.completed", "payment.failed"],
  "description": "Payment notifications webhook"
}
```

**Response:**

```json
{
  "webhook_id": "WH123456",
  "url": "https://your-domain.com/webhook-handler",
  "events": ["payment.created", "payment.completed", "payment.failed"],
  "description": "Payment notifications webhook",
  "secret": "whsec_abcdefghijklmnopqrstuvwxyz",
  "created_at": "2025-03-30T10:15:30Z"
}
```

#### List Webhooks

```
GET /webhooks
```

Lists all webhook subscriptions.

**Response:**

```json
{
  "webhooks": [
    {
      "webhook_id": "WH123456",
      "url": "https://your-domain.com/webhook-handler",
      "events": ["payment.created", "payment.completed", "payment.failed"],
      "description": "Payment notifications webhook",
      "created_at": "2025-03-30T10:15:30Z"
    },
    // More webhooks...
  ]
}
```

#### Delete Webhook

```
DELETE /webhooks/{webhook_id}
```

Deletes a webhook subscription.

**Response:**

```json
{
  "success": true,
  "message": "Webhook deleted successfully"
}
```

## Webhook Events

When a webhook event occurs, we'll send a POST request to your webhook URL with the following payload:

```json
{
  "id": "evt_123456",
  "type": "payment.completed",
  "created": "2025-03-30T10:20:45Z",
  "data": {
    // Event-specific data
  }
}
```

To verify webhook authenticity, we include a signature in the `X-Signature` header. The signature is a HMAC-SHA256 hash of the request body using your webhook secret as the key.

## SDK Methods

Our SDKs provide convenient wrappers around these API endpoints. Refer to the SDK documentation for your platform for more details:

- [Web SDK](../integration/web.md)
- [POS SDK](../integration/pos.md)
- [Kiosk SDK](../integration/kiosk.md)
