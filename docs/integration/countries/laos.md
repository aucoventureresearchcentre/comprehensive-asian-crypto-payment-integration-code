# Laos Payment Integration

This guide provides specific details for integrating with Laotian payment platforms.

## Supported Payment Platforms

The Asian Cryptocurrency Payment System supports the following payment platforms in Laos:

1. **U-Money** - E-wallet payment system
2. **LDB (Lao Development Bank)** - Bank payment system

## Currency

The official currency for Laos is the Lao Kip (LAK).

## U-Money Integration

### Configuration

To integrate with U-Money, you need to configure the following parameters in your `config.yaml` file:

```yaml
payment_platforms:
  laos_umoney:
    merchant_id: "your_merchant_id"
    merchant_key: "your_merchant_key"
    api_endpoint: "https://api.umoney.la"
    callback_url: "https://your-domain.com/callback/laos/umoney"
    redirect_url: "https://your-domain.com/redirect/laos/umoney"
    test_mode: false
```

### Usage Example

#### Web SDK

```javascript
// Create a payment using U-Money
const payment = await paymentSDK.createPayment({
  amount: 100000.00,  // LAK 100,000
  currency: 'LAK',
  paymentMethod: 'ewallet',
  paymentPlatform: 'laos_umoney',
  orderId: 'ORDER123456',
  description: 'Payment for Order #123456',
  customerName: 'John Doe',
  customerEmail: 'john.doe@example.com',
  customerPhone: '+85620123456'
});

// Redirect to U-Money payment page or display QR code
if (payment.paymentUrl) {
  window.location.href = payment.paymentUrl;
} else if (payment.qrCodeUrl) {
  displayQRCode(payment.qrCodeUrl);
}
```

#### POS SDK

```java
// Create a payment using U-Money
PaymentRequest request = new PaymentRequest.Builder()
    .setAmount(100000.00)  // LAK 100,000
    .setCurrency("LAK")
    .setPaymentMethod(PaymentMethod.EWALLET)
    .setPaymentPlatform("laos_umoney")
    .setOrderId("ORDER123456")
    .setDescription("Payment for Order #123456")
    .setCustomerName("John Doe")
    .setCustomerEmail("john.doe@example.com")
    .setCustomerPhone("+85620123456")
    .build();

Payment payment = paymentSDK.createPayment(request);

// Display QR code or payment instructions
if (payment.getQrCodeUrl() != null) {
    displayQRCode(payment.getQrCodeUrl());
} else if (payment.getPaymentUrl() != null) {
    displayPaymentInstructions(payment.getPaymentUrl());
}
```

#### Kiosk SDK

```cpp
// Create a payment using U-Money
PaymentRequest request = {
    .amount = 100000.00,  // LAK 100,000
    .currency = "LAK",
    .payment_method = PAYMENT_METHOD_EWALLET,
    .payment_platform = "laos_umoney",
    .order_id = "ORDER123456",
    .description = "Payment for Order #123456",
    .customer_name = "John Doe",
    .customer_email = "john.doe@example.com",
    .customer_phone = "+85620123456"
};

Payment* payment = asian_crypto_payment_create_payment(payment_sdk, &request);

// Display QR code or payment instructions
if (payment->qr_code_url != NULL) {
    display_qr_code(payment->qr_code_url);
} else if (payment->payment_url != NULL) {
    display_payment_instructions(payment->payment_url);
}
```

## LDB Integration

### Configuration

To integrate with LDB, you need to configure the following parameters in your `config.yaml` file:

```yaml
payment_platforms:
  laos_ldb:
    merchant_id: "your_merchant_id"
    merchant_secret: "your_merchant_secret"
    api_endpoint: "https://api.ldb.la"
    callback_url: "https://your-domain.com/callback/laos/ldb"
    redirect_url: "https://your-domain.com/redirect/laos/ldb"
    test_mode: false
```

### Usage Example

#### Web SDK

```javascript
// Create a payment using LDB
const payment = await paymentSDK.createPayment({
  amount: 100000.00,  // LAK 100,000
  currency: 'LAK',
  paymentMethod: 'bank_transfer',  // or 'qrcode'
  paymentPlatform: 'laos_ldb',
  orderId: 'ORDER123456',
  description: 'Payment for Order #123456',
  customerName: 'John Doe',
  customerEmail: 'john.doe@example.com',
  customerPhone: '+85620123456'
});

// Redirect to LDB payment page or display QR code
if (payment.paymentUrl) {
  window.location.href = payment.paymentUrl;
} else if (payment.qrCodeUrl) {
  displayQRCode(payment.qrCodeUrl);
}
```

#### POS SDK

```java
// Create a payment using LDB
PaymentRequest request = new PaymentRequest.Builder()
    .setAmount(100000.00)  // LAK 100,000
    .setCurrency("LAK")
    .setPaymentMethod(PaymentMethod.BANK_TRANSFER)  // or QR_CODE
    .setPaymentPlatform("laos_ldb")
    .setOrderId("ORDER123456")
    .setDescription("Payment for Order #123456")
    .setCustomerName("John Doe")
    .setCustomerEmail("john.doe@example.com")
    .setCustomerPhone("+85620123456")
    .build();

Payment payment = paymentSDK.createPayment(request);

// Process payment
if (payment.getQrCodeUrl() != null) {
    displayQRCode(payment.getQrCodeUrl());
} else if (payment.getPaymentUrl() != null) {
    displayPaymentInstructions(payment.getPaymentUrl());
}
```

#### Kiosk SDK

```cpp
// Create a payment using LDB
PaymentRequest request = {
    .amount = 100000.00,  // LAK 100,000
    .currency = "LAK",
    .payment_method = PAYMENT_METHOD_BANK_TRANSFER,  // or QR_CODE
    .payment_platform = "laos_ldb",
    .order_id = "ORDER123456",
    .description = "Payment for Order #123456",
    .customer_name = "John Doe",
    .customer_email = "john.doe@example.com",
    .customer_phone = "+85620123456"
};

Payment* payment = asian_crypto_payment_create_payment(payment_sdk, &request);

// Process payment
if (payment->qr_code_url != NULL) {
    display_qr_code(payment->qr_code_url);
} else if (payment->payment_url != NULL) {
    display_payment_instructions(payment->payment_url);
}
```

## Regulatory Compliance

When operating in Laos, you must comply with the following regulations:

1. **Bank of the Lao PDR (BOL)** - The central bank of Laos regulates payment systems and financial services.
2. **Anti-Money Laundering Intelligence Office (AMLIO)** - Handles anti-money laundering and counter-terrorism financing.

Ensure that you:

1. Obtain necessary licenses from the BOL for payment services
2. Implement Know Your Customer (KYC) and Anti-Money Laundering (AML) procedures
3. Maintain proper records of all transactions
4. Report suspicious transactions to AMLIO
5. Comply with data protection regulations

## Testing

For testing in the sandbox environment, use the following test credentials:

### U-Money Test Credentials

```
Merchant ID: test_merchant_umoney
Merchant Key: test_merchant_key_umoney
```

Test User:
- Phone Number: +85620123456
- PIN: 123456

### LDB Test Credentials

```
Merchant ID: test_merchant_ldb
Merchant Secret: test_merchant_secret_ldb
```

Test Account:
- Account Number: 0123456789
- Username: testuser
- Password: testpassword

## Troubleshooting

### Common Issues with U-Money

1. **Payment Failed with Error Code 400**
   - Verify that the customer's phone number is registered with U-Money
   - Check if the amount is within the allowed limits

2. **Callback Not Received**
   - Ensure your callback URL is publicly accessible
   - Check if the callback URL is correctly configured

### Common Issues with LDB

1. **Payment Page Not Loading**
   - Verify your merchant ID and secret
   - Check if the API endpoint is accessible

2. **Transaction Declined**
   - Ensure the test account details are entered correctly
   - Check if the amount is within the allowed limits

## Support

For Laos-specific support, contact our support team at laos-support@asiancryptopayment.com.
